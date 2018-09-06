package main

import (
	"net/http"
	"strconv"
	"time"

	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
)

type questionnaires struct {
	ID           int            `json:"questionnaireID" db:"id"`
	Title        string         `json:"title"           db:"title"`
	Description  string         `json:"description"     db:"description"`
	ResTimeLimit mysql.NullTime `json:"res_time_limit"  db:"res_time_limit"`
	DeletedAt    mysql.NullTime `json:"deleted_at"      db:"deleted_at"`
	ResSharedTo  string         `json:"res_shared_to"   db:"res_shared_to"`
	CreatedAt    time.Time      `json:"created_at"      db:"created_at"`
	ModifiedAt   time.Time      `json:"modified_at"     db:"modified_at"`
}

type questions struct {
	ID              int            `json:"id"                  db:"id"`
	QuestionnaireId int            `json:"questionnaireID"     db:"questionnaire_id"`
	PageNum         int            `json:"page_num"            db:"page_num"`
	QuestionNum     int            `json:"question_num"        db:"question_num"`
	Type            string         `json:"type"                db:"type"`
	Body            string         `json:"body"                db:"body"`
	IsRequrired     bool           `json:"is_required"         db:"is_required"`
	DeletedAt       mysql.NullTime `json:"deleted_at"          db:"deleted_at"`
	CreatedAt       time.Time      `json:"created_at"          db:"created_at"`
}

type scaleLabels struct {
	ID        int    `json:"questionID" db:"question_id"`
	BodyLeft  string `json:"body_left"  db:"body_left"`
	BodyRight string `json:"body_right" db:"body_right"`
}

func timeConvert(time mysql.NullTime) string {
	if time.Valid {
		return time.Time.String()
	} else {
		return "NULL"
	}
}

func getID(c echo.Context) error {
	user := c.Request().Header.Get("X-Showcase-User")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"traqID": user,
	})
}

// echoに追加するハンドラーは型に注意
// echo.Contextを引数にとってerrorを返り値とする
func getQuestionnaires(c echo.Context) error {
	// query parametar
	sort := c.QueryParam("sort")
	page := c.QueryParam("page")
	//nontargeted := c.QueryParam("nontargeted") == "true"

	if page == "" {
		page = "1"
	}
	num, err := strconv.Atoi(page)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
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
	allquestionnaires := []questionnaires{}

	if err := db.Select(&allquestionnaires,
		"SELECT * FROM questionnaires "+list[sort]+" lIMIT 20 OFFSET "+strconv.Itoa(20*(num-1))); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	type questionnairesInfo struct {
		ID           int       `json:"questionnaireID"`
		Title        string    `json:"title"`
		Description  string    `json:"description"`
		ResTimeLimit string    `json:"res_time_limit"`
		DeletedAt    string    `json:"deleted_at"`
		ResSharedTo  string    `json:"res_shared_to"`
		CreatedAt    time.Time `json:"created_at"`
		ModifiedAt   time.Time `json:"modified_at"`
		IsTargeted   bool      `json:"is_targeted"`
	}
	var ret []questionnairesInfo

	for _, v := range allquestionnaires {
		ret = append(ret,
			questionnairesInfo{
				ID:           v.ID,
				Title:        v.Title,
				Description:  v.Description,
				ResTimeLimit: timeConvert(v.ResTimeLimit),
				DeletedAt:    timeConvert(v.DeletedAt),
				ResSharedTo:  v.ResSharedTo,
				CreatedAt:    v.CreatedAt,
				ModifiedAt:   v.ModifiedAt,
				// とりあえず仮でtrueにしている
				IsTargeted: true})
	}

	// 構造体の定義で書いたJSONのキーで変換される
	return c.JSON(http.StatusOK, ret)
}

func getQuestionnaire(c echo.Context) error {

	questionnaireID := c.Param("id")

	questionnaire := questionnaires{}
	if err := db.Get(&questionnaire, "SELECT * FROM questionnaires WHERE id = ?", questionnaireID); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	targets := []string{}
	if err := db.Select(&targets, "SELECT user_traqid FROM targets WHERE questionnaire_id = ?", questionnaireID); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	administrators := []string{}
	if err := db.Select(&administrators, "SELECT user_traqid FROM administrators WHERE questionnaire_id = ?", questionnaireID); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	respondents := []string{}
	if err := db.Select(&respondents, "SELECT user_traqid FROM respondents WHERE questionnaire_id = ?", questionnaireID); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"questionnaireID": questionnaire.ID,
		"title":           questionnaire.Title,
		"description":     questionnaire.Description,
		"res_time_limit":  timeConvert(questionnaire.ResTimeLimit),
		"deleted_at":      timeConvert(questionnaire.DeletedAt),
		"created_at":      questionnaire.CreatedAt,
		"modified_at":     questionnaire.ModifiedAt,
		"res_shared_to":   questionnaire.ResSharedTo,
		"targets":         targets,
		"administrators":  administrators,
		"respondents":     respondents,
	})
}

