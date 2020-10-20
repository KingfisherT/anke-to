package router

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/traPtitech/anke-to/model"
)

const (
	userIDKey          = "userID"
	questionnaireIDKey = "questionnaireID"
	responseIDKey = "responseID"
)

/* 消せないアンケートの発生を防ぐための管理者
暫定的にハードコーディングで対応*/
var adminUserIDs = []string{"temma", "sappi_red", "ryoha", "mazrean", "YumizSui", "pure_white_404"}

// UserAuthenticate traPのメンバーかの認証
func UserAuthenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := model.GetUserID(c)
		// トークンを持たないユーザはアクセスできない
		if userID == "-" {
			return echo.NewHTTPError(http.StatusUnauthorized, "You are not logged in")
		}

		c.Set(userIDKey, userID)

		return next(c)
	}
}

// QuestionnaireAdministratorAuthenticate アンケートの管理者かどうかの認証
func QuestionnaireAdministratorAuthenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, err := getUserID(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get userID: %w", err))
		}

		strQuestionnaireID := c.Param("questionnaireID")
		questionnaireID, err := strconv.Atoi(strQuestionnaireID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid questionnaireID:%s(error: %w)", strQuestionnaireID, err))
		}

		for _, adminID := range adminUserIDs {
			if userID == adminID {
				c.Set(questionnaireIDKey, questionnaireID)

				return next(c)
			}
		}
		isAdmin, err := model.CheckQuestionnaireAdmin(userID, questionnaireID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to check if you are administrator: %w", err))
		}
		if !isAdmin {
			return c.String(http.StatusForbidden, "You are not a administrator of this questionnaire.")
		}

		c.Set(questionnaireIDKey, questionnaireID)

		return next(c)
	}
}

func RespondentAuthenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, err := getUserID(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get userID: %w", err))
		}

		strResponseID := c.Param("responseID")
		responseID, err := strconv.Atoi(strResponseID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid responseID:%s(error: %w)", strResponseID, err))
		}

		isRespondent, err := model.CheckRespondent(userID, responseID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to check if you are a respondent: %w", err))
		}
		if !isRespondent {
			return c.String(http.StatusForbidden, "You are not a respondent of this response.")
		}

		c.Set(responseIDKey, responseID)

		return next(c)
	}
}

func getUserID(c echo.Context) (string, error) {
	rowUserID := c.Get(userIDKey)
	userID, ok := rowUserID.(string)
	if !ok {
		return "", errors.New("invalid context userID")
	}

	return userID, nil
}

func getQuestionnaireID(c echo.Context) (int, error) {
	rowQuestionnaireID := c.Get(questionnaireIDKey)
	questionnaireID, ok := rowQuestionnaireID.(int)
	if !ok {
		return 0, errors.New("invalid context userID")
	}

	return questionnaireID, nil
}
