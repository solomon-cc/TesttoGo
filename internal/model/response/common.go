package response

// APIResponse 统一API响应结构
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// PaginationResponse 分页响应结构
type PaginationResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

// PaginationMeta 分页元数据
type PaginationMeta struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
	HasNext  bool  `json:"has_next"`
}

// IDResponse 创建资源时返回ID的响应
type IDResponse struct {
	ID uint `json:"id"`
}