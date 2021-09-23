package router

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/anke-to/model"
	"github.com/traPtitech/anke-to/model/mock_model"
)

type CallChecker struct {
	IsCalled bool
}

func (cc *CallChecker) Handler(c echo.Context) error {
	cc.IsCalled = true

	return c.NoContent(http.StatusOK)
}

func TestSetUserIDMiddleware(t *testing.T) {
	t.Parallel()

	assertion := assert.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRespondent := mock_model.NewMockIRespondent(ctrl)
	mockAdministrator := mock_model.NewMockIAdministrator(ctrl)
	mockQuestionnaire := mock_model.NewMockIQuestionnaire(ctrl)
	mockQuestion := mock_model.NewMockIQuestion(ctrl)

	middleware := NewMiddleware(mockAdministrator, mockRespondent, mockQuestion, mockQuestionnaire)

	type args struct {
		userID string
	}
	type expect struct {
		userID interface{}
	}
	type test struct {
		description string
		args
		expect
	}

	testCases := []test{
		{
			description: "正常なユーザーIDなのでユーザーID取得",
			args: args{
				userID: "mazrean",
			},
			expect: expect{
				userID: "mazrean",
			},
		},
		{
			description: "ユーザーIDが空なのでmds_boy",
			args: args{
				userID: "",
			},
			expect: expect{
				userID: "mds_boy",
			},
		},
		{
			description: "ユーザーIDが-なので-",
			args: args{
				userID: "-",
			},
			expect: expect{
				userID: "-",
			},
		},
	}

	for _, testCase := range testCases {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		req.Header.Set("X-Showcase-User", testCase.args.userID)

		e.HTTPErrorHandler(middleware.SetUserIDMiddleware(func(c echo.Context) error {
			assertion.Equal(testCase.expect.userID, c.Get(userIDKey), testCase.description, "userID")
			return c.NoContent(http.StatusOK)
		})(c), c)

		assertion.Equal(http.StatusOK, rec.Code, testCase.description, "status code")
	}
}

func TestResponseReadAuthenticate(t *testing.T) {
	t.Parallel()

	assertion := assert.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRespondent := mock_model.NewMockIRespondent(ctrl)
	mockAdministrator := mock_model.NewMockIAdministrator(ctrl)
	mockQuestionnaire := mock_model.NewMockIQuestionnaire(ctrl)
	mockQuestion := mock_model.NewMockIQuestion(ctrl)

	middleware := NewMiddleware(mockAdministrator, mockRespondent, mockQuestion, mockQuestionnaire)

	type args struct {
		isRespondent                                  bool
		CheckRespondentByResponseIDError              error
		haveReadPrivilege                             bool
		GetResponseReadPrivilegeInfoByResponseIDError error
		checkResponseReadPrivilegeError               error
	}
	type expect struct {
		statusCode int
		isCalled   bool
	}
	type test struct {
		description string
		args
		expect
	}

	testCases := []test{
		{
			description: "この回答の回答者である場合通す",
			args: args{
				isRespondent: true,
			},
			expect: expect{
				statusCode: http.StatusOK,
				isCalled:   true,
			},
		},
		{
			description: "CheckRespondentByResponseIDがエラーの場合500",
			args: args{
				CheckRespondentByResponseIDError: errors.New("error"),
			},
			expect: expect{
				statusCode: http.StatusInternalServerError,
				isCalled:   false,
			},
		},
		{
			description: "この回答の回答者でなくてもhaveReadPrivilegeがtrueの場合通す",
			args: args{
				isRespondent:      false,
				haveReadPrivilege: true,
			},
			expect: expect{
				statusCode: http.StatusOK,
				isCalled:   true,
			},
		},
		{
			description: "この回答の回答者でなく、haveReadPrivilegeがfalseの場合403",
			args: args{
				isRespondent:      false,
				haveReadPrivilege: false,
			},
			expect: expect{
				statusCode: http.StatusForbidden,
				isCalled:   false,
			},
		},
		{
			description: "GetResponseReadPrivilegeInfoByResponseIDがErrRecordNotFoundの場合400",
			args: args{
				isRespondent:      false,
				haveReadPrivilege: false,
				GetResponseReadPrivilegeInfoByResponseIDError: model.ErrRecordNotFound,
			},
			expect: expect{
				statusCode: http.StatusBadRequest,
				isCalled:   false,
			},
		},
		{
			description: "GetResponseReadPrivilegeInfoByResponseIDがエラー(ErrRecordNotFound以外)の場合500",
			args: args{
				isRespondent:      false,
				haveReadPrivilege: false,
				GetResponseReadPrivilegeInfoByResponseIDError: errors.New("error"),
			},
			expect: expect{
				statusCode: http.StatusInternalServerError,
				isCalled:   false,
			},
		},
		{
			description: "checkResponseReadPrivilegeがエラーの場合500",
			args: args{
				isRespondent:                    false,
				haveReadPrivilege:               false,
				checkResponseReadPrivilegeError: errors.New("error"),
			},
			expect: expect{
				statusCode: http.StatusInternalServerError,
				isCalled:   false,
			},
		},
	}

	for _, testCase := range testCases {
		userID := "testUser"
		responseID := 1
		var responseReadPrivilegeInfo model.ResponseReadPrivilegeInfo
		if testCase.args.checkResponseReadPrivilegeError != nil {
			responseReadPrivilegeInfo = model.ResponseReadPrivilegeInfo{
				ResSharedTo: "invalid value",
			}
		} else if testCase.args.haveReadPrivilege {
			responseReadPrivilegeInfo = model.ResponseReadPrivilegeInfo{
				ResSharedTo: "public",
			}
		} else {
			responseReadPrivilegeInfo = model.ResponseReadPrivilegeInfo{
				ResSharedTo: "administrators",
			}
		}

		mockRespondent.
			EXPECT().
			CheckRespondentByResponseID(userID, responseID).
			Return(testCase.args.isRespondent, testCase.args.CheckRespondentByResponseIDError)
		if !testCase.args.isRespondent && testCase.args.CheckRespondentByResponseIDError == nil {
			mockQuestionnaire.
				EXPECT().
				GetResponseReadPrivilegeInfoByResponseID(userID, responseID).
				Return(&responseReadPrivilegeInfo, testCase.args.GetResponseReadPrivilegeInfoByResponseIDError)
		}

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/responses/:responseID")
		c.SetParamNames("responseID")
		c.SetParamValues(strconv.Itoa(responseID))
		c.Set(userIDKey, userID)

		callChecker := CallChecker{}

		e.HTTPErrorHandler(middleware.ResponseReadAuthenticate(callChecker.Handler)(c), c)

		assertion.Equalf(testCase.expect.statusCode, rec.Code, testCase.description, "status code")
		assertion.Equalf(testCase.expect.isCalled, callChecker.IsCalled, testCase.description, "isCalled")
	}
}

