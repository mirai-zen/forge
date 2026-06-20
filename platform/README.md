# Platform Service

平台管理服务，负责管理项目、服务、环境的 CRUD 以及与 GitHub、K8s 的集成。

## 功能特性

- ✅ **项目管理**：创建/查询项目，自动在 GitHub 创建仓库
- ✅ **服务管理**：创建/查询服务，自动生成代码并提交 PR
- ✅ **环境管理**：自动初始化 dev/staging/prod 三个环境
- ✅ **部署触发**：调用 GitHub Actions 触发部署流程
- ✅ **状态查询**：查询服务在各环境的部署状态

## 技术栈

- **框架**: go-zero
- **数据库**: MySQL 8.0+
- **ORM**: GORM
- **Git**: GitHub API
- **构建**: Docker + Kind + ArgoCD

## 快速开始

### 1. 初始化数据库

```bash
# 连接到 MySQL
docker exec -i mysql mysql -uroot -proot123

# 执行初始化脚本
cat deploy/database/init.sql | docker exec -i mysql mysql
```

### 2. 配置服务

编辑 `configs/platform.yaml`：

```yaml
Name: platform-service
Host: 0.0.0.0
Port: 8880

MySQL:
  DataSource: root:root123@tcp(localhost:3306)/forge_platform?charset=utf8mb4&parseTime=True&loc=Local

GitHub:
  Token: your_github_token_here  # 替换为你的 GitHub Personal Access Token
  Org: mirai-zen
```

### 3. 启动服务

```bash
# 方式 1: 直接运行
cd platform
go run cmd/main.go -f configs/platform.yaml

# 方式 2: 使用启动脚本
chmod +x scripts/start.sh
./scripts/start.sh

# 方式 3: 编译后运行
go build -o bin/platform-service ./cmd/main.go
./bin/platform-service -f configs/platform.yaml
```

服务将在 `http://localhost:8880` 启动。

## API 接口

### 项目相关

#### 创建项目
```bash
POST /api/platform/projects/
Content-Type: application/json

{
  "name": "my-project",
  "gitOrg": "mirai-zen",
  "description": "My new project",
  "template": "go-zero-service"
}
```

#### 查询项目列表
```bash
GET /api/platform/projects/?keyword=my-project&page=1&page_size=20
```

#### 查询项目详情
```bash
GET /api/platform/projects/:id
```

### 服务相关

#### 创建服务
```bash
POST /api/platform/projects/:id/services
Content-Type: application/json

{
  "name": "user-service",
  "template": "go-zero-service",
  "params": "{\"port\": \"8081\"}"
}
```

#### 查询服务详情
```bash
GET /api/platform/services/:id
```

#### 部署服务
```bash
POST /api/platform/services/:id/deploy
Content-Type: application/json

{
  "env": "dev",
  "branch": "main"
}
```

#### 查询环境状态
```bash
GET /api/platform/services/:id/envs/:env
```

### 模板相关

#### 查询可用模板
```bash
GET /api/platform/templates/
```

## 项目结构

```
platform/
├── cmd/
│   └── main.go              # 入口文件
├── configs/
│   ├── config.example.yaml  # 配置模板
│   └── platform.yaml        # 本地配置
├── internal/
│   ├── config/              # 配置结构体
│   ├── handler/             # HTTP 处理器注册
│   ├── logic/               # 业务逻辑
│   ├── model/               # 数据模型
│   └── svc/                 # 服务上下文
├── scripts/
│   └── start.sh             # 启动脚本
├── go.mod
└── go.sum
```

## 开发指南

### 添加新接口

1. 在 `internal/handler/routes.go` 中添加路由
2. 在 `internal/logic/` 中创建对应的 Handler 结构体
3. 在 `internal/model/` 中定义数据模型（如需要）

### 数据库迁移

```bash
# 使用 GORM AutoMigrate
cd platform
go run cmd/main.go  # 会在启动时自动迁移表结构
```

### 测试

```bash
# 单元测试
go test ./internal/...

# API 测试
curl http://localhost:8880/api/platform/projects/
```

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `PLATFORM_CONFIG` | 配置文件路径 | `configs/platform.yaml` |
| `GITHUB_TOKEN` | GitHub Token（覆盖配置文件） | - |
| `MYSQL_DSN` | 数据库连接串（覆盖配置文件） | - |

## 部署

### Docker

```bash
# 构建镜像
docker build -t forge-platform:latest .

# 运行容器
docker run -d \
  --name platform \
  -p 8880:8880 \
  -v $(pwd)/configs/platform.yaml:/app/configs/platform.yaml \
  forge-platform:latest
```

### Kubernetes (ArgoCD)

```bash
# 创建 Namespace
kubectl create namespace forge-platform

# 部署
kubectl apply -f deploy/platform-deployment.yaml

# 查看状态
kubectl get pods -n forge-platform
kubectl logs -f deployment/platform -n forge-platform
```

## 故障排查

### 问题：连接数据库失败

```bash
# 检查 MySQL 是否运行
docker ps | grep mysql

# 检查数据库是否存在
docker exec -i mysql mysqlshow forge_platform
```

### 问题：GitHub API 报错

```bash
# 检查 Token 是否有效
curl -H "Authorization: token YOUR_TOKEN" https://api.github.com

# 检查速率限制
curl -H "Authorization: token YOUR_TOKEN" https://api.github.com/rate_limit
```

## 相关文档

- [Proto 定义](../../forge-proto/platform/platform.proto)
- [架构设计](../docs/forge-architecture.md)
- [Sprint 计划](../docs/sprint-plan.md)
- [开发规范](../docs/conventions.md)
