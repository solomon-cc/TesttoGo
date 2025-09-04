package request

type CreateQuestionRequest struct {
	Title       string `json:"title" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Difficulty  int    `json:"difficulty" binding:"required,min=1,max=5"`
	Grade       string `json:"grade"`     // 年级
	Subject     string `json:"subject"`   // 科目
	Topic       string `json:"topic"`     // 主题
	Options     string `json:"options"`
	Answer      string `json:"answer" binding:"required"`
	Explanation string `json:"explanation"`
	Tags        string `json:"tags"`
}

type UpdateQuestionRequest struct {
	Title       string `json:"title"`
	Difficulty  int    `json:"difficulty" binding:"min=1,max=5"`
	Grade       string `json:"grade"`     // 年级
	Subject     string `json:"subject"`   // 科目
	Topic       string `json:"topic"`     // 主题
	Options     string `json:"options"`
	Answer      string `json:"answer"`
	Explanation string `json:"explanation"`
	Tags        string `json:"tags"`
}

type CreatePaperRequest struct {
	Title       string `json:"title" binding:"required"`
	QuestionIDs []uint `json:"question_ids" binding:"required"`
	Duration    int    `json:"duration" binding:"required,min=1"`
}

type SubmitAnswerRequest struct {
	PaperID    uint   `json:"paper_id" binding:"required"`
	QuestionID uint   `json:"question_id" binding:"required"`
	Answer     string `json:"answer" binding:"required"`
}

// SingleAnswerRequest 单题答题请求
type SingleAnswerRequest struct {
	Answer string `json:"answer" binding:"required"`
}

// RandomQuestionRequest 随机获取题目请求
type RandomQuestionRequest struct {
	Type       string `json:"type"`       // 题目类型过滤
	Difficulty int    `json:"difficulty"` // 难度过滤 (1-5)
	Tags       string `json:"tags"`       // 标签过滤
	Count      int    `json:"count"`      // 题目数量，默认1
}
