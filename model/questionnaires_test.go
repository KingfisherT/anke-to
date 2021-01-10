package model

import (
	"errors"
	"math"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v3"
)

const questionnairesTestUserID = "questionnairesUser"
const invalidQuestionnairesTestUserID = "invalidQuestionnairesUser"

var questionnairesNow = time.Now()

type QuestionnairesTestData struct {
	questionnaire  Questionnaires
	targets        []string
	administrators []string
}

var (
	datas                   = []QuestionnairesTestData{}
	deletedQuestionnaireIDs = []int{}
	userTargetMap           = map[string][]int{}
	userAdministratorMap    = map[string][]int{}
)

func TestQuestionnaires(t *testing.T) {
	t.Parallel()

	setupQuestionnairesTest(t)

	t.Run("InsertQuestionnaire", insertQuestionnaireTest)
	t.Run("UpdateQuestionnaire", updateQuestionnaireTest)
	t.Run("DeleteQuestionnaire", deleteQuestionnaireTest)
	t.Run("GetQuestionnaires", getQuestionnairesTest)
	t.Run("GetAdminQuestionnaires", getAdminQuestionnairesTest)
}

func setupQuestionnairesTest(t *testing.T) {
	datas = []QuestionnairesTestData{
		{
			questionnaire: Questionnaires{
				Title:        "第1回集会らん☆ぷろ募集アンケートGetQuestionnaireTest",
				Description:  "第1回集会らん☆ぷろ参加者募集",
				ResTimeLimit: null.NewTime(time.Time{}, false),
				ResSharedTo:  "public",
				CreatedAt:    questionnairesNow,
				ModifiedAt:   questionnairesNow,
			},
			targets:        []string{},
			administrators: []string{},
		},
		{
			questionnaire: Questionnaires{
				Title:        "第1回集会らん☆ぷろ募集アンケートGetQuestionnaireTest",
				Description:  "第1回集会らん☆ぷろ参加者募集",
				ResTimeLimit: null.NewTime(time.Time{}, false),
				ResSharedTo:  "public",
				CreatedAt:    questionnairesNow.Add(time.Second),
				ModifiedAt:   questionnairesNow.Add(2 * time.Second),
			},
			targets:        []string{},
			administrators: []string{questionnairesTestUserID},
		},
		{
			questionnaire: Questionnaires{
				Title:        "第1回集会らん☆ぷろ募集アンケートGetQuestionnaireTest",
				Description:  "第1回集会らん☆ぷろ参加者募集",
				ResTimeLimit: null.NewTime(time.Time{}, false),
				ResSharedTo:  "public",
				CreatedAt:    questionnairesNow.Add(2 * time.Second),
				ModifiedAt:   questionnairesNow.Add(3 * time.Second),
			},
			targets:        []string{questionnairesTestUserID},
			administrators: []string{questionnairesTestUserID},
		},
		{
			questionnaire: Questionnaires{
				Title:        "第1回集会らん☆ぷろ募集アンケートGetQuestionnaireTest",
				Description:  "第1回集会らん☆ぷろ参加者募集",
				ResTimeLimit: null.NewTime(time.Time{}, false),
				ResSharedTo:  "public",
				CreatedAt:    questionnairesNow,
				ModifiedAt:   questionnairesNow,
				DeletedAt:    null.NewTime(questionnairesNow, true),
			},
			targets:        []string{},
			administrators: []string{},
		},
	}
	for i := 0; i < 20; i++ {
		datas = append(datas, QuestionnairesTestData{
			questionnaire: Questionnaires{
				Title:        "第1回集会らん☆ぷろ募集アンケート",
				Description:  "第1回集会らん☆ぷろ参加者募集",
				ResTimeLimit: null.NewTime(time.Time{}, false),
				ResSharedTo:  "public",
				CreatedAt:    questionnairesNow.Add(time.Duration(len(datas)) * time.Second),
				ModifiedAt:   questionnairesNow,
			},
			targets:        []string{},
			administrators: []string{},
		})
	}
	datas = append(datas, QuestionnairesTestData{
		questionnaire: Questionnaires{
			Title:        "第1回集会らん☆ぷろ募集アンケートGetQuestionnaireTest",
			Description:  "第1回集会らん☆ぷろ参加者募集",
			ResTimeLimit: null.NewTime(time.Time{}, false),
			ResSharedTo:  "public",
			CreatedAt:    questionnairesNow.Add(2 * time.Second),
			ModifiedAt:   questionnairesNow.Add(3 * time.Second),
		},
		targets:        []string{questionnairesTestUserID},
		administrators: []string{questionnairesTestUserID},
	})

	for i, data := range datas {
		if data.questionnaire.DeletedAt.Valid {
			deletedQuestionnaireIDs = append(deletedQuestionnaireIDs, data.questionnaire.ID)
		}

		err := db.Create(&datas[i].questionnaire).Error
		if err != nil {
			t.Errorf("failed to create questionnaire(%+v): %w", data, err)
		}

		for _, target := range data.targets {
			questionnaires, ok := userTargetMap[target]
			if !ok {
				questionnaires = []int{}
			}
			userTargetMap[target] = append(questionnaires, data.questionnaire.ID)

			err := db.Create(&Targets{
				QuestionnaireID: datas[i].questionnaire.ID,
				UserTraqid:      target,
			}).Error
			if err != nil {
				t.Errorf("failed to create target: %w", err)
			}
		}

		for _, administrator := range data.administrators {
			questionnaires, ok := userAdministratorMap[administrator]
			if !ok {
				questionnaires = []int{}
			}
			userAdministratorMap[administrator] = append(questionnaires, datas[i].questionnaire.ID)

			err := db.Create(&Administrators{
				QuestionnaireID: datas[i].questionnaire.ID,
				UserTraqid:      administrator,
			}).Error
			if err != nil {
				t.Errorf("failed to create target: %w", err)
			}
		}
	}
}