func getQuestions(c echo.Context) error {
	questionnaireID := c.Param("id")
	// 質問一覧の配列
	allquestions := []questions{}

	// アンケートidの一致する質問を取る
	if err := db.Select(
		&allquestions, "SELECT * FROM questions WHERE questionnaire_id = ?", questionnaireID); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	type questionInfo struct {
		QuestionID      int       `json:"question_ID"`
		PageNum         int       `json:"page_num"`
		QuestionNum     int       `json:"question_num"`
		QuestionType    string    `json:"question_type"`
		Body            string    `json:"body"`
		IsRequrired     bool      `json:"is_required"`
		CreatedAt       time.Time `json:"created_at"`
		Options         []string  `json:"options"`
		ScaleLabelRight string    `json:"scale_label_right"`
		ScaleLabelLeft  string    `json:"scale_label_left"`
	}
	var ret []questionInfo

	for _, v := range allquestions {
		options := []string{}
		if err := db.Select(
			&options, "SELECT body FROM options WHERE question_id = ? ORDER BY option_num",
			v.ID); err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		scalelabel := scaleLabels{}
		if err := db.Get(&scalelabel, "SELECT * FROM scale_labels WHERE question_id = ?", v.ID); err != nil {
			if err != sql.ErrNoRows {
				return c.JSON(http.StatusInternalServerError, err)
			} else {
				scalelabel.BodyLeft = ""
				scalelabel.BodyRight = ""
			}
		}

		ret = append(ret,
			questionInfo{
				QuestionID:      v.ID,
				PageNum:         v.PageNum,
				QuestionNum:     v.QuestionNum,
				QuestionType:    v.Type,
				Body:            v.Body,
				IsRequrired:     v.IsRequrired,
				CreatedAt:       v.CreatedAt,
				Options:         options,
				ScaleLabelRight: scalelabel.BodyRight,
				ScaleLabelLeft:  scalelabel.BodyLeft,
			})
	}

	return c.JSON(http.StatusOK, ret)
}

func postQuestionnaire(c echo.Context) error {

	// リクエストで投げられるJSONのスキーマ
	req := struct {
		Title          string    `json:"title"`
		Description    string    `json:"description"`
		ResTimeLimit   time.Time `json:"res_time_limit"`
		ResSharedTo    string    `json:"res_shared_to"`
		Targets        []string  `json:"targets"`
		Administrators []string  `json:"administrators"`
	}{}

	// JSONを構造体につける
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "title is null"})
	}

	if req.ResSharedTo == "" {
		req.ResSharedTo = "administrators"
	}

	var result sql.Result

	// アンケートの追加
	if req.ResTimeLimit.IsZero() {
		var err error
		result, err = db.Exec(
			"INSERT INTO questionnaires (title, description, res_shared_to) VALUES (?, ?, ?)",
			req.Title, req.Description, req.ResSharedTo)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
	} else {
		var err error
		result, err = db.Exec(
			"INSERT INTO questionnaires (title, description, res_time_limit, res_shared_to) VALUES (?, ?, ?, ?)",
			req.Title, req.Description, req.ResTimeLimit, req.ResSharedTo)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	// エラーチェック
	lastID, err := result.LastInsertId()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	for _, v := range req.Targets {
		if _, err := db.Exec(
			"INSERT INTO targets (questionnaire_id, user_traqid) VALUES (?, ?)",
			lastID, v); err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	for _, v := range req.Administrators {
		if _, err := db.Exec(
			"INSERT INTO administrators (questionnaire_id, user_traqid) VALUES (?, ?)",
			lastID, v); err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"questionnaireID": int(lastID),
		"title":           req.Title,
		"description":     req.Description,
		"res_time_limit":  req.ResTimeLimit,
		"deleted_at":      "NULL",
		"created_at":      time.Now(),
		"modified_at":     time.Now(),
		"res_shared_to":   req.ResSharedTo,
		"targets":         req.Targets,
		"administrators":  req.Administrators,
	})
}

func editQuestionnaire(c echo.Context) error {
	questionnaireID := c.Param("id")

	req := struct {
		Title          string    `json:"title"`
		Description    string    `json:"description"`
		ResTimeLimit   time.Time `json:"res_time_limit"`
		ResSharedTo    string    `json:"res_shared_to"`
		Targets        []string  `json:"targets"`
		Administrators []string  `json:"administrators"`
	}{}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "title is null"})
	}

	if req.ResSharedTo == "" {
		req.ResSharedTo = "administrators"
	}

	// アップデートする
	if req.ResTimeLimit.IsZero() {
		if _, err := db.Exec(
			"UPDATE questionnaires SET title = ?, description = ?, res_shared_to = ?, modified_at = CURRENT_TIMESTAMP WHERE id = ?",
			req.Title, req.Description, req.ResSharedTo, questionnaireID); err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
	} else {
		if _, err := db.Exec(
			"UPDATE questionnaires SET title = ?, description = ?, res_time_limit = ?, res_shared_to = ?, modified_at = CURRENT_TIMESTAMP WHERE id = ?",
			req.Title, req.Description, req.ResTimeLimit, req.ResSharedTo, questionnaireID); err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	// TargetsとAdministratorsの変更はまだ

	return c.NoContent(http.StatusOK)
}

