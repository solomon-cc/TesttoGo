package entity

import (
	"time"

	"gorm.io/gorm"
)

type QuestionType string

const (
	TypeChoice       QuestionType = "choice"       // 单选题
	TypeMultiChoice  QuestionType = "multichoice"  // 多选题
	TypeJudge        QuestionType = "judge"        // 判断题
	TypeFillIn       QuestionType = "fillin"       // 填空题
	TypeMath         QuestionType = "math"         // 加减法题
	TypeComparison   QuestionType = "comparison"   // 比较题（xx比xx多/少）
	TypeReasoning    QuestionType = "reasoning"    // 推理题（数字序列等）
	TypeVisual       QuestionType = "visual"       // 纯图片题
	TypeCircleSelect QuestionType = "circleselect" // 圈选题（把一样多的圈起来）
)

type Question struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Title       string         `gorm:"type:text" json:"title"`
	Type        QuestionType   `gorm:"type:varchar(20)" json:"type"`
	Difficulty  int            `gorm:"type:tinyint;default:1" json:"difficulty"` // 1-5
	Grade       string         `gorm:"type:varchar(20)" json:"grade"`            // 年级: grade1, grade2, etc.
	SubjectID   *uint          `json:"subject_id"`                               // 科目外键
	TopicID     *uint          `json:"topic_id"`                                 // 主题外键
	Subject     string         `gorm:"type:varchar(50)" json:"subject"`          // 科目: math, vocabulary, reading, literacy (保持向后兼容)
	Topic       string         `gorm:"type:varchar(100)" json:"topic"`           // 主题: addition, subtraction, etc. (保持向后兼容)
	Options     string         `gorm:"type:text" json:"options"`                 // JSON格式存储选项 - 支持文字和图片混合
	Answer      string         `gorm:"type:text" json:"answer"`
	Explanation string         `gorm:"type:text" json:"explanation"` // 答案解释
	CreatorID   uint           `json:"creator_id"`
	MediaURL    string         `gorm:"type:varchar(255)" json:"media_url"` // 单个媒体资源URL（保留兼容性）
	MediaURLs   string         `gorm:"type:text" json:"media_urls"`        // JSON格式存储多个媒体资源URL
	LayoutType  string         `gorm:"type:varchar(50)" json:"layout_type"` // 布局类型：single, horizontal, vertical, grid
	ElementData string         `gorm:"type:text" json:"element_data"`       // JSON格式存储元素位置和标签信息
	Tags        string         `gorm:"type:varchar(255)" json:"tags"`      // 逗号分隔的标签
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	SubjectRef *Subject `gorm:"foreignKey:SubjectID" json:"subject_ref,omitempty"`
	TopicRef   *Topic   `gorm:"foreignKey:TopicID" json:"topic_ref,omitempty"`
	Creator    User     `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
}

type UserAnswer struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	UserID     uint           `json:"user_id"`
	PaperID    uint           `json:"paper_id"` // 可选，单题答题时为0
	QuestionID uint           `json:"question_id"`
	Answer     string         `gorm:"type:text" json:"answer"`
	IsCorrect  bool           `json:"is_correct"`
	AnswerType string         `gorm:"type:varchar(20);default:'single'" json:"answer_type"` // single|paper
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Question Question `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
	User     User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