func insertQuestionnaireTest(t *testing.T) {
	t.Helper()
	t.Parallel()

	assertion := assert.New(t)

	type args struct {
		title        string
		description  string
		resTimeLimit null.Time
		resSharedTo  string
	}
	type expect struct {
		isErr bool
		err   error
	}

	type test struct {
		description string
		args
		expect
	}

	testCases := []test{
		{
			description: "time limit: no, res shared to: public",
			args: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Time{}, false),
				resSharedTo:  "public",
			},
		},
		{
			description: "time limit: yes, res shared to: public",
			args: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Now(), true),
				resSharedTo:  "public",
			},
		},
		{
			description: "time limit: no, res shared to: respondents",
			args: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Time{}, false),
				resSharedTo:  "respondents",
			},
		},
		{
			description: "time limit: no, res shared to: administrators",
			args: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Time{}, false),
				resSharedTo:  "administrators",
			},
		},
		{
			description: "long title",
			args: args{
				title:        strings.Repeat("a", 50),
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Time{}, false),
				resSharedTo:  "public",
			},
		},
		{
			description: "too long title",
			args: args{
				title:        strings.Repeat("a", 500),
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Time{}, false),
				resSharedTo:  "public",
			},
			expect: expect{
				isErr: true,
			},
		},
		{
			description: "long description",
			args: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  strings.Repeat("a", 2000),
				resTimeLimit: null.NewTime(time.Time{}, false),
				resSharedTo:  "public",
			},
		},
		{
			description: "too long description",
			args: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  strings.Repeat("a", 200000),
				resTimeLimit: null.NewTime(time.Time{}, false),
				resSharedTo:  "public",
			},
			expect: expect{
				isErr: true,
			},
		},
	}

	for _, testCase := range testCases {
		questionnaireID, err := InsertQuestionnaire(testCase.args.title, testCase.args.description, testCase.args.resTimeLimit, testCase.args.resSharedTo)

		if !testCase.expect.isErr {
			assertion.NoError(err, testCase.description, "no error")
		} else if testCase.expect.err != nil {
			assertion.Equal(testCase.expect.err, err, testCase.description, "error")
		}
		if err != nil {
			continue
		}

		questionnaire := Questionnaires{}
		err = db.Where("id = ?", questionnaireID).First(&questionnaire).Error
		if err != nil {
			t.Errorf("failed to get questionnaire(%s): %w", testCase.description, err)
		}

		assertion.Equal(testCase.args.title, questionnaire.Title, testCase.description, "title")
		assertion.Equal(testCase.args.description, questionnaire.Description, testCase.description, "description")
		assertion.WithinDuration(testCase.args.resTimeLimit.ValueOrZero(), questionnaire.ResTimeLimit.ValueOrZero(), 2*time.Second, testCase.description, "res_time_limit")
		assertion.Equal(testCase.args.resSharedTo, questionnaire.ResSharedTo, testCase.description, "res_shared_to")

		assertion.WithinDuration(time.Now(), questionnaire.CreatedAt, 2*time.Second, testCase.description, "created_at")
		assertion.WithinDuration(time.Now(), questionnaire.ModifiedAt, 2*time.Second, testCase.description, "modified_at")
	}
}