func TestResultAuthenticate(t *testing.T) {
	t.Parallel()

	assertion := assert.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRespondent := mock_model.NewMockIRespondent(ctrl)
	mockAdministrator := mock_model.NewMockIAdministrator(ctrl)
	mockQuestionnaire := mock_model.NewMockIQuestionnaire(ctrl)
	mockQuestion := mock_model.NewMockIQuestion(ctrl)

	middleware := NewMiddleware(mockAdministrator, mockRespondent, mockQuestion, mockQuestionnaire)

	type args struct {
		haveReadPrivilege                             bool
		GetResponseReadPrivilegeInfoByResponseIDError error
		checkResponseReadPrivilegeError               error
	}
	type expect struct {
		statusCode int
		isCalled   bool
	}
	type test struct {
		description string
		args
		expect
	}

	testCases := []test{
		{
			description: "haveReadPrivilegeがtrueの場合通す",
			args: args{
				haveReadPrivilege: true,
			},
			expect: expect{
				statusCode: http.StatusOK,
				isCalled:   true,
			},
		},
		{
			description: "haveReadPrivilegeがfalseの場合403",
			args: args{
				haveReadPrivilege: false,
			},
			expect: expect{
				statusCode: http.StatusForbidden,
				isCalled:   false,
			},
		},
		{
			description: "GetResponseReadPrivilegeInfoByResponseIDがErrRecordNotFoundの場合400",
			args: args{
				haveReadPrivilege: false,
				GetResponseReadPrivilegeInfoByResponseIDError: model.ErrRecordNotFound,
			},
			expect: expect{
				statusCode: http.StatusBadRequest,
				isCalled:   false,
			},
		},
		{
			description: "GetResponseReadPrivilegeInfoByResponseIDがエラー(ErrRecordNotFound以外)の場合500",
			args: args{
				haveReadPrivilege: false,
				GetResponseReadPrivilegeInfoByResponseIDError: errors.New("error"),
			},
			expect: expect{
				statusCode: http.StatusInternalServerError,
				isCalled:   false,
			},
		},
		{
			description: "checkResponseReadPrivilegeがエラーの場合500",
			args: args{
				haveReadPrivilege:               false,
				checkResponseReadPrivilegeError: errors.New("error"),
			},
			expect: expect{
				statusCode: http.StatusInternalServerError,
				isCalled:   false,
			},
		},
	}

	for _, testCase := range testCases {
		userID := "testUser"
		questionnaireID := 1
		var responseReadPrivilegeInfo model.ResponseReadPrivilegeInfo
		if testCase.args.checkResponseReadPrivilegeError != nil {
			responseReadPrivilegeInfo = model.ResponseReadPrivilegeInfo{
				ResSharedTo: "invalid value",
			}
		} else if testCase.args.haveReadPrivilege {
			responseReadPrivilegeInfo = model.ResponseReadPrivilegeInfo{
				ResSharedTo: "public",
			}
		} else {
			responseReadPrivilegeInfo = model.ResponseReadPrivilegeInfo{
				ResSharedTo: "administrators",
			}
		}

		mockQuestionnaire.
			EXPECT().
			GetResponseReadPrivilegeInfoByQuestionnaireID(userID, questionnaireID).
			Return(&responseReadPrivilegeInfo, testCase.args.GetResponseReadPrivilegeInfoByResponseIDError)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/results/%d", questionnaireID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/results/:questionnaireID")
		c.SetParamNames("questionnaireID")
		c.SetParamValues(strconv.Itoa(questionnaireID))
		c.Set(userIDKey, userID)

		callChecker := CallChecker{}

		e.HTTPErrorHandler(middleware.ResultAuthenticate(callChecker.Handler)(c), c)

		assertion.Equalf(testCase.expect.statusCode, rec.Code, testCase.description, "status code")
		assertion.Equalf(testCase.expect.isCalled, callChecker.IsCalled, testCase.description, "isCalled")
	}
}

