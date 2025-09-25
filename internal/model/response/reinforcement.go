package response

import "time"

// ReinforcementSettingResponse represents reinforcement setting data in API responses
type ReinforcementSettingResponse struct {
	ID            uint                       `json:"id"`
	Name          string                     `json:"name"`
	Description   string                     `json:"description"`
	CreatorID     uint                       `json:"creator_id"`
	CreatorName   string                     `json:"creator_name"`
	Mode          string                     `json:"mode"`
	ScheduleType  string                     `json:"schedule_type"`
	RatioValue    int                        `json:"ratio_value"`
	IntervalValue int                        `json:"interval_value"`
	IsActive      bool                       `json:"is_active"`
	Items         []ReinforcementItemResponse `json:"items"`
	// 前端需要的额外字段
	ItemIds       []uint                     `json:"item_ids"`
	TargetType    string                     `json:"target_type"`
	StudentIds    []uint                     `json:"student_ids"`
	HomeworkIds   []uint                     `json:"homework_ids"`
	CreatedAt     time.Time                  `json:"created_at"`
	UpdatedAt     time.Time                  `json:"updated_at"`
}

// ReinforcementItemResponse represents reinforcement item data
type ReinforcementItemResponse struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	Description   string    `json:"description"`
	ContentURL    string    `json:"content_url"`
	PreviewURL    string    `json:"preview_url"`
	MediaURL      string    `json:"media_url"` // For backward compatibility
	Color         string    `json:"color"`
	Icon          string    `json:"icon"`
	Duration      int       `json:"duration"`
	AnimationType string    `json:"animation_type"`
	GameType      string    `json:"game_type"`
	Difficulty    string    `json:"difficulty"`
	RewardPoints  int       `json:"reward_points"`
	Volume        int       `json:"volume"`
	Tags          string    `json:"tags"` // JSON string of tags array
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ReinforcementLogResponse represents reinforcement trigger log data
type ReinforcementLogResponse struct {
	ID                     uint                          `json:"id"`
	UserID                 uint                          `json:"user_id"`
	UserName               string                        `json:"user_name"`
	ReinforcementSettingID uint                          `json:"reinforcement_setting_id"`
	ReinforcementSetting   *ReinforcementSettingResponse `json:"reinforcement_setting,omitempty"`
	ReinforcementItemID    uint                          `json:"reinforcement_item_id"`
	ReinforcementItem      *ReinforcementItemResponse    `json:"reinforcement_item,omitempty"`
	HomeworkID             *uint                         `json:"homework_id"`
	SessionID              string                        `json:"session_id"`
	TriggerType            string                        `json:"trigger_type"`
	TriggerValue           int                           `json:"trigger_value"`
	ContextData            map[string]interface{}        `json:"context_data"`
	CreatedAt              time.Time                     `json:"created_at"`
}

// ReinforcementStatsResponse represents reinforcement statistics
type ReinforcementStatsResponse struct {
	TotalTriggers        int                           `json:"total_triggers"`
	TriggersByType       map[string]int                `json:"triggers_by_type"`
	TriggersByItem       map[string]int                `json:"triggers_by_item"`
	AverageTriggerValue  float64                       `json:"average_trigger_value"`
	MostUsedItem         *ReinforcementItemResponse    `json:"most_used_item"`
	RecentTriggers       []ReinforcementLogResponse    `json:"recent_triggers"`
	UserStats            []UserReinforcementStats      `json:"user_stats,omitempty"`
}

// UserReinforcementStats represents per-user reinforcement statistics
type UserReinforcementStats struct {
	UserID            uint      `json:"user_id"`
	UserName          string    `json:"user_name"`
	TotalTriggers     int       `json:"total_triggers"`
	LastTriggerDate   *time.Time `json:"last_trigger_date"`
	FavoriteItemType  string    `json:"favorite_item_type"`
	AverageInterval   float64   `json:"average_interval"` // For time-based reinforcements
}

// ReinforcementSettingListResponse represents paginated reinforcement settings list
type ReinforcementSettingListResponse struct {
	Items      []ReinforcementSettingResponse `json:"items"`
	Total      int64                          `json:"total"`
	Page       int                            `json:"page"`
	PageSize   int                            `json:"page_size"`
	TotalPages int                            `json:"total_pages"`
}

// ReinforcementItemListResponse represents paginated reinforcement items list
type ReinforcementItemListResponse struct {
	Items      []ReinforcementItemResponse `json:"items"`
	Total      int64                       `json:"total"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"page_size"`
	TotalPages int                         `json:"total_pages"`
}

// GradeResponse represents grade data
type GradeResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	AgeMin      int       `json:"age_min"`
	AgeMax      int       `json:"age_max"`
	Order       int       `json:"order"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SubjectResponse represents subject data
type SubjectResponse struct {
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	Code        string          `json:"code"`
	Description string          `json:"description"`
	Icon        string          `json:"icon"`
	Color       string          `json:"color"`
	Order       int             `json:"order"`
	IsActive    bool            `json:"is_active"`
	Topics      []TopicResponse `json:"topics,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// TopicResponse represents topic data
type TopicResponse struct {
	ID              uint      `json:"id"`
	SubjectID       uint      `json:"subject_id"`
	SubjectName     string    `json:"subject_name,omitempty"`
	Name            string    `json:"name"`
	Code            string    `json:"code"`
	Description     string    `json:"description"`
	FullDescription string    `json:"full_description"`
	Icon            string    `json:"icon"`
	Color           string    `json:"color"`
	Order           int       `json:"order"`
	IsActive        bool      `json:"is_active"`
	QuestionCount   int       `json:"question_count,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UserPerformanceResponse represents user performance statistics
type UserPerformanceResponse struct {
	UserID              uint      `json:"user_id"`
	Date                string    `json:"date"`
	TodayLearned        int       `json:"today_learned"`        // Questions answered today
	QuestionsAnswered   int       `json:"questions_answered"`   // Total for the date
	QuestionsCorrect    int       `json:"questions_correct"`
	TimeSpent           int       `json:"time_spent"`
	Streak              int       `json:"streak"`               // Current streak days
	WeeklyTarget        int       `json:"weekly_target"`
	HomeworkCompleted   int       `json:"homework_completed"`
	ReinforcementsEarned int      `json:"reinforcements_earned"`
	AccuracyRate        float64   `json:"accuracy_rate"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// VideoUploadResponse represents video upload result
type VideoUploadResponse struct {
	URL      string `json:"url"`
	VideoID  string `json:"video_id"`
	FileName string `json:"file_name"`
	Size     int64  `json:"size"`
}