func updateQuestionnaireTest(t *testing.T) {
	t.Helper()
	t.Parallel()

	assertion := assert.New(t)

	type args struct {
		title        string
		description  string
		resTimeLimit null.Time
		resSharedTo  string
	}
	type expect struct {
		isErr bool
		err   error
	}

	type test struct {
		description string
		before      args
		after       args
		expect
	}

	testCases := []test{
		{
			description: "update res_shared_to",
			before: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Time{}, false),
				resSharedTo:  "public",
			},
			after: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Time{}, false),
				resSharedTo:  "respondents",
			},
		},
		{
			description: "update title",
			before: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Time{}, false),
				resSharedTo:  "public",
			},
			after: args{
				title:        "第2回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Time{}, false),
				resSharedTo:  "public",
			},
		},
		{
			description: "update description",
			before: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Time{}, false),
				resSharedTo:  "public",
			},
			after: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第2回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Time{}, false),
				resSharedTo:  "public",
			},
		},
		{
			description: "update res_shared_to(res_time_limit is valid)",
			before: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Now(), true),
				resSharedTo:  "public",
			},
			after: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Now(), true),
				resSharedTo:  "respondents",
			},
		},
		{
			description: "update title(res_time_limit is valid)",
			before: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Now(), true),
				resSharedTo:  "public",
			},
			after: args{
				title:        "第2回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Now(), true),
				resSharedTo:  "public",
			},
		},
		{
			description: "update description(res_time_limit is valid)",
			before: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Now(), true),
				resSharedTo:  "public",
			},
			after: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第2回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Now(), true),
				resSharedTo:  "public",
			},
		},
		{
			description: "update res_time_limit(null->time)",
			before: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Time{}, false),
				resSharedTo:  "public",
			},
			after: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Now(), true),
				resSharedTo:  "public",
			},
		},
		{
			description: "update res_time_limit(time->time)",
			before: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Now(), true),
				resSharedTo:  "public",
			},
			after: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Now().Add(time.Minute), true),
				resSharedTo:  "public",
			},
		},
		{
			description: "update res_time_limit(time->null)",
			before: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Now(), true),
				resSharedTo:  "public",
			},
			after: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Time{}, false),
				resSharedTo:  "public",
			},
		},
	}

	for _, testCase := range testCases {
		before := &testCase.before
		questionnaire := Questionnaires{
			Title:        before.title,
			Description:  before.description,
			ResTimeLimit: before.resTimeLimit,
			ResSharedTo:  before.resSharedTo,
		}
		err := db.Create(&questionnaire).Error
		if err != nil {
			t.Errorf("failed to create questionnaire(%s): %w", testCase.description, err)
		}

		createdAt := questionnaire.CreatedAt
		questionnaireID := questionnaire.ID
		after := &testCase.after
		err = UpdateQuestionnaire(after.title, after.description, after.resTimeLimit, after.resSharedTo, questionnaireID)

		if !testCase.expect.isErr {
			assertion.NoError(err, testCase.description, "no error")
		} else if testCase.expect.err != nil {
			assertion.Equal(testCase.expect.err, err, testCase.description, "error")
		}
		if err != nil {
			continue
		}

		questionnaire = Questionnaires{}
		err = db.Where("id = ?", questionnaireID).First(&questionnaire).Error
		if err != nil {
			t.Errorf("failed to get questionnaire(%s): %w", testCase.description, err)
		}

		assertion.Equal(after.title, questionnaire.Title, testCase.description, "title")
		assertion.Equal(after.description, questionnaire.Description, testCase.description, "description")
		assertion.WithinDuration(after.resTimeLimit.ValueOrZero(), questionnaire.ResTimeLimit.ValueOrZero(), 2*time.Second, testCase.description, "res_time_limit")
		assertion.Equal(after.resSharedTo, questionnaire.ResSharedTo, testCase.description, "res_shared_to")

		assertion.WithinDuration(createdAt, questionnaire.CreatedAt, 2*time.Second, testCase.description, "created_at")
		assertion.WithinDuration(time.Now(), questionnaire.ModifiedAt, 2*time.Second, testCase.description, "modified_at")
	}

	invalidQuestionnaireID := 1000
	for {
		err := db.Where("id = ?", invalidQuestionnaireID).First(&Questionnaires{}).Error
		if gorm.IsRecordNotFoundError(err) {
			break
		}
		if err != nil {
			t.Errorf("failed to get questionnaire(make invalid questionnaireID): %w", err)
			break
		}

		invalidQuestionnaireID *= 10
	}

	invalidTestCases := []args{
		{
			title:        "第1回集会らん☆ぷろ募集アンケート",
			description:  "第1回集会らん☆ぷろ参加者募集",
			resTimeLimit: null.NewTime(time.Time{}, false),
			resSharedTo:  "public",
		},
		{
			title:        "第1回集会らん☆ぷろ募集アンケート",
			description:  "第1回集会らん☆ぷろ参加者募集",
			resTimeLimit: null.NewTime(time.Now(), true),
			resSharedTo:  "public",
		},
	}

	for _, arg := range invalidTestCases {
		err := UpdateQuestionnaire(arg.title, arg.description, arg.resTimeLimit, arg.resSharedTo, invalidQuestionnaireID)
		if !errors.Is(err, ErrNoRecordUpdated) {
			if err == nil {
				t.Errorf("Succeeded with invalid questionnaireID")
			} else {
				t.Errorf("failed to update questionnaire(invalid questionnireID): %w", err)
			}
		}
	}
}

