# 游戏服务器项目 (github.xubinbest.com/go-game-server)

## 📖 项目简介

这是一个基于Go语言开发的**微服务架构游戏服务器项目**，采用现代化的技术栈和云原生设计理念。项目为多人在线游戏提供完整的后端支持，包括用户管理、社交互动、游戏逻辑、排行榜等核心功能。

### 核心特点
- 🏗️ **微服务架构**：6个独立服务，职责清晰，易于扩展
- 🚀 **高性能**：支持万级并发，响应时间控制在100ms以内
- 🔒 **高可用**：99.9%服务可用性，完善的容错和降级机制
- ☁️ **云原生**：完整的Kubernetes部署方案，支持Helm Chart和YAML两种部署方式
- 📦 **配置中心**：支持Nacos配置中心，配置热更新
- 🎮 **游戏功能完整**：用户系统、社交系统、卡牌系统、宠物系统、装备系统等

## ✨ 功能特性

### 核心功能
- 🚀 **高性能微服务架构**：网关、用户、社交、游戏、排行榜、日志等6个服务
- 🔐 **完整的用户认证系统**：JWT认证、用户注册登录、密码加密
- 👥 **社交系统**：好友系统、公会系统、实时聊天（世界频道、私聊）
- 🎮 **游戏核心服务**：游戏逻辑、战斗系统、状态管理
- 📊 **实时排行榜**：支持多种排行榜类型，实时更新
- 💬 **WebSocket实时通信**：支持WebSocket长连接，实时消息推送
- 📝 **日志服务**：统一的日志收集和分析

### 游戏功能
- 🎴 **卡牌系统**：卡牌激活、升级、升星
- 🐾 **宠物系统**：宠物获取、培养、出战/休战
- 🎒 **背包系统**：物品管理、装备系统
- 📅 **月签到系统**：每日签到、累计奖励（使用位图优化存储）
- ⚔️ **装备系统**：装备穿戴、属性管理

### 技术特性
- 🐳 **Kubernetes容器化部署**：完整的K8s部署配置，支持Helm和YAML两种方式
- 🔍 **服务发现与配置中心**：支持Nacos/Etcd服务注册发现和配置管理
- 📝 **完善的日志系统**：使用zap进行结构化日志，支持日志轮转
- 📊 **监控体系**：Prometheus + Grafana监控，服务健康检查
- 🔄 **消息队列**：Kafka消息队列，支持异步消息处理
- 🗄️ **多数据库支持**：MySQL（结构化数据）+ MongoDB（非结构化数据）+ Redis（缓存）
- 🆔 **分布式ID生成**：雪花算法生成唯一ID，支持Redis集群协调
- 📋 **设计配置系统**：从配置中心加载游戏配置表（CSV/XLSX），支持热更新

## 🛠️ 技术栈

### 编程语言
- **Go 1.23.0+**：主要开发语言

### 核心框架
- **Web框架**：gorilla/mux（HTTP路由）
- **WebSocket**：gorilla/websocket（实时通信）
- **RPC框架**：gRPC + Protocol Buffers（服务间通信）
- **ORM**：GORM（MySQL操作）

### 数据库
- **MySQL 8.0+**：用户数据、社交关系、游戏数据等结构化数据
- **MongoDB**：聊天记录、游戏日志等非结构化数据
- **Redis 6.0+**：缓存、会话管理、分布式锁

### 中间件与服务
- **服务发现/配置中心**：Nacos 2.0+ / Etcd
- **消息队列**：Kafka 2.8+ (Sarama)
- **日志框架**：zap（结构化日志）
- **限流**：go.uber.org/ratelimit
- **认证**：golang-jwt/jwt/v5（JWT认证）

### 容器化与编排
- **容器化**：Docker
- **容器编排**：Kubernetes
- **包管理**：Helm Chart

### 监控与运维
- **监控**：Prometheus + Grafana
- **日志收集**：ELK（可选）
- **服务健康检查**：gRPC Health Check

