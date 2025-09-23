# 游戏服务器项目 (github.xubinbest.com/go-game-server)

## 项目简介
这是一个基于Go语言开发的微服务架构游戏服务器项目，采用现代化的技术栈和云原生设计理念。项目支持多种游戏服务组件，包括用户服务、社交服务、游戏服务、匹配服务、排行榜服务等，为游戏提供完整的后端支持。

## 功能特性
- 🚀 高性能微服务架构
- 🔐 完整的用户认证系统
- 👥 社交系统支持（好友、公会、聊天）
- 🎮 游戏核心服务
- 🎯 智能匹配系统
- 📊 实时排行榜
- 💬 WebSocket实时通信
- 🐳 Kubernetes容器化部署
- 🔍 Nacos服务发现与配置中心
- 📝 完善的日志系统（zap）
- 📊 Prometheus + Grafana监控
- 🔄 Kafka消息队列
- 🗄️ 多数据库支持（MySQL、MongoDB、Redis）
- 🆔 分布式ID生成（雪花算法）

## 技术栈
- **编程语言**: Go 1.23.0
- **主要框架和工具**:
  - Web框架：gorilla/mux
  - WebSocket：gorilla/websocket
  - 数据库：MySQL, MongoDB, Redis
  - 服务发现/配置中心：Nacos/Etcd
  - 日志：zap
  - 限流：ratelimit
  - RPC：gRPC + Protocol Buffers
  - 消息队列：Kafka (Sarama)
  - 容器编排：Kubernetes
  - 监控：Prometheus + Grafana
  - JWT认证：golang-jwt

## 项目结构
```
github.xubinbest.com/go-game-server/
├── cmd/                    # 服务入口
│   ├── gateway/           # 网关服务
│   ├── user-service/      # 用户服务
│   ├── social-service/    # 社交服务
│   ├── game-service/      # 游戏服务
│   ├── match-service/     # 匹配服务
│   ├── leaderboard-service/ # 排行榜服务
│   └── test/              # 测试服务
├── internal/              # 内部包
│   ├── auth/             # 认证相关
│   ├── cache/            # 缓存层
│   ├── config/           # 配置管理
│   ├── db/               # 数据库操作
│   │   ├── interfaces/   # 数据库接口
│   │   ├── models/       # 数据模型
│   │   ├── mysql/        # MySQL实现
│   │   └── mongodb/      # MongoDB实现
│   ├── designconfig/     # 设计配置
│   ├── gateway/          # 网关实现
│   ├── middleware/       # 中间件
│   ├── pb/               # Protocol Buffers 定义
│   ├── registry/         # 服务注册
│   ├── snowflake/        # ID生成器
│   ├── social/           # 社交功能
│   ├── user/             # 用户相关
│   ├── game_service/     # 游戏服务
│   ├── leaderboard/      # 排行榜服务
│   ├── mq/               # 消息队列
│   └── utils/            # 工具函数
├── dockerfile/           # Docker配置文件
├── K8s/                  # Kubernetes配置
│   ├── DEPLOYMENT_GUIDE.md         # 部署总览
│   ├── DEPLOYMENT_GUIDE_YAML.md    # YAML部署指南
│   ├── DEPLOYMENT_GUIDE_HELM.md    # Helm部署指南
│   ├── namespace.yaml              # 命名空间定义
│   ├── secret.yaml                 # 敏感信息Secret
│   ├── Helm/                       # 各基础组件Helm Chart
│   │   ├── mysql-cluster/         # MySQL集群
│   │   ├── redis-cluster/         # Redis集群
│   │   ├── nacos-cluster/         # Nacos集群
│   │   ├── kafka-cluster/         # Kafka集群
│   │   └── game-monitoring/       # 监控组件
│   ├── Yaml/                       # 各基础组件YAML配置
│   ├── Project/                    # 业务服务K8s配置
│   ├── ingress/                    # Ingress配置
│   └── ...                         # 其他K8s相关文件
├── data/                 # 游戏数据配置
│   ├── csv/             # CSV格式数据
│   └── xlsx/            # Excel格式数据
├── client/               # 客户端示例
├── scripts/              # 脚本文件
├── sql/                  # 数据库脚本
└── examples/             # 示例代码
```

## 核心服务

### 网关服务 (Gateway)
- **职责**: 请求路由、负载均衡、WebSocket连接管理
- **特性**: 消息转发、客户端认证、限流熔断
- **配置**: [cmd/gateway/config.yaml](cmd/gateway/config.yaml)

### 用户服务 (User Service)
- **职责**: 用户认证、注册、登录、装备管理
- **特性**: JWT认证、背包系统、装备系统
- **配置**: [cmd/user-service/config.yaml](cmd/user-service/config.yaml)