func deleteQuestionnaireTest(t *testing.T) {
	t.Helper()
	t.Parallel()

	assertion := assert.New(t)

	type args struct {
		title        string
		description  string
		resTimeLimit null.Time
		resSharedTo  string
	}
	type expect struct {
		isErr bool
		err   error
	}
	type test struct {
		args
		expect
	}

	testCases := []test{
		{
			args: args{
				title:        "第1回集会らん☆ぷろ募集アンケート",
				description:  "第1回集会らん☆ぷろ参加者募集",
				resTimeLimit: null.NewTime(time.Time{}, false),
				resSharedTo:  "public",
			},
		},
	}

	for _, testCase := range testCases {
		questionnaire := Questionnaires{
			Title:        testCase.args.title,
			Description:  testCase.args.description,
			ResTimeLimit: testCase.args.resTimeLimit,
			ResSharedTo:  testCase.args.resSharedTo,
		}
		err := db.Create(&questionnaire).Error
		if err != nil {
			t.Errorf("failed to create questionnaire(%s): %w", testCase.description, err)
		}

		questionnaireID := questionnaire.ID
		err = DeleteQuestionnaire(questionnaireID)

		if !testCase.expect.isErr {
			assertion.NoError(err, testCase.description, "no error")
		} else if testCase.expect.err != nil {
			assertion.Equal(testCase.expect.err, err, testCase.description, "error")
		}
		if err != nil {
			continue
		}

		questionnaire = Questionnaires{}
		err = db.
			Unscoped().
			Where("id = ?", questionnaireID).
			Find(&questionnaire).Error
		if err != nil {
			t.Errorf("failed to get questionnaire(%s): %w", testCase.description, err)
		}

		assertion.WithinDuration(time.Now(), questionnaire.DeletedAt.ValueOrZero(), 2*time.Second)
	}

	invalidQuestionnaireID := 1000
	for {
		err := db.Where("id = ?", invalidQuestionnaireID).First(&Questionnaires{}).Error
		if gorm.IsRecordNotFoundError(err) {
			break
		}
		if err != nil {
			t.Errorf("failed to get questionnaire(make invalid questionnaireID): %w", err)
			break
		}

		invalidQuestionnaireID *= 10
	}

	err := DeleteQuestionnaire(invalidQuestionnaireID)
	if !errors.Is(err, ErrNoRecordDeleted) {
		if err == nil {
			t.Errorf("Succeeded with invalid questionnaireID")
		} else {
			t.Errorf("failed to update questionnaire(invalid questionnireID): %w", err)
		}
	}
}