## 项目结构
```
github.xubinbest.com/go-game-server/
├── cmd/                    # 服务入口
│   ├── gateway/           # 网关服务（HTTP + WebSocket）
│   ├── user-service/      # 用户服务（gRPC）
│   ├── social-service/    # 社交服务（gRPC）
│   ├── game-service/      # 游戏服务（gRPC）
│   ├── leaderboard-service/ # 排行榜服务（gRPC）
│   ├── log-service/       # 日志服务（gRPC）
│   ├── match-service/     # 匹配服务（预留）
│   └── test/              # 测试服务
├── internal/              # 内部包
│   ├── auth/             # 认证相关
│   ├── cache/            # 缓存层
│   ├── config/           # 配置管理
│   ├── db/               # 数据库操作
│   │   ├── interfaces/   # 数据库接口定义
│   │   ├── models/       # 数据模型
│   │   ├── gorm/         # MySQL实现（GORM）
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

## 🎯 核心服务

### 1. 网关服务 (Gateway)
**职责**：统一入口、请求路由、负载均衡、WebSocket连接管理

**功能特性**：
- HTTP和WebSocket双协议支持
- 请求路由和消息转发（HTTP/gRPC/WebSocket）
- 客户端认证和授权
- 限流和熔断保护
- gRPC连接池管理
- 聊天消息广播

**配置**：[cmd/gateway/config.yaml](cmd/gateway/config.yaml)

### 2. 用户服务 (User Service)
**职责**：用户管理、游戏数据管理

**功能特性**：
- 用户注册、登录、JWT认证
- 用户信息管理（等级、经验等）
- 背包系统（物品增删改查）
- 装备系统（装备/卸下装备）
- 卡牌系统（激活、升级、升星）
- 宠物系统（获取、培养、出战/休战）
- 月签到系统（每日签到、累计奖励，使用位图优化）

**配置**：[cmd/user-service/config.yaml](cmd/user-service/config.yaml)

### 3. 社交服务 (Social Service)
**职责**：社交关系管理、聊天功能

**功能特性**：
- 好友系统（添加、删除、批量处理好友请求）
- 公会系统（创建、加入、管理、职位管理）
- 聊天功能（世界频道、私聊、历史消息）
- 消息持久化（MongoDB存储）

**配置**：[cmd/social-service/config.yaml](cmd/social-service/config.yaml)

### 4. 游戏服务 (Game Service)
**职责**：游戏核心逻辑、战斗系统

**功能特性**：
- 游戏状态管理
- 战斗计算
- 游戏规则引擎

### 5. 排行榜服务 (Leaderboard Service)
**职责**：排行榜数据管理、分数统计

**功能特性**：
- 实时排行榜更新
- 多种排行榜类型支持
- 分数排序和统计
- Redis缓存优化

**配置**：[cmd/leaderboard-service/config.yaml](cmd/leaderboard-service/config.yaml)

### 6. 日志服务 (Log Service)
**职责**：日志收集、存储和分析

**功能特性**：
- 统一日志收集
- 日志持久化存储
- 日志查询和分析

**配置**：[cmd/log-service/config.yaml](cmd/log-service/config.yaml)

### 7. 匹配服务 (Match Service)
**职责**：玩家匹配、房间管理（预留）

**功能特性**：
- 智能匹配算法
- 房间分配和管理

## 快速开始

### 环境要求

#### 开发环境
- **Go**: 1.23.0 或更高版本
- **Protocol Buffers**: protoc 编译器
- **Make**: 用于构建脚本（可选）

#### 运行环境
- **Docker**: 用于容器化部署
- **Kubernetes**: 1.20+ 集群
- **NFS服务器**: 用于持久化存储（可选）

#### 基础设施
- **MySQL**: 8.0+（结构化数据存储）
- **MongoDB**: 4.0+（非结构化数据存储）
- **Redis**: 6.0+（缓存和会话管理）
- **Nacos**: 2.0+（服务发现和配置中心）
- **Kafka**: 2.8+（消息队列）
- **Prometheus**: 监控指标收集
- **Grafana**: 监控数据可视化

### 本地开发

#### 1. 克隆项目
```bash
git clone [项目地址]
cd github.xubinbest.com/go-game-server
```

#### 2. 安装依赖
```bash
go mod download
```

#### 3. 生成Protocol Buffers文件
```bash
# Windows
./scripts/gen_proto.bat

# Linux/Mac
chmod +x scripts/gen_proto.bat
./scripts/gen_proto.bat
```

#### 4. 配置环境变量
```bash
# 设置配置位置（local表示使用本地配置文件）
export CONFIG_LOCATION=local

# 设置服务注册类型（nacos或etcd）
export REGISTRY_TYPE=nacos

