package entity

// QuestionOption 题目选项结构，支持文字和图片
type QuestionOption struct {
	Text     string `json:"text"`      // 选项文字内容
	ImageURL string `json:"image_url"` // 选项图片URL
	Value    string `json:"value"`     // 选项值，用于答案匹配
}

// ComparisonElement 比较题元素结构
type ComparisonElement struct {
	Name     string `json:"name"`      // 元素名称，如"蝴蝶"、"花朵"
	ImageURL string `json:"image_url"` // 元素图片URL
	Count    int    `json:"count"`     // 数量（用于数量比较）
}

// ComparisonData 比较题专用数据结构
type ComparisonData struct {
	Elements      []ComparisonElement `json:"elements"`        // 比较的元素
	QuestionType  string              `json:"question_type"`   // quantity, size, length
	CompareFormat string              `json:"compare_format"`  // 比较格式模板，如 "{0}比{1}多/少"
}

// ContentSegment 内容片段，支持文字、图片、填空
type ContentSegment struct {
	Type    string `json:"type"`              // text, image, blank
	Content string `json:"content"`           // 文字内容或图片URL
	BlankID string `json:"blank_id,omitempty"` // 如果是填空，对应的填空ID
}

// BlankItem 填空项
type BlankItem struct {
	ID          string `json:"id"`                    // 填空ID
	Answer      string `json:"answer"`                // 正确答案
	Placeholder string `json:"placeholder,omitempty"` // 填空提示文字
	Required    bool   `json:"required"`              // 是否必填
}

// SubQuestion 子题结构
type SubQuestion struct {
	ID      string           `json:"id"`      // 子题ID
	Content []ContentSegment `json:"content"` // 题干内容（文字+图片+空格混合）
	Blanks  []BlankItem      `json:"blanks"`  // 填空项列表
}

// ComplexQuestionData 复杂题目数据结构（用于复杂填空题和比较题）
type ComplexQuestionData struct {
	Type         string        `json:"type"`                    // complex
	HasMainImage bool          `json:"has_main_image"`          // 是否有主题目图片
	MainImageURL string        `json:"main_image_url,omitempty"` // 主题目图片URL
	SubQuestions []SubQuestion `json:"sub_questions"`           // 子题列表
}