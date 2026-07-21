# API Access Control & Application Management Service (API 访问控制与应用授权服务)

这是一个基于 Go 原生 `net/http` 与 MySQL 数据库构建的高性能 API 访问控制与应用授权服务。该服务专门用于管理客户端 App 的服务标识（App ID）、服务名称、版本号，生成与之对应的安全访问 Token，进行 API 访问鉴权、访问日志审计及用户意见反馈管理。

项目采用现代化的**前后端分离架构**：

1. **Vue 3 现代化管理后台 (`web/art-design-pro`)**：前端基于开源项目 [Art Design Pro](https://github.com/Daymychen/art-design-pro) (Vue 3 + TypeScript + Vite + Element Plus + Pinia) 打造的现代化管理中台（前端源码位于 `web/art-design-pro`，编译产物部署/嵌入于 `public/`）。支持多语言国际化（中/英）、暗黑模式、极简菜单导览及动态权限控制。
2. **Go 高性能 RESTful 后端 (`/admin`, `/api`)**：提供面向后台的纯 JSON 管理接口（支持 JWT/Session 认证）以及面向客户端应用的 Bearer Token 安全校验与受保护资源接口。

---

## 核心特性

1. **零第三方 Web 框架依赖**：路由、中间件与 HTTP 服务管理全部采用 Go 1.22+ 原生 `net/http` 高效实现。
2. **现代前后端分离架构**：
   - 后台管理完全采用基于 [Art Design Pro](https://github.com/Daymychen/art-design-pro) 的 Vue 3 + TypeScript SPA 架构，配合极简优雅的 Element Plus UI 组件库与 Remix Icon 图标库。
   - 客户端接口（`/api`）使用 Bearer Token 安全头部进行高速路由拦截与鉴定。
3. **实时 Token 有效期与软删除控制**：
   - Token 校验关联 `apps` 与 `blacklist` 表实时检查，应用删除采用逻辑删除模式（关联 Token 自动失效并保留审计日志）。
4. **功能模块丰富**：
   - **应用与 Token 管理**：版本注册、Token 签发与撤销、软删除及状态恢复。
   - **Token 黑名单与访问审计**：一键将异常请求 IP/UUID 加入黑名单，记录客户端 IP 归属地。
   - **用户意见反馈（User Feedback）**：客户端提交意见反馈，后台实时查看详情、切换“待处理/已处理”状态与清除反馈。
   - **节假日与万年历数据接口**：提供标准节假日及公历/农历安排数据。
5. **分层解耦设计**：模块化分层架构（DAL 数据访问层 -> Service 业务逻辑层 -> Handler 控制层 -> Middleware 拦截器层）。

---

## 目录结构

```
api-service/
├── config/
│   └── config.go             # 环境变量与系统配置
├── db/
│   ├── mysql.go              # MySQL 连接池初始化
│   ├── schema.sql            # MySQL 数据库 Schema 脚本
│   └── seed.go               # 默认管理员及动态菜单种子数据播种
├── models/
│   ├── admin.go              # 管理员数据模型
│   ├── app.go                # 客户端应用版本数据模型
│   ├── feedback.go           # 用户意见反馈数据模型
│   ├── log.go                # 访问日志数据模型
│   └── token.go              # Token 与黑名单数据模型
├── repository/
│   ├── admin_repository.go   # 管理员及 RBAC 权限仓储
│   ├── app_repository.go     # 应用版本管理仓储（含逻辑删除与自动撤销 Token）
│   ├── feedback_repository.go# 用户意见反馈数据仓储
│   ├── log_repository.go     # 访问日志与 IP 归属地仓储
│   └── token_repository.go   # Token 及黑名单管理仓储
├── service/
│   ├── admin_service.go      # 管理员认证业务逻辑
│   ├── app_service.go        # 应用版本管理业务逻辑
│   ├── feedback_service.go   # 意见反馈业务逻辑
│   ├── log_service.go        # 访问日志业务逻辑
│   └── token_service.go      # Token 签发、校验与撤销逻辑
├── handler/
│   ├── admin/                # 后台管理 RESTful JSON API 控制器
│   ├── api/                  # 客户端开放 RESTful JSON API 控制器
│   ├── response.go           # 统一 JSON 响应包装
│   └── route.go              # Controller 接口与声明式 Router 包装
├── middleware/
│   ├── admin_auth.go         # 后台 Admin 身份校验拦截器
│   ├── auth.go               # 客户端 API Token Bearer 拦截器
│   └── logger.go             # HTTP 请求日志中间件
├── public/                   # 前端编译静态产物目录（SPA HTML/JS/CSS）
├── web/
│   └── art-design-pro/       # Vue 3 现代化前端源码（Vite + TypeScript + Element Plus）
│       ├── src/
│       │   ├── api/          # 接口请求封装
│       │   ├── locales/      # i18n 国际化语言包 (zh / en)
│       │   └── views/        # 页面视图组件 (token/apps, token/feedback, token/logs 等)
│       └── vite.config.ts    # Vite 构建配置
├── main.go                   # 应用启动入口、静态资源嵌入与路由注册
├── go.mod                    # Go 依赖配置
└── go.sum                    # 依赖锁文件
```

---

## 快速开始

### 1. 数据库配置

请确保 MySQL 实例正常运行，创建数据库并将 [schema.sql](db/schema.sql) 导入：

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS api_service DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 导入表结构
mysql -u root -p api_service < db/schema.sql
```

### 2. 编译前端静态资源

根目录的 `public/` 中已默认包含编译好的前端产物。如需修改前端源码 (`web/art-design-pro/`) 并重新打包至 `public/` 目录：

```bash
# 进入前端工程目录
cd web/art-design-pro

# 安装依赖
pnpm install

# 编译打包 (产物将自动输出至根目录 ../../public)
pnpm build

# 切回项目根目录
cd ../..
```

### 3. 运行后端服务

```bash
# 启动 Go 服务 (首次启动将自动填充初始管理员账号及动态菜单权限)
go run main.go
```

服务启动后默认监听 `http://localhost:8080`。

---

## 客户端 API 测试指引

### 1. 访问受保护的服务接口

- **路由**：`GET /api/protected/resource`
- **说明**：在 HTTP 请求头中携带 Bearer Token 进行校验：
  ```bash
  curl -H "Authorization: Bearer <生成的Token>" http://localhost:8080/api/protected/resource
  ```
- **响应示例**：
  ```json
  {
    "success": true,
    "data": {
      "authenticated_app_id": "com.zqluo.CraftCal",
      "authenticated_version": "1.0.0",
      "message": "Access granted to protected resource!",
      "timestamp": "2026-07-21T14:20:00Z"
    }
  }
  ```

### 2. 提交用户意见反馈接口

- **路由**：`POST /api/feedback`
- **请求示例**：
  ```bash
  curl -X POST -H "Authorization: Bearer <生成的Token>" \
       -H "Content-Type: application/json" \
       -d '{"content": "应用体验非常好，建议增加夜间模式", "contact": "user@example.com", "user_uuid": "usr_9981"}' \
       http://localhost:8080/api/feedback
  ```
