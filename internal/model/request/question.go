package request

type CreateQuestionRequest struct {
	Title       string `json:"title" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Difficulty  int    `json:"difficulty" binding:"required,min=1,max=5"`
	Grade       string `json:"grade"`       // 年级
	Subject     string `json:"subject"`     // 科目
	Topic       string `json:"topic"`       // 主题
	Options     string `json:"options"`
	Answer      string `json:"answer" binding:"required"`
	Explanation string `json:"explanation"`
	MediaURLs   string `json:"media_urls"`   // JSON格式存储多个媒体资源URL
	LayoutType  string `json:"layout_type"`  // 布局类型
	ElementData string `json:"element_data"` // JSON格式存储元素信息
	Tags        string `json:"tags"`
}

type UpdateQuestionRequest struct {
	Title       string `json:"title"`
	Difficulty  int    `json:"difficulty" binding:"min=1,max=5"`
	Grade       string `json:"grade"`       // 年级
	Subject     string `json:"subject"`     // 科目
	Topic       string `json:"topic"`       // 主题
	Options     string `json:"options"`
	Answer      string `json:"answer"`
	Explanation string `json:"explanation"`
	MediaURLs   string `json:"media_urls"`   // JSON格式存储多个媒体资源URL
	LayoutType  string `json:"layout_type"`  // 布局类型
	ElementData string `json:"element_data"` // JSON格式存储元素信息
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

// ImportQuestionData 导入题目数据
type ImportQuestionData struct {
	Title        string   `json:"title" binding:"required"`
	Type         string   `json:"type" binding:"required"`
	Difficulty   int      `json:"difficulty"`
	Grade        string   `json:"grade"`
	Subject      string   `json:"subject"`
	Topic        string   `json:"topic"`
	Options      []string `json:"options"`
	Answer       string   `json:"answer"`
	Explanation  string   `json:"explanation"`
	MediaURLs    []string `json:"media_urls"`
	LayoutType   string   `json:"layout_type"`
	ElementData  string   `json:"element_data"`
	Tags         string   `json:"tags"`
	Status       string   `json:"status"`        // pending, approved, rejected
	ErrorMessage string   `json:"error_message"` // 错误信息
}

// ConfirmImportRequest 确认导入请求
type ConfirmImportRequest struct {
	FileID    string               `json:"file_id" binding:"required"`
	Questions []ImportQuestionData `json:"questions" binding:"required"`
}
