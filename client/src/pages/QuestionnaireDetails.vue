<template>
  <div class="details is-fullheight">
    <top-bar-message :message="message"></top-bar-message>
    <div class="tabs is-centered">
      <ul>
        <li
          v-for="(tab, index) in detailTabs"
          :key="index"
          class="tab"
          :class="{ 'is-active': selectedTab === tab }"
          @click="selectedTab = tab"
        >
          <a>{{ tab }}</a>
        </li>
      </ul>
      <a
        v-if="showEditButton"
        id="edit-button"
        :class="{ 'is-editing': isEditing }"
        @click="toggleEditing"
      >
        <span class="ti-pencil"></span>
      </a>
    </div>
    <div
      :class="{ 'is-editing has-navbar-fixed-bottom': isEditing }"
      class="details-child is-fullheight"
    >
      <information-summary
        v-if="currentTabComponent !== 'information-edit'"
        :information="summaryProps"
      ></information-summary>
      <component
        :is="currentTabComponent"
        class="details-child is-fullheight"
        :name="currentTabComponent"
        :edit-mode="isEditing ? 'question' : undefined"
        :information-props="informationProps"
        :questions-props="questions"
        :title="title"
        :input-errors="isEditing ? inputErrors : undefined"
        @set-data="setData"
        @set-question-content="setQuestionContent"
        @remove-question="removeQuestion"
      ></component>
      <edit-nav-bar v-if="isEditing">
        <button
          class="button is-medium send-button"
          :disabled="!submitOk || isSubmitting"
          @click="submitQuestionnaire"
        >
          <span class="ti-check"></span>
          <span>送信</span>
        </button>
        <button class="button is-medium cancel-button" @click="abortEditing">
          <span class="ti-close"></span>
        </button>
      </edit-nav-bar>
    </div>
  </div>
</template>

<script>
import moment from 'moment'
import { mapGetters } from 'vuex'
import router from '@/router'
import common from '@/bin/common'
import axios from '@/bin/axios'
import InformationSummary from '@/components/Information/InformationSummary'
import Information from '@/components/Information/Information'
import InformationEdit from '@/components/Information/InformationEdit'
import Questions from '@/components/Questions/Questions'
import QuestionsEdit from '@/components/Questions/QuestionsEdit'
import EditNavBar from '@/components/Utils/EditNavBar'
import TopBarMessage from '@/components/Utils/TopBarMessage'

