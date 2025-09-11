# ARCHITECTURE_DECISION_RECORD.md
## 项目名称：Golang 在线题库测试系统
## 文档版本：v1.1

---
## 一、项目背景与目标

本项目旨在开发一套基于 Golang 的在线题库系统，支持选择题、判断题、填空题、加减法题等题型，提供用户注册、登录与答题功能，用户分为：普通用户、老师和管理员。

目标包括：
- 支持多种题型，包括含图、语音、视频题
- 支持题目标签、难度分级机制，便于分类筛选
- 教师可管理题目、组卷，学生答题并查看成绩
- 管理员具有用户和题库的管理权限
- 前后端解耦，RESTful API 设计，支持移动端或 Web 前端调用

---

## 二、系统架构设计

### 2.1 架构概览

```text
[客户端] ⇄ [Gin API 网关] ⇄ [服务层] ⇄ [GORM + MySQL 数据库]
                              ⇄ [JWT Auth]
                              ⇄ [文件服务：本地/OSS]
```

### 2.2 模块划分

| 模块名称     | 描述 |
|--------------|------|
| 用户模块     | 注册、登录、鉴权、角色控制 |
| 题库模块     | 创建题目、题型支持、图片/语音/视频上传 |
| 分类模块     | 标签与难度管理 |
| 组卷模块     | 教师可创建试卷、自定义题目组合 |
| 答题模块     | 随机出题/按卷答题、答题记录、自动评分 |
| 权限模块     | 用户权限验证（JWT + Role） |
| 管理模块     | 用户、题目、试卷管理等 |

---

## 三、技术选型

| 组件         | 技术或方案             | 理由 |
|--------------|-------------------|------|
| Web 框架     | Gin               | 高性能，轻量 REST API 支持 |
| ORM          | GORM              | Golang 主流 ORM，开发效率高 |
| 数据库       | MySQL5.7          | 结构化数据稳定支持 |
| 配置管理     | Viper             | 支持多环境配置 |
| 鉴权         | JWT + Middleware  | 无状态鉴权，适合 REST API |
| 密码加密     | bcrypt            | 安全标准，防止明文泄漏 |
| 文件上传     | 本地存储 or OSS       | 支持图片、语音、视频题目资源 |
| 测试工具     | Postman / Go test | API 测试与单元测试 |

---

## 四、目录结构约定

```bash
online-quiz/
├── api/                  # 路由与请求绑定
├── internal/
│   ├── controller/
│   ├── service/
│   ├── dao/
│   ├── model/
│   │   ├── entity/
│   │   └── request/
│   └── middleware/
├── pkg/                  # 公共库
├── config/               # 配置文件
├── resource/             # 静态资源（如图片、音频、视频）
├── main.go
```

---

## 五、数据模型设计（MySQL）

### 5.1 用户表 `users`

| 字段名       | 类型        | 说明 |
|--------------|-------------|------|
| id           | BIGINT      | 主键 |
| username     | VARCHAR(50) | 用户名，唯一 |
| password     | VARCHAR(255)| 加密后的密码 |
| role         | ENUM        | 'user' | 'teacher' | 'admin' |
| created_at   | DATETIME    | 创建时间 |
| updated_at   | DATETIME    | 更新时间 |

### 5.2 题目表 `questions`

| 字段名       | 类型        | 说明 |
|--------------|-------------|------|
| id           | BIGINT      | 主键 |
| title        | TEXT        | 题干内容 |
| type         | ENUM        | 'choice', 'judge', 'blank', 'math', 'audio', 'video' |
| media_url    | TEXT        | 题目相关音频/视频资源 |
| image_url    | TEXT        | 图片地址（可为空） |
| options      | JSON        | 仅 choice 类型使用 |
| answer       | TEXT        | 正确答案 |
| difficulty   | ENUM        | 'easy', 'medium', 'hard' |
| tags         | JSON        | 标签列表（如：["语文", "识图"]）|
| created_by   | BIGINT      | 外键：users.id |
| created_at   | DATETIME    | 创建时间 |

### 5.3 答题记录表 `answers`

| 字段名       | 类型        | 说明 |
|--------------|-------------|------|
| id           | BIGINT      | 主键 |
| user_id      | BIGINT      | 外键：用户 |
| question_id  | BIGINT      | 外键：题目 |
| user_answer  | TEXT        | 用户提交的答案 |
| is_correct   | BOOLEAN     | 是否正确 |
| answered_at  | DATETIME    | 提交时间 |

### 5.4 试卷表 `papers`

| 字段名       | 类型        | 说明 |
|--------------|-------------|------|
| id           | BIGINT      | 主键 |
| name         | VARCHAR(100)| 试卷名称 |
| description  | TEXT        | 描述信息 |
| created_by   | BIGINT      | 外键：users.id（教师） |
| created_at   | DATETIME    | 创建时间 |

### 5.5 试卷题目关联表 `paper_questions`

| 字段名       | 类型        | 说明 |
|--------------|-------------|------|
| id           | BIGINT      | 主键 |
| paper_id     | BIGINT      | 所属试卷 |
| question_id  | BIGINT      | 所含题目 |

---

## 六、接口设计（新增内容）

### 6.1 题目分类相关

- **GET /api/tags**
  获取所有题目标签

- **GET /api/questions?tags=数学&difficulty=medium**
  按标签和难度筛选题目

---

### 6.2 媒体资源支持

- **POST /api/upload**
  上传图片、语音、视频文件，返回 `media_url` 或 `image_url`

---

### 6.3 组卷系统（教师专属）

- **POST /api/papers**
  创建试卷（填写名称、描述）

- **POST /api/papers/{id}/add_question**
  向试卷中添加题目

- **GET /api/papers**
  查看教师创建的试卷

- **GET /api/papers/{id}/questions**
  查看试卷内题目列表

- **POST /api/paper/{id}/submit**
  学生提交整卷答案

---

## 七、安全与权限设计

- 所有接口使用 JWT 认证中间件
- 路由权限按角色控制：
    - 用户：答题、查看成绩、按试卷答题
    - 教师：管理题库、组卷、发布考试
    - 管理员：全面管理用户、题库、试卷

---

## 八、后续拓展建议

- 实时考试模式（定时 + WebSocket 监控）
- 成绩统计、导出功能
- 视频防作弊（WebRTC 实时画面）
- 标签推荐系统（基于用户行为）

---

## 九、部署建议

- 使用 Docker Compose 管理数据库与服务
- 本地资源支持 Nginx 或对象存储挂载
- 使用 systemd 管理服务进程
- 生产环境配置自动分离（viper + .env）

---

## 十、变更记录

| 版本 | 日期       | 修改内容                      |
|------|------------|-------------------------------|
| v1.0 | 2025-08-06 | 初始版本                      |
| v1.1 | 2025-08-06 | 添加题目标签/难度/组卷/媒体题 |
