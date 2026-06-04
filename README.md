# API Access Control Service (API 访问控制与应用授权服务)

这是一个基于 Go 原生 `net/http` 和 MySQL 数据库构建的高性能 API 访问控制与应用授权服务。该服务专门用于管理客户端 App 的服务 ID（App ID）、服务名称、版本号，生成与之对应的安全访问 Token，并在 API 访问时进行身份验证。

该项目实现了**前后台架构的分离**：
1. **后端管理后台 (`/admin`)**：面向管理员的控制中心，使用**服务器端渲染（SSR）**技术、**HTML Form 提交表单（无 AJAX）**以及 **Session Cookie** 管理身份状态。支持菜单折叠，其折叠状态在浏览器中自动持久化。
2. **前端客户端服务 (`/api`)**：提供给外部客户端调用的服务接口，通过在 HTTP Header 中携带 **`Authorization: Bearer <token>`** 进行身份验证。

借助于关联的 `App ID` 和 `Version` 状态，该服务能够实现对 Token 有效期的**实时控制**（例如动态禁用某 App 版本，使其对应已生成的 Token 立即失效）。

---

## 核心特性

1. **零第三方 Web 框架依赖**：路由、中间件、渲染和 Server 管理全部采用 Go 1.22+ 原生 `net/http` 实现。
2. **前后架构分离与安全隔离**：
   - 后台管理（`/admin`）使用 Session Cookie 状态管理，表单采用 HTML Form 原生提交以消除 AJAX 依赖，并配合 SVG 图像验证码进行安全登录。
   - 客户端接口（`/api`）使用 Bearer Token 安全头部，两者互不干扰。
3. **声明式路由与彻底解耦**：采用结构体实现 `Controller` 接口的模式（`InitRoutes`），并利用 Go 1.22 的模式匹配功能（`"GET /path"`），彻底消除了业务方法中冗余的 Method 判断，实现了细粒度中间件注入与路由组装的高度解耦。
4. **实时 Token 有效期控制**：Token 验证采用数据库关联查询方式，实时检查 Token 本身及其关联 App 版本的激活状态。
5. **分层架构设计**：项目采用模块化的分层设计（DAL 数据访问层 -> Service 业务层 -> Handler 控制层 -> Middleware 拦截器层），便于扩展与维护。
6. **精美交互体验**：提供玻璃拟态（Glassmorphism）暗色风格的控制后台，支持点击折叠菜单栏（状态持久化，且无首屏布局抖动）。

---

## 目录结构

```
api-service/
├── config/
│   └── config.go             # 环境变量与配置管理
├── db/
│   ├── mysql.go              # MySQL 连接池初始化
│   ├── schema.sql            # MySQL 数据库 Schema 脚本
│   └── seed.go               # 默认管理员种子数据填充
├── models/
│   ├── admin.go              # 管理员及 Session 状态数据模型
│   ├── app.go                # 客户端应用版本数据模型
│   └── token.go              # Token 与 Token 关联明细数据模型
├── repository/
│   ├── admin_repository.go   # 管理员及 Session 相关的数据库操作
│   ├── app_repository.go     # 针对 apps 表的数据库操作
│   └── token_repository.go   # 针对 tokens 表的数据库操作
├── service/
│   ├── admin_service.go      # 管理员登录及账号管理业务逻辑
│   ├── app_service.go        # 应用及版本管理的业务逻辑
│   └── token_service.go      # Token 的生成、实时校验与撤销逻辑
├── handler/
│   ├── admin/                # 针对控制台的所有后端 Handler 实现（按领域切分：app, auth, log 等）
│   ├── api/                  # 针对外部客户端的纯 JSON API Handler
│   ├── response.go           # 统一 JSON 响应格式辅助方法
│   └── route.go              # 控制器接口抽象（Controller）与声明式 Router 实现
├── middleware/
│   ├── admin_auth.go         # 针对 /admin 的 Cookie Session 校验与重定向拦截器
│   ├── auth.go               # 针对 /api 的 Token Bearer 安全校验中间件
│   └── logger.go             # 请求及耗时信息记录中间件
├── public/                   # 静态资源目录
│   ├── css/
│   │   └── style.css         # 系统样式表（玻璃拟态、暗黑主题、侧边栏折叠等）
│   └── js/
│       └── dashboard.js      # 侧边栏折叠交互、Modal 控制与复制辅助（无 AJAX）
├── web/                      # HTML 模板目录
│   ├── layouts/
│   │   └── master.html       # 主页面布局框架及侧边栏
│   └── views/
│       ├── login.html        # 管理员登录视图（直连 SVG Captcha，自带防刷新）
│       ├── dashboard.html    # 应用管理控制中心视图
│       ├── users.html        # 管理员账号管理视图
│       └── tokens.html       # 应用版本 Token 签发与撤销视图
├── main.go                   # 应用启动入口及路由注册
├── go.mod                    # Go 模块文件
└── go.sum                    # 依赖锁文件
```

