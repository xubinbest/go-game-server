# 游戏服务器项目 (github.xubinbest.com/go-game-server)

## 项目简介
这是一个基于Go语言开发的微服务架构游戏服务器项目，采用现代化的技术栈和云原生设计理念。项目支持多种游戏服务组件，包括用户服务、社交服务、游戏服务、匹配服务、排行榜服务、日志服务等，为游戏提供完整的后端支持。

### 🎯 项目特色
- **完整的游戏生态**: 涵盖用户管理、社交系统、游戏核心、排行榜等完整功能
- **高性能架构**: 基于gRPC通信，支持WebSocket实时通信，Redis集群缓存
- **云原生部署**: 完整的Kubernetes部署方案，支持Helm Chart和YAML两种方式
- **可观测性**: 集成Prometheus监控、Grafana可视化、结构化日志
- **数据驱动**: 支持CSV/Excel配置数据，灵活的游戏数据管理

## 功能特性

### 🏗️ 架构特性
- 🚀 **高性能微服务架构**: 基于gRPC通信，支持水平扩展
- 🐳 **Kubernetes容器化部署**: 完整的K8s部署方案
- 🔍 **服务发现**: 支持Nacos/Etcd服务注册与发现
- 📊 **监控体系**: Prometheus + Grafana完整监控方案

### 🎮 游戏功能
- 🔐 **用户系统**: 注册、登录、JWT认证、用户信息管理
- 🎒 **背包系统**: 物品管理、装备系统、卡牌收集
- 🐾 **宠物系统**: 宠物收集、升级、出战管理
- 📅 **签到系统**: 月签到、累计奖励机制
- 👥 **社交系统**: 好友系统、公会管理、实时聊天
- 🎯 **匹配系统**: 智能匹配算法、房间管理
- 📊 **排行榜**: 实时排行榜、分数统计
- 💬 **实时通信**: WebSocket支持，消息路由

### 🛠️ 技术特性
- 🗄️ **多数据库支持**: MySQL、MongoDB、Redis集群
- 🔄 **消息队列**: Kafka消息队列，支持异步处理
- 📝 **日志系统**: 结构化日志（zap），支持日志轮转
- 🆔 **分布式ID**: 雪花算法生成唯一ID
- ⚡ **缓存策略**: Redis集群缓存，提升性能
- 🔒 **安全机制**: 限流、认证、数据加密

## 技术栈

### 🔧 核心技术
- **编程语言**: Go 1.23.0
- **通信协议**: gRPC + Protocol Buffers
- **Web框架**: gorilla/mux
- **WebSocket**: gorilla/websocket
- **认证**: JWT (golang-jwt/jwt/v5)

### 🗄️ 数据存储
- **关系数据库**: MySQL 8.0+ (GORM)
- **文档数据库**: MongoDB (mongo-driver)
- **缓存数据库**: Redis 6.0+ (go-redis/v9)
- **消息队列**: Kafka (Sarama)

### 🏗️ 基础设施
- **服务发现**: Nacos 2.0+ / Etcd
- **容器编排**: Kubernetes
- **监控**: Prometheus + Grafana
- **日志**: Zap (结构化日志)
- **限流**: Uber ratelimit
- **ID生成**: 雪花算法

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
- **职责**: 游戏核心逻辑、战斗系统、游戏状态管理
- **特性**: 玩家加入/离开游戏、游戏状态查询、玩家操作处理
- **API**: JoinGame, LeaveGame, GetGameState, PlayerAction

### 匹配服务 (Match Service)
- **职责**: 玩家匹配、房间管理
- **特性**: 智能匹配算法、房间分配
- **状态**: 开发中

### 日志服务 (Log Service)
- **职责**: 日志收集、存储、分析
- **特性**: 结构化日志处理、日志聚合
- **配置**: [cmd/log-service/config.yaml](cmd/log-service/config.yaml)

### 排行榜服务 (Leaderboard Service)
- **职责**: 排行榜数据管理、分数统计
- **特性**: 实时排行榜、分数排序、个人排名查询
- **API**: ReportScore, GetLeaderboard, GetRank
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

