package request

type CreateQuestionRequest struct {
	Title       string `json:"title" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Difficulty  int    `json:"difficulty" binding:"required,min=1,max=5"`
	Options     string `json:"options"`
	Answer      string `json:"answer" binding:"required"`
	Explanation string `json:"explanation"`
	Tags        string `json:"tags"`
}

type UpdateQuestionRequest struct {
	Title       string `json:"title"`
	Difficulty  int    `json:"difficulty" binding:"min=1,max=5"`
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
