# Swagger 配置说明

## 环境变量配置

### 生产环境安全设置

在生产环境中，建议使用以下环境变量来控制 Swagger 访问：

```bash
# 启用生产模式
export GIN_MODE=release

# 启用 Swagger（仅在需要时设置）
export ENABLE_SWAGGER=true

# Swagger 认证凭据（推荐使用环境变量而非配置文件）
export SWAGGER_USERNAME=your_username
export SWAGGER_PASSWORD=your_secure_password
```

### 开发环境设置

开发环境中可以使用配置文件中的默认值：

```yaml
swagger:
  username: "admin"
  password: "F5KDq2exI1bTvIcISdWF"
```

## 安全特性

1. **生产环境保护**: 在 `GIN_MODE=release` 且未设置 `ENABLE_SWAGGER=true` 时，Swagger UI 将返回 404
2. **环境变量优先**: 优先使用环境变量中的认证凭据
3. **凭据验证**: 确保认证凭据不为空，否则返回 503 错误
4. **基础认证**: 使用 HTTP Basic Auth 保护 Swagger 文档

## API 响应结构

### 统一响应格式

```go
type APIResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

### 分页响应格式

```go
type PaginationResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
    Meta    PaginationMeta `json:"meta"`
}
```

## 使用建议

1. 在生产环境中，确保设置强密码或禁用 Swagger
2. 定期更新认证凭据
3. 考虑使用 HTTPS 保护 API 文档访问
4. 监控 Swagger 访问日志