func deleteQuestionnaire(c echo.Context) error {
	questionnaireID := c.Param("id")

	if _, err := db.Exec(
		"UPDATE questionnaires SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?", questionnaireID); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.NoContent(http.StatusOK)
}

func getQuestion(c echo.Context) error {

	questionID := c.Param("id")

	question := questions{}
	if err := db.Get(&question, "SELECT * FROM questions WHERE id = ?", questionID); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	options := []string{}
	if err := db.Select(&options, "SELECT body FROM options WHERE question_id = ? ORDER BY option_num", questionID); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	scalelabel := scaleLabels{}
	if err := db.Get(&scalelabel, "SELECT * FROM scale_labels WHERE question_id = ?", questionID); err != nil {
		if err != sql.ErrNoRows {
			return c.JSON(http.StatusInternalServerError, err)
		} else {
			scalelabel.BodyLeft = ""
			scalelabel.BodyRight = ""
		}
	}

	type questionInfo struct {
		QuestionID      int       `json:"question_ID"`
		PageNum         int       `json:"page_num"`
		QuestionNum     int       `json:"question_num"`
		QuestionType    string    `json:"question_type"`
		Body            string    `json:"body"`
		IsRequrired     bool      `json:"is_required"`
		CreatedAt       time.Time `json:"created_at"`
		Options         []string  `json:"options"`
		ScaleLabelRight string    `json:"scale_label_right"`
		ScaleLabelLeft  string    `json:"scale_label_left"`
	}

	return c.JSON(http.StatusOK,
		questionInfo{
			QuestionID:      question.ID,
			PageNum:         question.PageNum,
			QuestionNum:     question.QuestionNum,
			QuestionType:    question.Type,
			Body:            question.Body,
			IsRequrired:     question.IsRequrired,
			CreatedAt:       question.CreatedAt,
			Options:         options,
			ScaleLabelRight: scalelabel.BodyRight,
			ScaleLabelLeft:  scalelabel.BodyLeft,
		})
}

func postQuestion(c echo.Context) error {
	req := struct {
		QuestionnaireID int      `json:"questionnaireID"`
		QuestionType    string   `json:"question_type"`
		QuestionNum     int      `json:"question_num"`
		PageNum         int      `json:"page_num"`
		Body            string   `json:"body"`
		IsRequrired     bool     `json:"is_required"`
		Options         []string `json:"options"`
		ScaleLabelRight string   `json:"scale_label_right"`
		ScaleLabelLeft  string   `json:"scale_label_left"`
	}{}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	result, err := db.Exec(
		"INSERT INTO questions (questionnaire_id, page_num, question_num, type, body, is_required) VALUES (?, ?, ?, ?, ?, ?)",
		req.QuestionnaireID, req.QuestionNum, req.PageNum, req.QuestionType, req.Body, req.IsRequrired)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	lastID, err2 := result.LastInsertId()
	if err2 != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	for i, v := range req.Options {
		if _, err := db.Exec(
			"INSERT INTO options (question_id, option_num, body) VALUES (?, ?, ?)",
			lastID, i+1, v); err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	if req.ScaleLabelLeft != "" || req.ScaleLabelRight != "" {
		if _, err := db.Exec(
			"INSERT INTO scale_labels (question_id, body_left, body_right) VALUES (?, ?, ?)",
			lastID, req.ScaleLabelLeft, req.ScaleLabelRight); err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"questionID":        int(lastID),
		"questionnaireID":   req.QuestionnaireID,
		"question_type":     req.QuestionType,
		"question_num":      req.QuestionNum,
		"page_num":          req.PageNum,
		"body":              req.Body,
		"is_required":       req.IsRequrired,
		"options":           req.Options,
		"scale_label_right": req.ScaleLabelRight,
		"scale_label_left":  req.ScaleLabelLeft,
	})
}

func editQuestion(c echo.Context) error {
	questionID := c.Param("id")

	req := struct {
		QuestionnaireID int      `json:"questionnaireID"`
		QuestionType    string   `json:"question_type"`
		QuestionNum     int      `json:"question_num"`
		PageNum         int      `json:"page_num"`
		Body            string   `json:"body"`
		IsRequrired     bool     `json:"is_required"`
		Options         []string `json:"options"`
		ScaleLabelRight string   `json:"scale_label_right"`
		ScaleLabelLeft  string   `json:"scale_label_left"`
	}{}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if _, err := db.Exec(
		"UPDATE questions SET questionnaire_id = ?, page_num = ?, question_num = ?, type = ?, body = ?, is_required = ? WHERE id = ?",
		req.QuestionnaireID, req.PageNum, req.QuestionNum, req.QuestionType, req.Body, req.IsRequrired, questionID); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.NoContent(http.StatusOK)
}

func deleteQuestion(c echo.Context) error {
	questionID := c.Param("id")

	if _, err := db.Exec(
		"UPDATE questions SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?", questionID); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.NoContent(http.StatusOK)
}