### 社交服务 (Social Service)
- **职责**: 好友系统、公会系统、聊天功能
- **特性**: 实时聊天、好友关系、公会管理
- **配置**: [cmd/social-service/config.yaml](cmd/social-service/config.yaml)

### 游戏服务 (Game Service)
- **职责**: 游戏核心逻辑、战斗系统
- **特性**: 游戏状态管理、战斗计算

### 匹配服务 (Match Service)
- **职责**: 玩家匹配、房间管理
- **特性**: 智能匹配算法、房间分配

### 排行榜服务 (Leaderboard Service)
- **职责**: 排行榜数据管理、分数统计
- **特性**: 实时排行榜、分数排序
- **配置**: [cmd/leaderboard-service/config.yaml](cmd/leaderboard-service/config.yaml)

## 快速开始

### 环境要求
- Go 1.23.0 或更高版本
- Docker
- Kubernetes集群
- NFS服务器
- MySQL 8.0+
- Redis 6.0+
- Nacos 2.0+
- Kafka 2.8+

### 构建步骤

1. 克隆项目
```bash
git clone [项目地址]
cd github.xubinbest.com/go-game-server
```

2. 安装依赖
```bash
go mod download
```

3. 生成Protocol Buffers文件
```bash
./scripts/gen_proto.bat
```

4. 构建Docker镜像
```bash
# 构建所有服务
make build-all

# 或单独构建服务
make build-gateway
make build-social
make build-user
make build-leaderboard
```

5. 推送镜像到仓库
```bash
# 推送所有服务镜像
make push-all

# 或单独推送服务镜像
make push-gateway
make push-social
make push-user
make push-leaderboard
```

## 部署说明

详细的部署指南请参考以下文档：

### 部署总览
- [K8s/DEPLOYMENT_GUIDE.md](K8s/DEPLOYMENT_GUIDE.md) - 完整的部署指南

### 部署方式选择
- [K8s/DEPLOYMENT_GUIDE_YAML.md](K8s/DEPLOYMENT_GUIDE_YAML.md) - 使用YAML文件部署
- [K8s/DEPLOYMENT_GUIDE_HELM.md](K8s/DEPLOYMENT_GUIDE_HELM.md) - 使用Helm Chart部署

### 部署步骤

1. **基础组件部署**
   - 命名空间和Secret配置
   - 存储类配置
   - MySQL集群部署
   - Redis集群部署
   - Nacos集群部署
   - Kafka集群部署

2. **监控组件部署**
   - Prometheus部署
   - Grafana部署

3. **业务服务部署**
   - Gateway服务
   - Social服务
   - User服务
   - Leaderboard服务
   - Ingress配置

## 开发指南

### 代码规范
- 遵循Go标准代码规范
- 使用gofmt格式化代码
- 编写单元测试
- 添加必要的注释
- 使用统一的错误处理机制

### 提交规范
```
feat: 新功能
fix: 修复bug
docs: 文档更新
style: 代码格式
refactor: 重构
test: 测试相关
chore: 构建过程或辅助工具的变动
```

### 开发工具
- **协议生成**: [scripts/gen_proto.bat](scripts/gen_proto.bat)
- **K8s同步**: [scripts/sync_k8s.bat](scripts/sync_k8s.bat)
- **启动脚本**: [bin/scrpit/](bin/scrpit)

## 数据库设计

### 多数据库架构
- **MySQL**: 用户数据、社交关系、游戏数据等结构化数据
- **MongoDB**: 聊天记录、游戏日志、非结构化数据
- **Redis**: 会话管理、排行榜、实时数据缓存

### 数据模型
- **用户模型**: [internal/db/models/user.go](internal/db/models/user.go)
- **好友模型**: [internal/db/models/friend.go](internal/db/models/friend.go)
- **公会模型**: [internal/db/models/guild.go](internal/db/models/guild.go)
- **背包模型**: [internal/db/models/inventory.go](internal/db/models/inventory.go)

## 监控和日志

### 日志系统
- 使用zap进行结构化日志
- 支持日志轮转
- 集成ELK日志分析

### 监控系统
- Prometheus指标收集
- Grafana仪表盘展示
- 服务健康检查
- 性能指标监控

### 告警机制
- 服务异常告警
- 性能指标告警
- 资源使用告警

## 常见问题

### 1. 服务启动失败
- 检查配置文件是否正确
- 确认依赖服务是否正常运行
- 查看日志文件排查问题

### 2. 性能问题
- 检查系统资源使用情况
- 优化数据库查询
- 调整缓存策略

### 3. 连接问题
- 检查网络配置
- 验证服务发现配置
- 确认防火墙设置

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交变更 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建Pull Request

## 许可证
[添加许可证信息]

## 联系方式
[添加联系方式]

## 更新日志

### v1.0.0
- 初始版本发布
- 支持基础的游戏服务功能
- 完整的微服务架构
- Kubernetes部署支持 