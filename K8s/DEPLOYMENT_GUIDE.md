# Kubernetes 部署总览

本项目支持两种 Kubernetes 部署方式：

- [YAML 文件部署指南](DEPLOYMENT_GUIDE_YAML.md)
- [Helm Chart 部署指南](DEPLOYMENT_GUIDE_HELM.md)

请根据实际需求选择合适的部署方式。

---

## 📋 目录结构说明

```
K8s/
├── namespace.yaml              # 命名空间配置
├── secret.yaml                 # 敏感信息（数据库密码等）
├── resource-quota.yaml         # 资源配额限制
├── limit-range.yaml            # 默认资源限制范围
├── Yaml/                       # 各基础组件 YAML 配置
│   ├── mysql/                  # MySQL数据库
│   ├── redis-cluster/          # Redis集群
│   ├── nacos/                  # Nacos注册中心
│   ├── prometheus/             # Prometheus监控
│   └── grafana/               # Grafana可视化
├── Project/                    # 业务服务配置
│   ├── gateway/               # Gateway服务
│   ├── user-service/          # 用户服务
│   ├── social-service/        # 社交服务
│   ├── leaderboard-service/   # 排行榜服务
│   └── log/                   # 日志服务
│       ├── deployment.yaml    # 部署配置
│       ├── service.yaml       # 服务配置
│       ├── pdb.yaml           # Pod中断预算
│       ├── network-policy.yaml # 网络策略
│       └── hpa.yaml           # 水平自动扩缩容
├── templates/                  # 配置模板（用于快速创建新服务）
├── Helm/                      # 各基础组件 Helm Chart
└── ingress/                   # Ingress 配置
```

---

## 🚀 快速开始

### 1. 基础资源部署（必须）

```bash
# 创建命名空间
kubectl apply -f K8s/namespace.yaml

# 创建Secret（包含数据库密码等敏感信息）
kubectl apply -f K8s/secret.yaml

# 创建资源配额和限制范围（推荐）
kubectl apply -f K8s/resource-quota.yaml
kubectl apply -f K8s/limit-range.yaml
```

### 2. 基础设施部署

按照 [YAML 部署指南](DEPLOYMENT_GUIDE_YAML.md) 或 [Helm 部署指南](DEPLOYMENT_GUIDE_HELM.md) 部署基础设施组件。

### 3. 业务服务部署

每个业务服务包含以下配置文件：
- `deployment.yaml` - 部署配置（包含健康检查、资源限制等）
- `service.yaml` - 服务配置
- `pdb.yaml` - Pod中断预算（确保高可用）
- `network-policy.yaml` - 网络策略（网络安全隔离）
- `hpa.yaml` - 水平自动扩缩容（根据负载自动调整副本数）

---

## ✨ 新增功能说明

### 1. 资源管理
- **ResourceQuota**: 限制整个命名空间的资源使用总量
- **LimitRange**: 为未指定资源限制的容器提供默认值

### 2. 高可用保障
- **PodDisruptionBudget (PDB)**: 确保在节点维护时保持最少可用Pod数量
- **RollingUpdate策略**: 零停机更新，确保至少有一个Pod可用

### 3. 网络安全
- **NetworkPolicy**: 实现网络隔离，只允许必要的流量
  - Gateway: 允许来自Ingress的流量，可访问所有后端服务
  - 后端服务: 只允许来自Gateway的流量

### 4. 自动扩缩容
- **HPA (HorizontalPodAutoscaler)**: 根据CPU/内存使用率自动调整副本数
  - 默认配置: CPU 70%, 内存 80%
  - 最小副本: 3, 最大副本: 8-10

### 5. 健康检查
- **gRPC健康检查**: 所有gRPC服务使用 `grpc_health_probe` 进行健康检查
- **超时和失败阈值**: 配置了合理的超时时间和失败重试次数

### 6. 配置模板
- **templates/**: 提供标准化的配置模板，用于快速创建新服务
- 详细使用说明请参考 [templates/README.md](templates/README.md)

---

## 📚 详细文档

- [YAML 文件部署详细流程](DEPLOYMENT_GUIDE_YAML.md) - 使用原生YAML文件部署
- [Helm Chart 部署详细流程](DEPLOYMENT_GUIDE_HELM.md) - 使用Helm Chart部署
- [配置模板使用指南](templates/README.md) - 使用模板快速创建新服务

---

## ⚠️ 重要提示

1. **Secret安全**: `secret.yaml` 包含敏感信息，请勿提交到版本控制系统
2. **资源配额**: 部署前确保集群有足够的资源配额
3. **网络策略**: 需要CNI插件支持（如Calico、Cilium）
4. **HPA**: 需要Metrics Server或Prometheus Adapter才能正常工作
5. **健康检查**: 确保所有服务镜像包含 `grpc_health_probe` 工具