func TestCheckResponseReadPrivilege(t *testing.T) {
	t.Parallel()

	assertion := assert.New(t)

	type args struct {
		responseReadPrivilegeInfo model.ResponseReadPrivilegeInfo
	}
	type expect struct {
		haveReadPrivilege bool
		isErr             bool
		err               error
	}
	type test struct {
		description string
		args
		expect
	}

	testCases := []test{
		{
			description: "res_shared_toがpublic、administrators、respondentsのいずれでもない場合エラー",
			args: args{
				responseReadPrivilegeInfo: model.ResponseReadPrivilegeInfo{
					ResSharedTo: "invalid value",
				},
			},
			expect: expect{
				isErr: true,
			},
		},
		{
			description: "res_shared_toがpublicの場合true",
			args: args{
				responseReadPrivilegeInfo: model.ResponseReadPrivilegeInfo{
					ResSharedTo: "public",
				},
			},
			expect: expect{
				haveReadPrivilege: true,
			},
		},
		{
			description: "res_shared_toがadministratorsかつadministratorの場合true",
			args: args{
				responseReadPrivilegeInfo: model.ResponseReadPrivilegeInfo{
					ResSharedTo:     "administrators",
					IsAdministrator: true,
				},
			},
			expect: expect{
				haveReadPrivilege: true,
			},
		},
		{
			description: "res_shared_toがadministratorsかつadministratorでない場合false",
			args: args{
				responseReadPrivilegeInfo: model.ResponseReadPrivilegeInfo{
					ResSharedTo:     "administrators",
					IsAdministrator: false,
				},
			},
		},
		{
			description: "res_shared_toがrespondentsかつadministratorの場合true",
			args: args{
				responseReadPrivilegeInfo: model.ResponseReadPrivilegeInfo{
					ResSharedTo:     "respondents",
					IsAdministrator: true,
				},
			},
			expect: expect{
				haveReadPrivilege: true,
			},
		},
		{
			description: "res_shared_toがrespondentsかつrespondentの場合true",
			args: args{
				responseReadPrivilegeInfo: model.ResponseReadPrivilegeInfo{
					ResSharedTo:  "respondents",
					IsRespondent: true,
				},
			},
			expect: expect{
				haveReadPrivilege: true,
			},
		},
		{
			description: "res_shared_toがrespondentsかつ、administratorでもrespondentでない場合false",
			args: args{
				responseReadPrivilegeInfo: model.ResponseReadPrivilegeInfo{
					ResSharedTo:     "respondents",
					IsAdministrator: false,
					IsRespondent:    false,
				},
			},
			expect: expect{
				haveReadPrivilege: false,
			},
		},
	}

	for _, testCase := range testCases {
		haveReadPrivilege, err := checkResponseReadPrivilege(&testCase.args.responseReadPrivilegeInfo)

		if testCase.expect.isErr {
			assertion.Errorf(err, testCase.description, "error")
		} else {
			assertion.NoErrorf(err, testCase.description, "no error")
			assertion.Equalf(testCase.expect.haveReadPrivilege, haveReadPrivilege, testCase.description, "haveReadPrivilege")
		}
	}
}