func getQuestionnairesTest(t *testing.T) {
	t.Helper()

	assertion := assert.New(t)

	sortFuncMap := map[string]func(questionnaires []QuestionnaireInfo) func(i, j int) bool{
		"created_at": func(questionnaires []QuestionnaireInfo) func(i, j int) bool {
			return func(i, j int) bool {
				return questionnaires[i].CreatedAt.After(questionnaires[j].CreatedAt)
			}
		},
		"-created_at": func(questionnaires []QuestionnaireInfo) func(i, j int) bool {
			return func(i, j int) bool {
				return questionnaires[i].CreatedAt.Before(questionnaires[j].CreatedAt)
			}
		},
		"title": func(questionnaires []QuestionnaireInfo) func(i, j int) bool {
			return func(i, j int) bool {
				return questionnaires[i].Title < questionnaires[j].Title
			}
		},
		"-title": func(questionnaires []QuestionnaireInfo) func(i, j int) bool {
			return func(i, j int) bool {
				return questionnaires[i].Title > questionnaires[j].Title
			}
		},
		"modified_at": func(questionnaires []QuestionnaireInfo) func(i, j int) bool {
			return func(i, j int) bool {
				return questionnaires[i].ModifiedAt.After(questionnaires[j].ModifiedAt)
			}
		},
		"-modified_at": func(questionnaires []QuestionnaireInfo) func(i, j int) bool {
			return func(i, j int) bool {
				return questionnaires[i].CreatedAt.After(questionnaires[j].CreatedAt)
			}
		},
		"": func(questionnaires []QuestionnaireInfo) func(i, j int) bool {
			return func(i, j int) bool {
				return questionnaires[i].ID < questionnaires[j].ID
			}
		},
	}

	type args struct {
		userID      string
		sort        string
		search      string
		pageNum     int
		nontargeted bool
	}
	type expect struct {
		isErr      bool
		err        error
		isCheckLen bool
		length     int
	}
	type test struct {
		description string
		args
		expect
	}

	testCases := []test{
		{
			description: "userID:valid, sort:no, search:no, page:1",
			args: args{
				userID:      questionnairesTestUserID,
				sort:        "",
				search:      "",
				pageNum:     1,
				nontargeted: false,
			},
		},
		{
			description: "userID:valid, sort:created_at, search:no, page:1",
			args: args{
				userID:      questionnairesTestUserID,
				sort:        "created_at",
				search:      "",
				pageNum:     1,
				nontargeted: false,
			},
		},
		{
			description: "userID:valid, sort:-created_at, search:no, page:1",
			args: args{
				userID:      questionnairesTestUserID,
				sort:        "-created_at",
				search:      "",
				pageNum:     1,
				nontargeted: false,
			},
		},
		{
			description: "userID:valid, sort:title, search:no, page:1",
			args: args{
				userID:      questionnairesTestUserID,
				sort:        "title",
				search:      "",
				pageNum:     1,
				nontargeted: false,
			},
		},
		{
			description: "userID:valid, sort:-title, search:no, page:1",
			args: args{
				userID:      questionnairesTestUserID,
				sort:        "-title",
				search:      "",
				pageNum:     1,
				nontargeted: false,
			},
		},
		{
			description: "userID:valid, sort:modified_at, search:no, page:1",
			args: args{
				userID:      questionnairesTestUserID,
				sort:        "modified_at",
				search:      "",
				pageNum:     1,
				nontargeted: false,
			},
		},
		{
			description: "userID:valid, sort:-modified_at, search:no, page:1",
			args: args{
				userID:      questionnairesTestUserID,
				sort:        "-modified_at",
				search:      "",
				pageNum:     1,
				nontargeted: false,
			},
		},
		{
			description: "userID:valid, sort:no, search:GetQuestionnaireTest$, page:1",
			args: args{
				userID:      questionnairesTestUserID,
				sort:        "",
				search:      "GetQuestionnaireTest$",
				pageNum:     1,
				nontargeted: false,
			},
			expect: expect{
				isCheckLen: true,
				length:     4,
			},
		},
		{
			description: "userID:valid, sort:no, search:no, page:2",
			args: args{
				userID:      questionnairesTestUserID,
				sort:        "",
				search:      "",
				pageNum:     2,
				nontargeted: false,
			},
		},
		{
			description: "too large page",
			args: args{
				userID:      questionnairesTestUserID,
				sort:        "",
				search:      "",
				pageNum:     100000,
				nontargeted: false,
			},
			expect: expect{
				isErr: true,
				err:   ErrTooLargePageNum,
			},
		},
		{
			description: "userID:valid, sort:no, search:no, page:1, nontargetted",
			args: args{
				userID:      questionnairesTestUserID,
				sort:        "",
				search:      "",
				pageNum:     1,
				nontargeted: true,
			},
		},
		{
			description: "userID:valid, sort:no, search:notFoundQuestionnaire, page:1",
			args: args{
				userID:      questionnairesTestUserID,
				sort:        "",
				search:      "notFoundQuestionnaire",
				pageNum:     1,
				nontargeted: true,
			},
			expect: expect{
				isCheckLen: false,
				length:     0,
			},
		},
		{
			description: "userID:valid, sort:invalid, search:no, page:1",
			args: args{
				userID:      questionnairesTestUserID,
				sort:        "hogehoge",
				search:      "",
				pageNum:     1,
				nontargeted: false,
			},
			expect: expect{
				isErr: true,
				err:   ErrInvalidSortParam,
			},
		},
	}

	for _, testCase := range testCases {
		questionnaires, pageMax, err := GetQuestionnaires(testCase.args.userID, testCase.args.sort, testCase.args.search, testCase.args.pageNum, testCase.args.nontargeted)

		if !testCase.expect.isErr {
			assertion.NoError(err, testCase.description, "no error")
		} else if testCase.expect.err != nil {
			if !errors.Is(err, testCase.expect.err) {
				t.Errorf("invalid error(%s): expected: %+v, actual: %+v", testCase.description, testCase.expect.err, err)
			}
		}
		if err != nil {
			continue
		}

		var questionnaireNum int
		err = db.
			Model(&Questionnaires{}).
			Where("deleted_at IS NULL").
			Count(&questionnaireNum).Error
		if err != nil {
			t.Errorf("failed to count questionnaire(%s): %w", testCase.description, err)
		}

		actualQuestionnaireIDs := []int{}
		for _, questionnaire := range questionnaires {
			actualQuestionnaireIDs = append(actualQuestionnaireIDs, questionnaire.ID)
		}
		if testCase.args.nontargeted {
			for _, targettedQuestionnaireID := range userTargetMap[questionnairesTestUserID] {
				assertion.NotContains(actualQuestionnaireIDs, targettedQuestionnaireID, testCase.description, "not contain(targetted)")
			}
		}
		for _, deletedQuestionnaireID := range deletedQuestionnaireIDs {
			assertion.NotContains(actualQuestionnaireIDs, deletedQuestionnaireID, testCase.description, "not contain(deleted)")
		}

		for _, questionnaire := range questionnaires {
			assertion.Regexp(testCase.args.search, questionnaire.Title, testCase.description, "regexp")
		}

		if len(testCase.args.search) == 0 && !testCase.args.nontargeted {
			assertion.Equal((questionnaireNum+19)/20, pageMax, testCase.description, "pageMax")
			assertion.Len(questionnaires, int(math.Min(float64(questionnaireNum-20*(testCase.pageNum-1)), 20.0)), testCase.description, "page")
		}

		if testCase.expect.isCheckLen {
			assertion.Len(questionnaires, testCase.expect.length, testCase.description, "length")
		}

		copyQuestionnaires := make([]QuestionnaireInfo, len(questionnaires))
		copy(copyQuestionnaires, questionnaires)
		sort.SliceStable(copyQuestionnaires, sortFuncMap[testCase.args.sort](questionnaires))
		expectQuestionnaireIDs := []int{}
		for _, questionnaire := range copyQuestionnaires {
			expectQuestionnaireIDs = append(expectQuestionnaireIDs, questionnaire.ID)
		}
		assertion.ElementsMatch(expectQuestionnaireIDs, actualQuestionnaireIDs, testCase.description, "sort")
	}
}

