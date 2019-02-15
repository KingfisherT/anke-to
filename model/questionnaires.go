package model

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
)

type Questionnaires struct {
	ID           int            `json:"questionnaireID" db:"id"`
	Title        string         `json:"title"           db:"title"`
	Description  string         `json:"description"     db:"description"`
	ResTimeLimit mysql.NullTime `json:"res_time_limit"  db:"res_time_limit"`
	DeletedAt    mysql.NullTime `json:"deleted_at"      db:"deleted_at"`
	ResSharedTo  string         `json:"res_shared_to"   db:"res_shared_to"`
	CreatedAt    time.Time      `json:"created_at"      db:"created_at"`
	ModifiedAt   time.Time      `json:"modified_at"     db:"modified_at"`
}

type QuestionnairesInfo struct {
	ID           int    `json:"questionnaireID"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ResTimeLimit string `json:"res_time_limit"`
	ResSharedTo  string `json:"res_shared_to"`
	CreatedAt    string `json:"created_at"`
	ModifiedAt   string `json:"modified_at"`
	IsTargeted   bool   `json:"is_targeted"`
}

// エラーが起きれば(nil, err)
// 起こらなければ(allquestions, nil)を返す
func GetAllQuestionnaires(c echo.Context) ([]Questionnaires, error) {
	// query parametar
	sort := c.QueryParam("sort")
	page := c.QueryParam("page")

	if page == "" {
		page = "1"
	}
	num, err := strconv.Atoi(page)
	if err != nil {
		c.Logger().Error(err)
		return nil, echo.NewHTTPError(http.StatusBadRequest)
	}

	var list = map[string]string{
		"":             "",
		"created_at":   "ORDER BY created_at",
		"-created_at":  "ORDER BY created_at DESC",
		"title":        "ORDER BY title",
		"-title":       "ORDER BY title DESC",
		"modified_at":  "ORDER BY modified_at",
		"-modified_at": "ORDER BY modified_at DESC",
	}
	// アンケート一覧の配列
	allquestionnaires := []Questionnaires{}

	if err := DB.Select(&allquestionnaires,
		"SELECT * FROM questionnaires WHERE deleted_at IS NULL "+list[sort]+" lIMIT 20 OFFSET "+strconv.Itoa(20*(num-1))); err != nil {
		c.Logger().Error(err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}
	return allquestionnaires, nil
}

func GetQuestionnaires(c echo.Context, targettype TargetType) ([]QuestionnairesInfo, error) {
	allquestionnaires, err := GetAllQuestionnaires(c)
	if err != nil {
		return nil, err
	}

	userID := GetUserID(c)

	// 自分またはtraPが含まれているアンケートのID
	targetedQuestionnaireID := []int{}
	if err := DB.Select(&targetedQuestionnaireID,
		`SELECT DISTINCT questionnaire_id FROM targets WHERE user_traqid = ? OR user_traqid = 'traP'`,
		userID); err != nil {
		c.Logger().Error(err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}

	ret := []QuestionnairesInfo{}
	for _, v := range allquestionnaires {
		var targeted = false
		for _, w := range targetedQuestionnaireID {
			if w == v.ID {
				targeted = true
			}
		}
		if (targettype == TargetType(Targeted) && !targeted) || (targettype == TargetType(Nontargeted) && targeted) {
			continue
		}

		ret = append(ret,
			QuestionnairesInfo{
				ID:           v.ID,
				Title:        v.Title,
				Description:  v.Description,
				ResTimeLimit: NullTimeToString(v.ResTimeLimit),
				ResSharedTo:  v.ResSharedTo,
				CreatedAt:    v.CreatedAt.Format(time.RFC3339),
				ModifiedAt:   v.ModifiedAt.Format(time.RFC3339),
				IsTargeted:   targeted})
	}

	if len(ret) == 0 {
		return nil, echo.NewHTTPError(http.StatusNotFound)
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].ModifiedAt > ret[j].ModifiedAt
	})

	return ret, nil
}

func GetQuestionnaire(c echo.Context, questionnaireID int) (Questionnaires, error) {
	questionnaire := Questionnaires{}
	if err := DB.Get(&questionnaire, "SELECT * FROM questionnaires WHERE id = ? AND deleted_at IS NULL", questionnaireID); err != nil {
		c.Logger().Error(err)
		if err == sql.ErrNoRows {
			return Questionnaires{}, echo.NewHTTPError(http.StatusNotFound)
		} else {
			return Questionnaires{}, echo.NewHTTPError(http.StatusInternalServerError)
		}
	}
	return questionnaire, nil
}

func GetQuestionnaireInfo(c echo.Context, questionnaireID int) (Questionnaires, []string, []string, []string, error) {
	questionnaire, err := GetQuestionnaire(c, questionnaireID)
	if err != nil {
		return Questionnaires{}, nil, nil, nil, err
	}

	targets, err := GetTargets(c, questionnaireID)
	if err != nil {
		return Questionnaires{}, nil, nil, nil, err
	}

	administrators, err := GetAdministrators(c, questionnaireID)
	if err != nil {
		return Questionnaires{}, nil, nil, nil, err
	}

	respondents, err := GetRespondents(c, questionnaireID)
	if err != nil {
		return Questionnaires{}, nil, nil, nil, err
	}
	return questionnaire, targets, administrators, respondents, nil
}

func GetTitleAndLimit(c echo.Context, questionnaireID int) (string, string, error) {
	res := struct {
		Title        string         `db:"title"`
		ResTimeLimit mysql.NullTime `db:"res_time_limit"`
	}{}
	if err := DB.Get(&res,
		"SELECT title, res_time_limit FROM questionnaires WHERE id = ? AND deleted_at IS NULL",
		questionnaireID); err != nil {
		c.Logger().Error(err)
		if err == sql.ErrNoRows {
			return "", "", echo.NewHTTPError(http.StatusNotFound)
		} else {
			return "", "", echo.NewHTTPError(http.StatusInternalServerError)
		}
	}
	return res.Title, NullTimeToString(res.ResTimeLimit), nil
}