---

## 快速开始

### 1. 数据库配置

请确保您的 MySQL 实例正在运行，创建数据库并将 [schema.sql](file:///e:/go_workplaces/api-service/db/schema.sql) 导入：

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS api_service DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 导入表结构
mysql -u root -p api_service < db/schema.sql
```

### 2. 运行服务

```bash
# 运行项目 (首次启动将自动在 admin_users 中填充初始账号：admin / admin123)
go run main.go
```

服务启动后将默认在 `http://localhost:8080` 进行监听。

---

## 后台管理操作流程

1. **登录控制台**：
   - 访问 `http://localhost:8080/admin`，若未登录，将自动跳转至 `/admin/login`。
   - 输入初始管理员账号 `admin` / `admin123` 及验证码进行登录。
   - 验证码直接输出为 SVG，点击验证码图片会追加随机参数进行刷新。验证码字符串保存在 Session Cookie 所指向的服务器内存中，表单整体通过 HTML Form POST 提交，无任何 AJAX。
2. **管理侧边栏**：
   - 登录进入控制中心后，点击侧边栏顶部的折叠按钮即可折叠侧边栏为仅显示图标。状态会被持久化保存在 `localStorage` 中。
3. **注册新应用版本**：
   - 点击“注册新应用”按钮，输入 `App ID`、`应用名称`、`版本号` 和 `TTL`，点击“确认注册”将通过 HTML Form POST 提交并在成功后重定向刷新。
4. **签发与撤销 Token**：
   - 在应用行中点击“生成 Token”来为此版本签发一个新的访问密钥，生成成功后会弹出模版框提示复制（密钥仅显示一次）。
   - 点击“管理 Token”进入 `/admin/tokens` 子页面，可查看该版本已发放的所有 Token，点击“撤销”按钮可通过 Form POST 立即注销该 Token 的访问权限。

---

## 客户端服务 API 测试指引

### 1. 访问受保护的服务接口
* **路由**：`GET /api/protected/resource`
* **说明**：通过在请求头中携带从后台获取的 Token 进行验证，返回解析出的 `app_id` 与 `version`。
* **测试命令**：
  ```bash
  curl -H "Authorization: Bearer <生成的Token>" http://localhost:8080/api/protected/resource
  ```
* **期望响应**：
  ```json
  {
    "success": true,
    "data": {
      "authenticated_app_id": "weather_app",
      "authenticated_version": "1.0.0",
      "message": "Access granted to protected resource!",
      "timestamp": "2026-06-03T18:15:00Z"
    }
  }
  ```

### 2. 实时拦截测试：后台禁用应用版本
- 在控制台的主页中，将对应应用行的“实时激活状态”开关关闭（点击时会自动提交表单完成刷新）。
- 再次尝试携带原 Token 请求客户端接口，即使该 Token 仍处于有效期内，也会立即被拦截：
  ```bash
  curl -H "Authorization: Bearer <生成的Token>" http://localhost:8080/api/protected/resource
  ```
- **期望响应**：
  ```json
  {
    "success": false,
    "error": "Associated application version is inactive"
  }
  ```

### 3. 实时拦截测试：后台撤销 Token
- 在后台点击“管理 Token”进入列表，选择刚才的 Token 点击“撤销”确认提交。
- 再次使用该 Token 访问受保护接口：
  ```bash
  curl -H "Authorization: Bearer <生成的Token>" http://localhost:8080/api/protected/resource
  ```
- **期望响应**：
  ```json
  {
    "success": false,
    "error": "Token has been revoked"
  }
  ```

---

## 实时控制的原理实现

在用户访问受保护的客户端 API 时，[auth.go](file:///e:/go_workplaces/api-service/middleware/auth.go) 中间件会拦截并执行以下流程：

1. 提取 Header 中的 Token，在 MySQL 中进行 **JOIN** 查询获取如下信息：
   - Token 撤销状态 (`is_revoked`)
   - Token 生存状态 (`expires_at`)
   - 对应 App 版本的启用状态 (`is_active`)
2. 任何一个判断项不符合要求，均立刻阻断并返回 `401 Unauthorized`。
3. 校验通过后，将 App 身份信息注入 `context` 传导至后续控制器。