# Nacos配置
export NACOS_NAMESPACE=your_namespace
export NACOS_SERVER=your_nacos_host
export NACOS_PORT=8848
export NACOS_GROUP=DEFAULT_GROUP
export NACOS_TIMEOUT=5000
```

#### 5. 启动服务（本地开发）
```bash
# 启动网关服务
./bin/scrpit/start_gateway.cmd

# 启动用户服务
./bin/scrpit/start_user.cmd

# 启动社交服务
./bin/scrpit/start_social.cmd
```

### Docker构建

#### 构建镜像
```bash
# 构建所有服务
make build-all

# 或单独构建服务
make build-gateway
make build-social
make build-user
make build-leaderboard
make build-log
```

#### 推送镜像
```bash
# 推送所有服务镜像
make push-all

# 或单独推送服务镜像
make push-gateway
make push-social
make push-user
make push-leaderboard
make push-log
```

**注意**：镜像仓库地址在 `Makefile` 中配置，默认使用 `192.168.101.2:5000`

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

## 💻 开发指南

### 代码规范

- **文件行数限制**：
  - 每个Go文件不超过250行
  - 每个文件夹文件数不超过8个
- **代码风格**：
  - 遵循Go官方代码规范
  - 使用 `gofmt` 格式化代码
  - 使用 `golint` 检查代码质量
- **注释规范**：
  - 公共函数和类型必须添加注释
  - 使用Go标准注释格式
- **错误处理**：
  - 使用统一的错误处理机制
  - 错误信息要清晰明确

### Git提交规范

遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
feat: 新功能
fix: 修复bug
docs: 文档更新
style: 代码格式（不影响代码运行的变动）
refactor: 重构（既不是新增功能，也不是修复bug）
test: 测试相关
chore: 构建过程或辅助工具的变动
perf: 性能优化
ci: CI配置文件和脚本的变动
```

### 开发工具

- **协议生成**: [scripts/gen_proto.bat](scripts/gen_proto.bat) - 生成gRPC代码
- **K8s同步**: [scripts/sync_k8s.bat](scripts/sync_k8s.bat) - 同步K8s配置
- **启动脚本**: [bin/scrpit/](bin/scrpit) - 本地开发启动脚本

### 项目架构原则

- **依赖注入**：使用构造函数注入，避免全局变量
- **接口抽象**：数据库操作通过接口抽象，支持多种实现
- **业务逻辑**：业务逻辑在应用层实现，数据库只做CRUD操作
- **配置管理**：配置通过配置中心管理，支持热更新
- **错误处理**：统一的错误处理和日志记录

## 🗄️ 数据库设计

### 多数据库架构

项目采用多数据库架构，根据数据特性选择合适的存储方案：

| 数据库 | 用途 | 存储内容 |
|--------|------|---------|
| **MySQL** | 结构化数据 | 用户数据、社交关系、游戏数据、装备、背包、卡牌、宠物等 |
| **MongoDB** | 非结构化数据 | 聊天记录、游戏日志、用户行为数据等 |
| **Redis** | 缓存和会话 | 会话管理、排行榜、实时数据缓存、分布式锁 |

### 数据模型

#### 用户相关
- **用户模型**: [internal/db/models/user.go](internal/db/models/user.go) - 用户基本信息、等级、经验
- **月签到模型**: [internal/db/models/user.go](internal/db/models/user.go) - 月签到数据（使用位图优化）

#### 社交相关
- **好友模型**: [internal/db/models/friend.go](internal/db/models/friend.go) - 好友关系
- **公会模型**: [internal/db/models/guild.go](internal/db/models/guild.go) - 公会信息、成员关系

#### 游戏相关
- **背包模型**: [internal/db/models/inventory.go](internal/db/models/inventory.go) - 物品、装备
- **卡牌模型**: [internal/db/models/card.go](internal/db/models/card.go) - 卡牌数据
- **宠物模型**: [internal/db/models/pet.go](internal/db/models/pet.go) - 宠物数据

### 设计配置系统

**重要**：物品和装备模板不存储在数据库表中，而是通过 `internal/designconfig/` 包从配置中心（Nacos）加载到内存中。

- **配置表格式**：支持CSV和XLSX格式
- **配置表类型**：物品、装备、卡牌、宠物、等级、月签到等
- **热更新**：支持配置热更新，无需重启服务
- **配置位置**：[data/csv/](data/csv/) 和 [data/xlsx/](data/xlsx/)

