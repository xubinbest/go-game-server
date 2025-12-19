# Kubernetes YAML 文件部署指南

## 📋 目录结构说明

```
K8s/
├── namespace.yaml              # 命名空间
├── secret.yaml                 # 敏感信息（如数据库密码等）
├── resource-quota.yaml         # 资源配额限制
├── limit-range.yaml            # 默认资源限制范围
├── Yaml/                       # 基础设施组件
│   ├── nfs-provisioner/        # NFS存储类与provisioner
│   ├── mysql/                  # MySQL数据库
│   ├── redis-cluster/          # Redis集群
│   ├── nacos/                  # Nacos注册中心
│   ├── prometheus/             # Prometheus监控
│   └── grafana/               # Grafana可视化
├── Project/                    # 业务服务
│   ├── gateway/               # Gateway服务
│   ├── user-service/          # User服务
│   ├── social-service/        # Social服务
│   ├── leaderboard-service/   # 排行榜服务
│   └── log/                   # 日志服务
└── ingress/                    # Ingress配置
```

---

## 1. 前提条件

1. ✅ 已安装并配置 `kubectl` 命令行工具
2. ✅ 已配置 Kubernetes 集群访问权限
3. ✅ NFS 服务器（如 192.168.101.10）已就绪，且 `/data/nfs` 目录已创建
4. ✅ `/data/nfs` 下有 `mysql`、`redis`、`nacos`、`prometheus`、`grafana` 子目录
5. ✅ 集群已安装 Metrics Server（HPA需要）
6. ✅ CNI插件支持NetworkPolicy（如Calico、Cilium）

---

## 2. 基础资源部署（第一步）

### 1. 创建命名空间
```bash
kubectl apply -f K8s/namespace.yaml
```

### 2. 创建Secret
```bash
kubectl apply -f K8s/secret.yaml
```
⚠️ **安全提示**: `secret.yaml` 包含敏感信息，请勿提交到版本控制系统

### 3. 创建资源配额和限制范围（推荐）
```bash
kubectl apply -f K8s/resource-quota.yaml
kubectl apply -f K8s/limit-range.yaml
```

**说明**:
- **ResourceQuota**: 限制整个命名空间的资源使用总量
  - CPU: 请求10核，限制20核
  - 内存: 请求20Gi，限制40Gi
  - Pod数量: 最多50个
- **LimitRange**: 为未指定资源限制的容器提供默认值

---

## 3. 基础设施组件部署

### 1. 部署 NFS 存储类与 Provisioner
```bash
kubectl apply -f K8s/Yaml/nfs-provisioner/nfs-storageclass.yaml
kubectl apply -f K8s/Yaml/nfs-provisioner/rbac.yaml
kubectl apply -f K8s/Yaml/nfs-provisioner/deployment.yaml
```

### 2. 部署 MySQL
```bash
kubectl apply -f K8s/Yaml/mysql/deployment.yaml
kubectl apply -f K8s/Yaml/mysql/service.yaml
kubectl apply -f K8s/Yaml/mysql/pvc.yaml
```

### 3. 部署 Redis 集群
```bash
kubectl apply -f K8s/Yaml/redis-cluster/redis-configmap.yaml
kubectl apply -f K8s/Yaml/redis-cluster/redis-cluster-statefulset.yaml
kubectl apply -f K8s/Yaml/redis-cluster/redis-cluster-svc.yaml
kubectl apply -f K8s/Yaml/redis-cluster/redis-cluster-headless-svc.yaml
# ⚠️ 初始化集群（必须）
kubectl apply -f K8s/Yaml/redis-cluster/redis-cluster-init-job.yaml
```

### 4. 部署 Nacos
```bash
kubectl apply -f K8s/Yaml/nacos/nacos-cluster-conf.yaml
kubectl apply -f K8s/Yaml/nacos/nacos-custom-properties.yaml
kubectl apply -f K8s/Yaml/nacos/statefulset.yaml
kubectl apply -f K8s/Yaml/nacos/service.yaml
kubectl apply -f K8s/Yaml/nacos/nacos-pdb.yaml
```

⚠️ **首次部署前，需初始化 MySQL 数据库**:
```bash
# 登录MySQL后执行
source K8s/Yaml/nacos/mysql-schema.sql
```

验证集群状态：
```bash
kubectl exec -it nacos-0 -n game-server -- curl http://localhost:8848/nacos/v1/ns/operator/cluster/state
```

---

## 4. 监控组件部署

### 1. 部署 Prometheus
```bash
kubectl apply -f K8s/Yaml/prometheus/prometheus-configmap.yaml
kubectl apply -f K8s/Yaml/prometheus/prometheus-deployment.yaml
kubectl apply -f K8s/Yaml/prometheus/prometheus-service.yaml
kubectl apply -f K8s/Yaml/prometheus/prometheus-pvc.yaml
```

