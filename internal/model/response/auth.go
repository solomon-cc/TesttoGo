package response

// AuthResponse 认证相关响应
type AuthResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}