### 数据库初始化

```bash
# 初始化游戏数据库
mysql -u root -p < sql/game_db.sql

# 初始化日志数据库
mysql -u root -p < sql/log_db.sql
```

## 📊 监控和日志

### 日志系统

- **日志框架**：使用 `zap` 进行结构化日志
- **日志级别**：DEBUG、INFO、WARN、ERROR、FATAL
- **日志格式**：JSON格式，便于日志分析
- **日志轮转**：支持日志文件轮转，防止磁盘空间不足
- **日志收集**：可集成ELK（Elasticsearch、Logstash、Kibana）进行日志分析

### 监控系统

- **指标收集**：Prometheus收集服务指标
- **可视化**：Grafana展示监控数据
- **健康检查**：gRPC Health Check协议
- **性能监控**：
  - 请求响应时间
  - 请求QPS
  - 错误率
  - 资源使用率（CPU、内存）

### 告警机制

- **服务异常告警**：服务宕机、健康检查失败
- **性能指标告警**：响应时间过长、QPS异常
- **资源使用告警**：CPU、内存使用率过高

### 监控配置

监控组件部署配置位于：
- **Helm Chart**: [K8s/Helm/game-monitoring/](K8s/Helm/game-monitoring/)
- **YAML配置**: [K8s/Yaml/prometheus/](K8s/Yaml/prometheus/) 和 [K8s/Yaml/grafana/](K8s/Yaml/grafana/)

## ❓ 常见问题

### 1. 服务启动失败

**问题**：服务无法启动或立即退出

**解决方案**：
- 检查环境变量是否配置正确（`CONFIG_LOCATION`、`REGISTRY_TYPE`等）
- 确认依赖服务是否正常运行（MySQL、Redis、Nacos等）
- 查看日志文件排查具体错误
- 检查配置文件格式是否正确（YAML格式）

### 2. 性能问题

**问题**：响应时间过长、QPS低

**解决方案**：
- 检查系统资源使用情况（CPU、内存、网络）
- 优化数据库查询（添加索引、优化SQL）
- 调整缓存策略（增加缓存命中率）
- 检查gRPC连接池配置
- 查看是否有慢查询日志

### 3. 连接问题

**问题**：服务间无法通信、连接超时

**解决方案**：
- 检查网络配置（K8s Service、Ingress）
- 验证服务发现配置（Nacos/Etcd连接）
- 确认防火墙设置
- 检查服务注册是否成功
- 验证gRPC端口是否正确

### 4. 配置中心问题

**问题**：无法从Nacos加载配置

**解决方案**：
- 检查Nacos连接配置（host、port、namespace）
- 确认配置文件中配置的dataId和group是否存在
- 检查Nacos服务是否正常运行
- 验证网络连通性

### 5. 数据库连接问题

**问题**：无法连接数据库

**解决方案**：
- 检查数据库连接字符串（DSN）
- 确认数据库用户权限
- 验证网络连通性
- 检查数据库连接池配置

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

## 📚 相关文档

- [Kratos迁移评估报告](docs/kratos_migration_assessment.md) - Kratos框架改造可行性评估
- [K8s部署指南](K8s/DEPLOYMENT_GUIDE.md) - 完整的Kubernetes部署指南
- [YAML部署指南](K8s/DEPLOYMENT_GUIDE_YAML.md) - 使用YAML文件部署
- [Helm部署指南](K8s/DEPLOYMENT_GUIDE_HELM.md) - 使用Helm Chart部署
- [客户端示例](client/README.md) - 客户端使用示例

## 🔄 更新日志

### v1.0.0 (当前版本)
- ✅ 初始版本发布
- ✅ 完整的微服务架构（6个服务）
- ✅ 用户系统（注册、登录、背包、装备、卡牌、宠物、月签到）
- ✅ 社交系统（好友、公会、聊天）
- ✅ 游戏服务和排行榜服务
- ✅ WebSocket实时通信
- ✅ Kubernetes部署支持（Helm + YAML）
- ✅ 多数据库支持（MySQL + MongoDB + Redis）
- ✅ 配置中心集成（Nacos）
- ✅ 监控和日志系统

## 📄 许可证

[添加许可证信息]

## 👥 贡献者

[添加贡献者信息]

## 📮 联系方式

[添加联系方式]

---

**项目状态**：🟢 活跃开发中  
**最后更新**：2025年 