1. **克隆项目**
```bash
git clone [项目地址]
cd github.xubinbest.com/go-game-server
```

2. **安装依赖**
```bash
go mod download
```

3. **生成Protocol Buffers文件**
```bash
# Windows
./scripts/gen_proto.bat

# Linux/Mac
protoc --go_out=. --go-grpc_out=. internal/pb/*.proto
```

4. **构建Docker镜像**
```bash
# 构建所有服务
make build-all

# 或单独构建服务
make build-gateway
make build-social
make build-user
make build-leaderboard
make build-log-service
```

5. **推送镜像到仓库**
```bash
# 推送所有服务镜像
make push-all

# 或单独推送服务镜像
make push-gateway
make push-social
make push-user
make push-leaderboard
make push-log-service
```

### 本地开发

1. **启动基础服务**
```bash
# 启动Redis集群
# 启动MySQL
# 启动Nacos
# 启动Kafka
```

2. **运行服务**
```bash
# 启动网关服务
go run cmd/gateway/main.go

# 启动用户服务
go run cmd/user-service/main.go

# 启动社交服务
go run cmd/social-service/main.go

# 启动排行榜服务
go run cmd/leaderboard-service/main.go
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
- **卡牌模型**: [internal/db/models/card.go](internal/db/models/card.go)
- **宠物模型**: [internal/db/models/pet.go](internal/db/models/pet.go)

### 游戏配置数据
- **卡牌配置**: [data/csv/card.csv](data/csv/card.csv)
- **装备配置**: [data/csv/equip.csv](data/csv/equip.csv)
- **物品配置**: [data/csv/item.csv](data/csv/item.csv)
- **宠物配置**: [data/csv/pet.csv](data/csv/pet.csv)
- **等级配置**: [data/csv/level.csv](data/csv/level.csv)

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

## API文档

### 服务接口概览

#### 用户服务 (User Service)
- **认证**: Register, Login
- **背包**: GetInventory, AddItem, RemoveItem, UseItem
- **装备**: GetEquipments, EquipItem, UnequipItem
- **卡牌**: GetUserCards, ActivateCard, UpgradeCard, UpgradeCardStar
- **宠物**: GetUserPets, AddPet, SetPetBattleStatus, AddPetExp
- **签到**: GetMonthlySignInfo, MonthlySign, ClaimMonthlySignReward

#### 社交服务 (Social Service)
- **好友**: GetFriendList, SendFriendRequest, HandleFriendRequest, DeleteFriend
- **公会**: CreateGuild, GetGuildInfo, ApplyToGuild, InviteToGuild, KickGuildMember
- **聊天**: SendChatMessage, GetChatMessages

#### 游戏服务 (Game Service)
- **游戏**: JoinGame, LeaveGame, GetGameState, PlayerAction

#### 排行榜服务 (Leaderboard Service)
- **排行榜**: ReportScore, GetLeaderboard, GetRank

### WebSocket消息格式
```protobuf
message WSMessage {
  string service = 1;  // 服务名称
  string method = 2;   // 方法名称
  bytes payload = 3;   // Protocol Buffers 序列化后的数据
}
```

## 更新日志

### v1.0.0 (当前版本)
- ✅ 完整的微服务架构设计
- ✅ 用户服务：认证、背包、装备、卡牌、宠物、签到系统
- ✅ 社交服务：好友系统、公会管理、实时聊天
- ✅ 游戏服务：基础游戏逻辑
- ✅ 排行榜服务：分数统计、排名查询
- ✅ 日志服务：结构化日志处理
- ✅ Kubernetes部署支持（Helm + YAML）
- ✅ 监控体系：Prometheus + Grafana
- ✅ 多数据库支持：MySQL + MongoDB + Redis
- ✅ 消息队列：Kafka集成
- ✅ 客户端示例代码

### 待开发功能
- 🔄 匹配服务：智能匹配算法
- 🔄 游戏核心：战斗系统、技能系统
- 🔄 更多游戏玩法：副本、PVP等 