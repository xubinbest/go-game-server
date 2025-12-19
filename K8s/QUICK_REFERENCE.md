# Kubernetes 配置快速参考

## 📋 配置文件清单

### 基础资源
| 文件 | 说明 | 必需 |
|------|------|------|
| `namespace.yaml` | 命名空间 | ✅ |
| `secret.yaml` | 敏感信息 | ✅ |
| `resource-quota.yaml` | 资源配额 | 推荐 |
| `limit-range.yaml` | 资源限制范围 | 推荐 |

### 业务服务配置（每个服务）
| 文件 | 说明 | 必需 |
|------|------|------|
| `deployment.yaml` | 部署配置 | ✅ |
| `service.yaml` | 服务配置 | ✅ |
| `pdb.yaml` | Pod中断预算 | 推荐 |
| `network-policy.yaml` | 网络策略 | 推荐 |
| `hpa.yaml` | 水平自动扩缩容 | 可选 |

---

## 🚀 快速部署命令

### 1. 基础资源
```bash
kubectl apply -f K8s/namespace.yaml
kubectl apply -f K8s/secret.yaml
kubectl apply -f K8s/resource-quota.yaml
kubectl apply -f K8s/limit-range.yaml
```

### 2. 单个服务完整部署
```bash
SERVICE_NAME="user-service"  # 替换为实际服务名

kubectl apply -f K8s/Project/$SERVICE_NAME/deployment.yaml
kubectl apply -f K8s/Project/$SERVICE_NAME/service.yaml
kubectl apply -f K8s/Project/$SERVICE_NAME/pdb.yaml
kubectl apply -f K8s/Project/$SERVICE_NAME/network-policy.yaml
kubectl apply -f K8s/Project/$SERVICE_NAME/hpa.yaml
```

### 3. 所有业务服务部署
```bash
for service in gateway user-service social-service leaderboard-service log; do
  kubectl apply -f K8s/Project/$service/deployment.yaml
  kubectl apply -f K8s/Project/$service/service.yaml
  kubectl apply -f K8s/Project/$service/pdb.yaml
  kubectl apply -f K8s/Project/$service/network-policy.yaml
  kubectl apply -f K8s/Project/$service/hpa.yaml
done
```

---

## 🔍 常用检查命令

### 查看资源状态
```bash
# 所有资源
kubectl get all -n game-server

# Pod状态
kubectl get pods -n game-server

# 服务状态
kubectl get svc -n game-server

# Pod中断预算
kubectl get pdb -n game-server

# 网络策略
kubectl get networkpolicies -n game-server

# HPA状态
kubectl get hpa -n game-server

# 资源配额
kubectl get resourcequota,limitrange -n game-server
```

### 查看详细信息
```bash
# Pod详细信息
kubectl describe pod <pod-name> -n game-server

# Service详细信息
kubectl describe svc <service-name> -n game-server

# HPA详细信息
kubectl describe hpa <hpa-name> -n game-server

# 资源配额使用情况
kubectl describe resourcequota game-server-quota -n game-server
```

### 查看日志
```bash
# Pod日志
kubectl logs <pod-name> -n game-server

# 实时日志
kubectl logs -f <pod-name> -n game-server

# 所有Pod日志
kubectl logs -l app=<service-name> -n game-server
```

---

## 📊 服务端口映射

| 服务 | gRPC端口 | 说明 |
|------|----------|------|
| gateway | 8080 (HTTP), 8081 (WS) | 网关服务 |
| user-service | 50052 | 用户服务 |
| social-service | 50051 | 社交服务 |
| leaderboard-service | 50055 | 排行榜服务 |
| log-service | 50056 | 日志服务 |
| test-service | 50053 | 测试服务 |

---

## ⚙️ 配置参数说明

### ResourceQuota（资源配额）
- **CPU请求**: 10核
- **CPU限制**: 20核
- **内存请求**: 20Gi
- **内存限制**: 40Gi
- **Pod数量**: 最多50个
- **Service数量**: 最多20个

### LimitRange（资源限制范围）
- **默认CPU**: 500m请求，2核限制
- **默认内存**: 1Gi请求，4Gi限制
- **最小CPU**: 100m
- **最小内存**: 128Mi

### HPA（水平自动扩缩容）
- **CPU目标**: 70%
- **内存目标**: 80%
- **最小副本**: 3
- **最大副本**: 8-10（根据服务不同）

### PDB（Pod中断预算）
- **3副本服务**: `minAvailable: 2`
- **保证**: 至少2个Pod可用

### 更新策略
- **类型**: RollingUpdate
- **maxSurge**: 1
- **maxUnavailable**: 0（零停机更新）

---

## 🔐 网络策略规则

### Gateway
- **Ingress**: 允许来自Ingress Controller和同命名空间的流量
- **Egress**: 允许访问所有后端服务、Nacos、Redis、DNS

### 后端服务（user/social/leaderboard/log）
- **Ingress**: 只允许来自Gateway的流量
- **Egress**: 允许访问数据库、Redis、Nacos、DNS、Kafka（log-service）

---

## 🛠️ 故障排查

### Pod无法启动
```bash
# 1. 查看Pod状态
kubectl get pods -n game-server

# 2. 查看Pod事件
kubectl describe pod <pod-name> -n game-server

# 3. 查看Pod日志
kubectl logs <pod-name> -n game-server

# 4. 检查资源配额
kubectl describe resourcequota game-server-quota -n game-server
```

### 服务无法访问
```bash
# 1. 检查Service和Endpoint
kubectl get svc,endpoints -n game-server

# 2. 检查网络策略
kubectl get networkpolicies -n game-server
kubectl describe networkpolicy <policy-name> -n game-server

# 3. 测试连接
kubectl exec -it <pod-name> -n game-server -- curl <service-name>:<port>
```

### HPA不工作
```bash
# 1. 检查Metrics Server
kubectl get apiservice | grep metrics

# 2. 检查HPA状态
kubectl describe hpa <hpa-name> -n game-server

# 3. 检查Pod资源使用
kubectl top pods -n game-server
```

### 健康检查失败
```bash
# 1. 检查健康检查配置
kubectl describe pod <pod-name> -n game-server | grep -A 10 "Liveness\|Readiness"

# 2. 手动测试健康检查
kubectl exec -it <pod-name> -n game-server -- /bin/grpc_health_probe -addr=:<port>

# 3. 检查镜像是否包含grpc_health_probe
kubectl exec -it <pod-name> -n game-server -- ls -la /bin/grpc_health_probe
```

---

## 📝 常用操作

### 扩缩容
```bash
# 手动扩缩容
kubectl scale deployment <deployment-name> --replicas=5 -n game-server

# 查看HPA自动扩缩容
kubectl get hpa -n game-server -w
```

### 更新镜像
```bash
# 更新部署镜像
kubectl set image deployment/<deployment-name> <container-name>=<new-image> -n game-server

# 查看更新状态
kubectl rollout status deployment/<deployment-name> -n game-server

# 回滚
kubectl rollout undo deployment/<deployment-name> -n game-server
```

### 删除资源
```bash
# 删除单个服务
kubectl delete -f K8s/Project/<service-name>/ -n game-server

# 删除所有业务服务
for service in gateway user-service social-service leaderboard-service log; do
  kubectl delete -f K8s/Project/$service/ -n game-server
done
```

---

## 🔗 相关文档

- [部署总览](DEPLOYMENT_GUIDE.md)
- [YAML部署指南](DEPLOYMENT_GUIDE_YAML.md)
- [Helm部署指南](DEPLOYMENT_GUIDE_HELM.md)
- [配置模板使用指南](templates/README.md)