### 2. 部署 Grafana
```bash
kubectl apply -f K8s/Yaml/grafana/grafana-deployment.yaml
kubectl apply -f K8s/Yaml/grafana/grafana-service.yaml
kubectl apply -f K8s/Yaml/grafana/grafana-pvc.yaml
```

---

## 5. 业务服务部署

每个业务服务包含以下配置文件，建议按顺序部署：

### 1. Gateway 服务
```bash
# 部署配置
kubectl apply -f K8s/Project/gateway/deployment.yaml
kubectl apply -f K8s/Project/gateway/service.yaml

# Pod中断预算（确保高可用）
kubectl apply -f K8s/Project/gateway/pdb.yaml

# 网络策略（网络安全隔离）
kubectl apply -f K8s/Project/gateway/network-policy.yaml

# 水平自动扩缩容（可选，需要Metrics Server）
kubectl apply -f K8s/Project/gateway/hpa.yaml
```

### 2. User Service
```bash
kubectl apply -f K8s/Project/user-service/deployment.yaml
kubectl apply -f K8s/Project/user-service/service.yaml
kubectl apply -f K8s/Project/user-service/pdb.yaml
kubectl apply -f K8s/Project/user-service/network-policy.yaml
kubectl apply -f K8s/Project/user-service/hpa.yaml
```

### 3. Social Service
```bash
kubectl apply -f K8s/Project/social-service/deployment.yaml
kubectl apply -f K8s/Project/social-service/service.yaml
kubectl apply -f K8s/Project/social-service/pdb.yaml
kubectl apply -f K8s/Project/social-service/network-policy.yaml
kubectl apply -f K8s/Project/social-service/hpa.yaml
```

### 4. Leaderboard Service
```bash
kubectl apply -f K8s/Project/leaderboard-service/deployment.yaml
kubectl apply -f K8s/Project/leaderboard-service/service.yaml
kubectl apply -f K8s/Project/leaderboard-service/pdb.yaml
kubectl apply -f K8s/Project/leaderboard-service/network-policy.yaml
kubectl apply -f K8s/Project/leaderboard-service/hpa.yaml
```

### 5. Log Service
```bash
kubectl apply -f K8s/Project/log/deployment.yaml
kubectl apply -f K8s/Project/log/service.yaml
kubectl apply -f K8s/Project/log/pdb.yaml
kubectl apply -f K8s/Project/log/network-policy.yaml
kubectl apply -f K8s/Project/log/hpa.yaml
```

### 6. Test Service（测试服务，可选）
```bash
kubectl apply -f K8s/Project/test/deployment.yaml
kubectl apply -f K8s/Project/test/service.yaml
# 注意: test-service只有1个副本，不需要PDB、NetworkPolicy和HPA
```

---

## 6. Ingress 配置

```bash
kubectl apply -f K8s/ingress/gateway-ingress.yaml
kubectl apply -f K8s/ingress/nacos-ingress.yaml
kubectl apply -f K8s/ingress/monitoring-ingress.yaml
kubectl apply -f K8s/ingress/kafka-ui-ingress.yaml
```

⚠️ **重要**: 
- 请确保已安装 Ingress Controller
- 请将配置文件中的示例域名替换为实际域名
- 配置好域名DNS解析

---

## 7. 验证与访问

### 1. 检查资源状态
```bash
# 查看所有资源
kubectl get all,pvc,pv -n game-server

# 查看Pod中断预算
kubectl get pdb -n game-server

# 查看网络策略
kubectl get networkpolicies -n game-server

# 查看HPA状态
kubectl get hpa -n game-server

# 查看资源配额
kubectl get resourcequota,limitrange -n game-server
```

### 2. 检查服务日志
```bash
kubectl logs -f <pod-name> -n game-server
```

### 3. 检查服务健康状态
```bash
# 查看Pod状态
kubectl get pods -n game-server

# 查看Pod详细信息
kubectl describe pod <pod-name> -n game-server

# 检查健康检查探针
kubectl get pods -n game-server -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.conditions[?(@.type=="Ready")].status}{"\n"}{end}'
```

### 4. 通过 Ingress 访问服务（生产推荐）
- **Nacos**: `http://nacos.yourdomain.com/nacos`
- **Prometheus**: `http://monitoring.yourdomain.com/prometheus`
- **Grafana**: `http://monitoring.yourdomain.com/grafana`
- **Gateway HTTP**: `http://gateway.yourdomain.com/`
- **Gateway WebSocket**: `ws://ws.gateway.yourdomain.com/ws`