export default {
  name: 'QuestionnaireDetails',
  components: {
    'information-summary': InformationSummary,
    information: Information,
    'information-edit': InformationEdit,
    questions: Questions,
    'questions-edit': QuestionsEdit,
    'edit-nav-bar': EditNavBar,
    'top-bar-message': TopBarMessage
  },
  props: {},
  data() {
    return {
      detailTabs: ['Information', 'Questions'],
      showEditButton: false,
      noTimeLimit: true,
      information: {},
      questions: [],
      newQuestionnaireId: undefined,
      removedQuestionIds: [],
      message: {
        showMessage: false
      },
      isSubmitting: false
    }
  },
  computed: {
    ...mapGetters(['getMyTraqId']),
    selectedTab: {
      get() {
        return this.$route.query.tab && this.$route.query.tab === 'questions'
          ? 'Questions'
          : 'Information'
      },
      set(newTab) {
        router.push({
          name: this.$route.name,
          params: { questionnaireId: this.questionnaireId },
          query: { tab: newTab.toLowerCase() }
        })
      }
    },
    administrates() {
      // 管理者かどうかを返す
      // getInformation() が完了する前は false を返す
      return this.information.administrators
        ? common.administrates(
            this.information.administrators,
            this.getMyTraqId
          )
        : false
    },
    questionnaireId() {
      if (this.isNewQuestionnaire) {
        return undefined
      } else {
        return Number(this.$route.params.id)
      }
    },
    isNewQuestionnaire() {
      return this.$route.name === 'QuestionnaireDetailsNew'
    },
    submitOk() {
      return common.noErrors(this.inputErrors)
    },
    isEditing() {
      return (
        this.$route.name === 'QuestionnaireDetailsEdit' ||
        this.$route.name === 'QuestionnaireDetailsNew'
      )
    },
    currentTabComponent() {
      switch (this.selectedTab) {
        case 'Information':
          if (this.isEditing) {
            return 'information-edit'
          } else {
            return 'information'
          }
        case 'Questions':
          if (this.isEditing) {
            return 'questions-edit'
          } else {
            return 'questions'
          }
        default:
          console.error('unexpected selectedTab')
          return null
      }
    },
    title() {
      return this.information.title || '新規アンケート作成'
    },
    informationProps() {
      return {
        information: this.information,
        administrates: this.administrates,
        questionnaireId: this.questionnaireId,
        noTimeLimit: this.noTimeLimit
      }
    },
    summaryProps() {
      let ret = {
        title: this.information.title
      }
      if (this.selectedTab === 'Information') {
        ret.description = this.information.description
        ret.timeLimit = this.getDateStr(this.information.res_time_limit)
      }
      return ret
    },
    newTimeLimit() {
      // 1週間後の23:59
      return moment().add(7, 'days').endOf('day').format().slice(0, -6)
    },
    inputErrors() {
      return {
        noTitle: {
          message: 'タイトルは入力必須です',
          isError: this.information.title === ''
        },
        noQuestions: {
          message: '質問がありません',
          isError: this.questions.length === 0
        }
      }
    }
  },
  watch: {
    $route: function (newRoute, oldRoute) {
      if (newRoute.params.id !== oldRoute.params.id) {
        this.showEditButton = false
        this.getInformation()
        this.getQuestions()
        this.newQuestionnaireId = undefined
        if (oldRoute.name !== 'QuestionnaireDetailsNew') this.resetMessage()
      }
    },
    noTimeLimit: function (newBool, oldBool) {
      if (
        oldBool &&
        !newBool &&
        (this.information.res_time_limit === 'null' ||
          this.information.res_time_limit === '')
      ) {
        // 新しく回答期限を作ろうとしたとき
        this.information.res_time_limit = this.newTimeLimit
      }
    }
  },
  async created() {
    await this.getInformation()
    await this.getQuestions()
  },
  mounted() {},
  methods: {
    alertNetworkError: common.alertNetworkError,
    getDateStr: common.getDateStr,
    getInformation() {
      // サーバーにアンケートの情報をリクエストする
      if (this.isNewQuestionnaire) {
        this.information = {
          title: '',
          description: '',
          res_shared_to: 'public',
          res_time_limit: this.newTimeLimit,
          respondents: [],
          administrators: [this.getMyTraqId],
          targets: []
        }
      } else {
        return axios
          .get('/questionnaires/' + this.questionnaireId)
          .then(res => {
            this.information = res.data
            if (this.administrates) {
              this.enableEditButton()
            } else {
              this.disableEditButton()
            }
            if (
              this.information.res_time_limit &&
              this.information.res_time_limit !== 'null'
            ) {
              this.noTimeLimit = false
            }
          })
      }
    },
    getQuestions() {
      this.questions = []
      if (!this.isNewQuestionnaire) {
        axios
          .get('/questionnaires/' + this.questionnaireId + '/questions')
          .then(res => {
            this.questions = []
            res.data.forEach(data => {
              this.questions.push(common.convertDataToQuestion(data))
            })
          })
      }
    },
    submitQuestionnaire() {
      if (this.isSubmitting) return // 二重サブミット防止

      const informationData = {
        title: this.information.title,
        description: this.information.description,
        res_time_limit: this.noTimeLimit
          ? undefined
          : moment(
              this.information.res_time_limit,
              'YYYY-MM-DDTHH:mm'
            ).toISOString(),
        res_shared_to: this.information.res_shared_to,
        targets: this.information.targets,
        administrators: this.information.administrators
      }

      this.isSubmitting = true
      if (this.isNewQuestionnaire) {
        // アンケートの新規作成

        axios
          .post('/questionnaires', informationData)
          .then(resp => {
            if (resp.status == 400) {
              throw new Error('Bad Request')
            }
            // 返ってきたquestionnaireIDを保存
            this.newQuestionnaireId = resp.data.questionnaireID
          })
          .then(() => {
            // 質問をサーバーに送信
            return this.sendQuestions(this.newQuestionnaireId, 0)
          })
          .then(() => {
            // 作成したアンケートの個別ページに遷移
            this.showMessage('アンケートを作成しました', 'green')
            router.push('/questionnaires/' + this.newQuestionnaireId)
          })
          .catch(error => {
            // エラーが起きた場合は、送信済みのInformationを削除する
            axios.delete('/questionnaires/' + this.newQuestionnaireId)
            console.log(error)
            this.alertNetworkError()
          })
          .finally(() => {
            this.isSubmitting = false
          })
      } else {
        // 既存のアンケートの編集

        axios
          .patch('/questionnaires/' + this.questionnaireId, informationData)
          .then(resp => {
            if (resp.status == 400) {
              throw new Error('Bad Request')
            }
            // 質問を送信
            return this.sendQuestions(this.questionnaireId, 0)
          })
          .then(() => {
            // 削除された質問それぞれについてDELETEリクエストを送信
            Promise.all(
              this.removedQuestionIds.map(questionId =>
                axios.delete('/questions/' + questionId)
              )
            )
              .then(() => {
                this.removedQuestionIds = []
              })
              .catch(err => {
                console.error(err)
              })
          })
          .then(this.getInformation) // 情報をアップデート
          .then(this.getQuestions) // 質問をアップデート
          .then(() => {
            this.showMessage('アンケートを編集しました', 'green')
            this.disableEditing()
          }) // 編集を終了
          .catch(error => {
            // this.showMessage('通信エラー', 'red')
            console.log(error)
            this.alertNetworkError()
          })
          .finally(() => {
            this.isSubmitting = false
          })
      }
    },
    sendQuestions(questionnaireId, index) {
      // questions配列の、index番目以降の質問をサーバーに送信する
      const question = this.questions[index]
      const data = this.createQuestionData(index)

      if (this.isNewQuestion(question)) {
        return axios
          .post('/questionnaires/' + questionnaireId + '/questions', data)
          .then(() => {
            if (index < this.questions.length - 1) {
              // 残りの質問を送信
              return this.sendQuestions(questionnaireId, index + 1)
            }
          })
          .catch(err => {
            this.showMessage('質問の作成に失敗しました', 'red')
            console.log(err.response)
          })
      } else {
        return axios
          .patch('/questions/' + question.questionId, data)
          .then(() => {
            if (index < this.questions.length - 1) {
              // 残りの質問を送信
              return this.sendQuestions(questionnaireId, index + 1)
            }
          })
          .catch(err => {
            this.showMessage('質問の更新に失敗しました', 'red')
            console.log(err.response)
          })
      }
    },
    createQuestionData(index) {
      // 与えられた質問1つ分のデータをサーバーに送るフォーマットのquestionDataにして返す
      const question = this.questions[index]
      let data = {
        questionnaireID: this.isNewQuestionnaire
          ? this.newQuestionnaireId
          : this.questionnaireId,
        question_type: question.type,
        question_num: index,
        page_num: question.pageNum,
        body: question.questionBody,
        is_required: question.isRequired,
        options: [],
        scale_label_left: '',
        scale_label_right: '',
        scale_min: 0,
        scale_max: 0
      }
      switch (question.type) {
        case 'Checkbox':
        case 'MultipleChoice':
          question.options.forEach(option => {
            data.options.push(option.label)
          })
          break
        case 'LinearScale':
          data.scale_label_left = question.scaleLabels.left
          data.scale_label_right = question.scaleLabels.right
          data.scale_min = question.scaleRange.left
          data.scale_max = question.scaleRange.right
          break
      }
      return data
    },
    isNewQuestion(question) {
      return question.questionId < 0
    },
    enableEditButton() {
      this.showEditButton = true
    },
    disableEditButton() {
      this.showEditButton = false
    },
    toggleEditing() {
      if (this.isNewQuestionnaire) return

      if (this.isEditing) this.disableEditing()
      else this.enableEditing()
    },
    enableEditing() {
      if (this.isNewQuestionnaire) return
      router.push({
        name: 'QuestionnaireDetailsEdit',
        params: this.$route.params,
        query: this.$route.query
      })
    },
    disableEditing() {
      if (this.isNewQuestionnaire) return
      router.push({
        name: 'QuestionnaireDetails',
        params: this.$route.params,
        query: this.$route.query
      })
    },
    abortEditing() {
      // TODO: 変更したかどうかを検出
      // const alertMessage = this.isNewQuestionnaire ? 'アンケートを破棄します。よろしいですか？' : '変更を破棄します。よろしいですか？'
      // if (window.confirm(alertMessage)) {
      if (this.isNewQuestionnaire) {
        router.push('/administrates')
      } else {
        this.disableEditing()
        this.getInformation().then(this.getQuestions)
      }
      // }
    },
    setData(name, data) {
      switch (name) {
        case 'questions':
          this.questions = data
          break
        case 'information':
          this.information = data
          break
        case 'noTimeLimit':
          this.noTimeLimit = data
          break
      }
    },
    setQuestionContent(index, label, value) {
      this.questions[index][label] = value
    },
    removeQuestion(index) {
      if (window.confirm('この質問を削除しますか？')) {
        const id = this.questions[index].questionId
        if (id > 0) {
          // サーバーに存在する質問を削除した場合はリストに追加
          this.removedQuestionIds.push(id)
        }
        this.questions.splice(index, 1)
      }
    },
    async showMessage(body, color) {
      console.log(body)
      this.message = {
        showMessage: true,
        color: color,
        body: body
      }
      await new Promise(resolve => setTimeout(resolve, 3000))
      this.resetMessage()
    },
    resetMessage() {
      this.message = {
        showMessage: false
      }
    }
  }
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped></style>