func getAdminQuestionnairesTest(t *testing.T) {
	t.Helper()
	t.Parallel()

	assertion := assert.New(t)

	type args struct {
		userID string
	}
	type expect struct {
		isErr bool
		err   error
	}
	type test struct {
		description string
		args
		expect
	}

	testCases := []test{
		{
			description: "user:valid",
			args: args{
				userID: questionnairesTestUserID,
			},
		},
		{
			description: "empty response",
			args: args{
				userID: invalidQuestionnairesTestUserID,
			},
		},
	}

	for _, testCase := range testCases {
		questionnaires, err := GetAdminQuestionnaires(testCase.userID)

		if !testCase.expect.isErr {
			assertion.NoError(err, testCase.description, "no error")
		} else if testCase.expect.err != nil {
			if !errors.Is(err, testCase.expect.err) {
				t.Errorf("invalid error(%s): expected: %+v, actual: %+v", testCase.description, testCase.expect.err, err)
			}
		}
		if err != nil {
			continue
		}

		actualQuestionnaireIDs := make([]int, 0, len(questionnaires))
		actualIDQuestionnaireMap := make(map[int]Questionnaires, len(questionnaires))
		for _, questionnaire := range questionnaires {
			actualQuestionnaireIDs = append(actualQuestionnaireIDs, questionnaire.ID)
			actualIDQuestionnaireMap[questionnaire.ID] = questionnaire
		}

		assertion.Subset(userAdministratorMap[testCase.args.userID], actualQuestionnaireIDs, testCase.description, "administrate")

		expectQuestionnaires := []Questionnaires{}
		err = db.
			Where("id IN (?)", actualQuestionnaireIDs).
			Find(&expectQuestionnaires).Error
		if err != nil {
			t.Errorf("failed to get questionnaires(%s): %w", testCase.description, err)
		}

		for _, expectQuestionnaire := range expectQuestionnaires {
			actualQuestionnaire := actualIDQuestionnaireMap[expectQuestionnaire.ID]

			assertion.Equal(expectQuestionnaire, actualQuestionnaire, testCase.description, "questionnaire")
		}
	}
}