⚠️ 请将 `yourdomain.com` 替换为实际域名

### 5. 通过 Port-forward 访问（开发调试）
```bash
kubectl port-forward svc/mysql 3306:3306 -n game-server
kubectl port-forward svc/redis 6379:6379 -n game-server
kubectl port-forward svc/gateway 8080:8080 -n game-server
```

---

## 8. 配置说明

### 健康检查配置
所有gRPC服务使用 **gRPC健康检查协议**：
- **Liveness Probe**: 检测容器是否存活
- **Readiness Probe**: 检测容器是否就绪接收流量
- **工具**: `grpc_health_probe`（需要包含在镜像中）

### 资源限制
所有服务都配置了资源限制：
- **Requests**: 容器启动时保证的资源
- **Limits**: 容器可使用的最大资源

### 更新策略
所有服务使用 **RollingUpdate** 策略：
- `maxSurge: 1` - 更新时最多新增1个Pod
- `maxUnavailable: 0` - 确保至少有一个Pod可用（零停机更新）

### Pod中断预算（PDB）
确保在节点维护时保持最少可用Pod数量：
- 3副本服务: `minAvailable: 2`
- 保证至少2个Pod可用

### 网络策略（NetworkPolicy）
实现网络隔离：
- **Gateway**: 允许来自Ingress的流量，可访问所有后端服务
- **后端服务**: 只允许来自Gateway的流量
- **数据库访问**: 只允许来自应用服务的流量

### 水平自动扩缩容（HPA）
根据CPU/内存使用率自动调整副本数：
- **CPU目标**: 70%
- **内存目标**: 80%
- **最小副本**: 3
- **最大副本**: 8-10（根据服务不同）

---

## 9. 注意事项与常见问题

### 注意事项
1. ✅ **Nacos集群**: 建议至少3个节点，节点间需网络互通
2. ✅ **所有服务**: 均部署在 `game-server` 命名空间下
3. ✅ **Secret安全**: `secret.yaml` 请勿提交到版本控制
4. ✅ **Redis集群**: 需先执行 `init-job` 初始化
5. ✅ **监控组件**: 需额外配置数据源和仪表盘
6. ✅ **持久化数据**: 均存储在NFS服务器
7. ✅ **资源配额**: 部署前确保集群有足够的资源配额
8. ✅ **网络策略**: 需要CNI插件支持（如Calico、Cilium）
9. ✅ **HPA**: 需要Metrics Server或Prometheus Adapter
10. ✅ **健康检查**: 确保所有服务镜像包含 `grpc_health_probe` 工具

### 常见问题排查

#### Pod CrashLoopBackOff
```bash
# 查看Pod日志
kubectl logs <pod-name> -n game-server

# 查看Pod事件
kubectl describe pod <pod-name> -n game-server

# 检查资源限制
kubectl top pod <pod-name> -n game-server
```

#### 服务无法访问
```bash
# 检查Service和Endpoint
kubectl get svc,endpoints -n game-server

# 检查网络策略
kubectl get networkpolicies -n game-server
kubectl describe networkpolicy <policy-name> -n game-server
```

#### HPA不工作
```bash
# 检查Metrics Server
kubectl get apiservice | grep metrics

# 检查HPA状态
kubectl describe hpa <hpa-name> -n game-server
```

#### 资源配额不足
```bash
# 查看资源配额使用情况
kubectl describe resourcequota game-server-quota -n game-server

# 查看LimitRange
kubectl describe limitrange game-server-limits -n game-server
```

---

## 10. 使用配置模板创建新服务

如果需要创建新服务，可以使用配置模板：

```bash
# 参考模板使用指南
cat K8s/templates/README.md
```

详细说明请参考: [配置模板使用指南](templates/README.md)

---

## 11. 一键部署脚本（可选）

可以创建部署脚本简化部署流程：

```bash
#!/bin/bash
# deploy-all.sh

# 基础资源
kubectl apply -f K8s/namespace.yaml
kubectl apply -f K8s/secret.yaml
kubectl apply -f K8s/resource-quota.yaml
kubectl apply -f K8s/limit-range.yaml

# 基础设施（按顺序）
# ... 添加基础设施部署命令

# 业务服务
for service in gateway user-service social-service leaderboard-service log; do
  kubectl apply -f K8s/Project/$service/deployment.yaml
  kubectl apply -f K8s/Project/$service/service.yaml
  kubectl apply -f K8s/Project/$service/pdb.yaml
  kubectl apply -f K8s/Project/$service/network-policy.yaml
  kubectl apply -f K8s/Project/$service/hpa.yaml
done

# Ingress
kubectl apply -f K8s/ingress/
```