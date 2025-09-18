package request

// CreateReinforcementSettingRequest represents the request to create a reinforcement setting
type CreateReinforcementSettingRequest struct {
	Name          string `json:"name" binding:"required,min=1,max=100"`
	Description   string `json:"description"`
	Mode          string `json:"mode" binding:"required,oneof=study test"`
	ScheduleType  string `json:"schedule_type" binding:"required,oneof=VR FR VI FI"`
	RatioValue    int    `json:"ratio_value" binding:"min=1,max=100"`
	IntervalValue int    `json:"interval_value" binding:"min=1,max=1440"` // max 24 hours
	ItemIDs       []uint `json:"item_ids" binding:"required,min=1"`
}

// UpdateReinforcementSettingRequest represents the request to update a reinforcement setting
type UpdateReinforcementSettingRequest struct {
	Name          *string `json:"name,omitempty"`
	Description   *string `json:"description,omitempty"`
	Mode          *string `json:"mode,omitempty"`
	ScheduleType  *string `json:"schedule_type,omitempty"`
	RatioValue    *int    `json:"ratio_value,omitempty"`
	IntervalValue *int    `json:"interval_value,omitempty"`
	ItemIDs       []uint  `json:"item_ids,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
}

// CreateReinforcementItemRequest represents the request to create a reinforcement item
type CreateReinforcementItemRequest struct {
	Name          string   `json:"name" binding:"required,min=1,max=100"`
	Type          string   `json:"type" binding:"required,oneof=video game virtual animation sound flower star badge fireworks"`
	Description   string   `json:"description"`
	ContentURL    string   `json:"content_url"`
	PreviewURL    string   `json:"preview_url"`
	MediaURL      string   `json:"media_url"` // For backward compatibility
	Color         string   `json:"color"`
	Icon          string   `json:"icon"`
	Duration      int      `json:"duration" binding:"min=0,max=600"` // Extended range for videos
	AnimationType string   `json:"animation_type"`
	GameType      string   `json:"game_type"`
	Difficulty    string   `json:"difficulty" binding:"omitempty,oneof=easy medium hard"`
	RewardPoints  int      `json:"reward_points" binding:"min=0,max=1000"`
	Volume        int      `json:"volume" binding:"min=0,max=100"`
	Tags          []string `json:"tags"`
}

// UpdateReinforcementItemRequest represents the request to update a reinforcement item
type UpdateReinforcementItemRequest struct {
	Name          *string   `json:"name,omitempty"`
	Type          *string   `json:"type,omitempty"`
	Description   *string   `json:"description,omitempty"`
	ContentURL    *string   `json:"content_url,omitempty"`
	PreviewURL    *string   `json:"preview_url,omitempty"`
	MediaURL      *string   `json:"media_url,omitempty"`
	Color         *string   `json:"color,omitempty"`
	Icon          *string   `json:"icon,omitempty"`
	Duration      *int      `json:"duration,omitempty"`
	AnimationType *string   `json:"animation_type,omitempty"`
	GameType      *string   `json:"game_type,omitempty"`
	Difficulty    *string   `json:"difficulty,omitempty"`
	RewardPoints  *int      `json:"reward_points,omitempty"`
	Volume        *int      `json:"volume,omitempty"`
	Tags          []string  `json:"tags,omitempty"`
	IsActive      *bool     `json:"is_active,omitempty"`
}

// RecordReinforcementTriggerRequest represents logging a reinforcement trigger
type RecordReinforcementTriggerRequest struct {
	ReinforcementSettingID uint                   `json:"reinforcement_setting_id" binding:"required"`
	ReinforcementItemID    uint                   `json:"reinforcement_item_id" binding:"required"`
	HomeworkID             *uint                  `json:"homework_id,omitempty"`
	SessionID              string                 `json:"session_id" binding:"required"`
	TriggerType            string                 `json:"trigger_type" binding:"required"`
	TriggerValue           int                    `json:"trigger_value"`
	ContextData            map[string]interface{} `json:"context_data"`
}

// ListReinforcementSettingsRequest represents query parameters for listing reinforcement settings
type ListReinforcementSettingsRequest struct {
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=100"`
	CreatorID uint   `form:"creator_id"`
	Mode      string `form:"mode"`
	IsActive  *bool  `form:"is_active"`
}

// ListReinforcementItemsRequest represents query parameters for listing reinforcement items
type ListReinforcementItemsRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Type     string `form:"type"`
	IsActive *bool  `form:"is_active"`
}

// ReinforcementStatsRequest represents query parameters for reinforcement statistics
type ReinforcementStatsRequest struct {
	UserID     uint   `form:"user_id"`
	HomeworkID uint   `form:"homework_id"`
	DateFrom   string `form:"date_from"`
	DateTo     string `form:"date_to"`
}

// CreateGradeRequest represents the request to create a grade
type CreateGradeRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Code        string `json:"code" binding:"required,min=1,max=20"`
	Description string `json:"description"`
	AgeMin      int    `json:"age_min" binding:"required,min=3,max=18"`
	AgeMax      int    `json:"age_max" binding:"required,min=3,max=18"`
	Order       int    `json:"order"`
}

// CreateSubjectRequest represents the request to create a subject
type CreateSubjectRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Code        string `json:"code" binding:"required,min=1,max=20"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	Order       int    `json:"order"`
}

// CreateTopicRequest represents the request to create a topic
type CreateTopicRequest struct {
	SubjectID       uint   `json:"subject_id" binding:"required"`
	Name            string `json:"name" binding:"required,min=1,max=100"`
	Code            string `json:"code" binding:"required,min=1,max=50"`
	Description     string `json:"description"`
	FullDescription string `json:"full_description"`
	Icon            string `json:"icon"`
	Color           string `json:"color"`
	Order           int    `json:"order"`
}