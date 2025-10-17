# 用户行为日志处理系统

## 概述

本系统实现了完整的用户行为日志处理功能，支持多种类型的用户行为日志记录，包括用户创建、登录、登出、物品操作和货币操作等。

## 架构设计

### 1. 消息队列层 (MQ Layer)
- **Kafka生产者**: 负责发送用户行为日志消息到Kafka
- **Kafka消费者**: log_service中的消费者负责接收并处理日志消息
- **消息格式**: 统一的JSON格式，包含日志类型和具体数据

### 2. 服务层 (Service Layer)
- **LogService**: 核心服务，负责消费Kafka消息并处理不同类型的日志
- **消息路由**: 根据日志类型自动路由到对应的处理函数

### 3. 数据访问层 (Data Access Layer)
- **接口定义**: `UserLogDatabase` 接口定义了所有日志相关的数据库操作
- **GORM实现**: 实现了MySQL数据库的日志存储
- **批量操作**: 支持批量插入日志，提高性能

### 4. 数据模型层 (Model Layer)
- **数据库模型**: 定义了5种用户日志的数据结构
- **自动迁移**: 数据库表结构自动创建和更新

## 支持的日志类型

### 1. 用户创建日志 (UserCreateLog)
```go
type UserCreateLog struct {
    UserName     string    `json:"user_name"`
    Time         time.Time `json:"time"`
    CreateIP     string    `json:"create_ip"`
    CreateDevice string    `json:"create_device"`
}
```

### 2. 用户登录日志 (UserLoginLog)
```go
type UserLoginLog struct {
    UserID      int64     `json:"user_id"`
    UserName    string    `json:"user_name"`
    Time        time.Time `json:"time"`
    LoginIP     string    `json:"login_ip"`
    LoginDevice string    `json:"login_device"`
}
```

### 3. 用户登出日志 (UserLogoutLog)
```go
type UserLogoutLog struct {
    UserID       int64     `json:"user_id"`
    UserName     string    `json:"user_name"`
    Time         time.Time `json:"time"`
    LogoutIP     string    `json:"logout_ip"`
    LogoutDevice string    `json:"logout_device"`
}
```

### 4. 用户物品日志 (UserItemLog)
```go
type UserItemLog struct {
    UserID     int64     `json:"user_id"`
    UserName   string    `json:"user_name"`
    ItemID     int64     `json:"item_id"`
    ItemAmount int64     `json:"item_amount"`
    Opt        int32     `json:"opt"`        // 1: 获取, 2: 消耗
    Time       time.Time `json:"time"`
    ItemIP     string    `json:"item_ip"`
    ItemDevice string    `json:"item_device"`
}
```

### 5. 用户货币日志 (UserMoneyLog)
```go
type UserMoneyLog struct {
    UserID      int64     `json:"user_id"`
    UserName    string    `json:"user_name"`
    Money       int64     `json:"money"`
    MoneyType   int32     `json:"money_type"` // 1: 金币, 2: 钻石
    Opt         int32     `json:"opt"`        // 1: 获取, 2: 消耗
    Time        time.Time `json:"time"`
    MoneyIP     string    `json:"money_ip"`
    MoneyDevice string    `json:"money_device"`
}
```

## 使用方法

### 1. 发送日志消息

```go
// 获取Kafka生产者
producer, err := mq.NewKafkaFactory(&cfg.KafkaConfigs).GetProducer(mq.UserBehavior)
if err != nil {
    log.Fatal("Failed to create kafka producer", err)
}

// 创建用户登录日志
loginLog := &userlog.UserLoginLog{
    UserID:      12345,
    UserName:    "testuser",
    Time:        time.Now(),
    LoginIP:     "192.168.1.100",
    LoginDevice: "Windows 10 Chrome",
}

// 发送日志
err = userlog.SendUserLoginLog(producer, loginLog)
if err != nil {
    log.Error("Failed to send user login log", err)
}
```

### 2. 查询日志

```go
// 查询用户登录日志
logs, err := db.GetUserLoginLogs(ctx, userID, 10, 0) // limit=10, offset=0
if err != nil {
    log.Error("Failed to get user login logs", err)
}

for _, log := range logs {
    fmt.Printf("User %s logged in at %s from %s\n", 
        log.UserName, log.Time.Format("2006-01-02 15:04:05"), log.LogoutIP)
}
```

## 数据库表结构

系统会自动创建以下数据库表：

- `user_create_logs`: 用户创建日志表
- `user_login_logs`: 用户登录日志表
- `user_logout_logs`: 用户登出日志表
- `user_item_logs`: 用户物品日志表
- `user_money_logs`: 用户货币日志表

每个表都包含适当的索引以优化查询性能。

## 配置要求

### Kafka配置
```yaml
kafka_configs:
  user_behavior:
    brokers: ["localhost:9092"]
    topic: "user_behavior_logs"
    group_id: "log_service_group"
```

### 数据库配置
确保MySQL数据库已配置并启用，系统会自动创建所需的表结构。

## 性能优化

1. **批量插入**: 支持批量插入日志，减少数据库连接开销
2. **异步处理**: 使用Kafka实现异步日志处理，不阻塞主业务流程
3. **索引优化**: 为常用查询字段添加索引
4. **连接池**: 使用数据库连接池管理连接

## 扩展性

系统设计具有良好的扩展性：

1. **新增日志类型**: 只需添加新的日志模型和处理函数
2. **多数据库支持**: 接口设计支持MySQL、MongoDB等多种数据库
3. **水平扩展**: Kafka支持多分区，可以实现水平扩展

## 监控和日志

- 所有操作都有详细的日志记录
- 支持错误监控和告警
- 可以通过日志查询用户行为轨迹

## 注意事项

1. **数据一致性**: 使用事务确保数据一致性
2. **错误处理**: 完善的错误处理机制，确保系统稳定性
3. **性能监控**: 建议监控Kafka消费延迟和数据库性能
4. **数据清理**: 建议定期清理历史日志数据，避免数据库过大
