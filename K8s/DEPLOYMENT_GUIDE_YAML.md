# Kubernetes YAML 文件部署指南

## 目录结构说明
- K8s/namespace.yaml         # 命名空间
- K8s/secret.yaml            # 敏感信息（如数据库密码等）
- K8s/Yaml/nfs-provisioner/  # NFS存储类与provisioner
- K8s/Yaml/mysql/            # MySQL数据库
- K8s/Yaml/redis-cluster/    # Redis集群
- K8s/Yaml/redis/            # Redis单节点（如有）
- K8s/Yaml/nacos/            # Nacos注册中心
- K8s/Yaml/prometheus/       # Prometheus监控
- K8s/Yaml/grafana/          # Grafana可视化
- K8s/Project/gateway/       # Gateway服务
- K8s/Project/user-service/  # User服务
- K8s/Project/social-service/# Social服务
- K8s/Project/leaderboard-service/ # 排行榜服务
- K8s/ingress/               # Ingress配置

---

## 1. 前提条件
1. 已安装并配置 kubectl 命令行工具
2. 已配置 Kubernetes 集群访问权限
3. NFS 服务器（如 192.168.101.10）已就绪，且 /data/nfs 目录已创建
4. /data/nfs 下有 mysql、redis、nacos、prometheus、grafana 子目录

---

## 2. 基础设施组件部署

### 1. 创建命名空间和 Secret
```bash
kubectl apply -f K8s/namespace.yaml
kubectl apply -f K8s/secret.yaml
```

### 2. 部署 NFS 存储类与 Provisioner
```bash
kubectl apply -f K8s/Yaml/nfs-provisioner/nfs-storageclass.yaml
kubectl apply -f K8s/Yaml/nfs-provisioner/rbac.yaml
kubectl apply -f K8s/Yaml/nfs-provisioner/deployment.yaml
```

### 3. 部署 MySQL
```bash
kubectl apply -f K8s/Yaml/mysql/deployment.yaml
kubectl apply -f K8s/Yaml/mysql/service.yaml
kubectl apply -f K8s/Yaml/mysql/pvc.yaml
```

### 4. 部署 Redis（单节点或集群）
#### 单节点
```bash
kubectl apply -f K8s/Yaml/redis/deployment.yaml
kubectl apply -f K8s/Yaml/redis/service.yaml
kubectl apply -f K8s/Yaml/redis/pvc.yaml
```
#### Redis 集群
```bash
kubectl apply -f K8s/Yaml/redis-cluster/redis-configmap.yaml
kubectl apply -f K8s/Yaml/redis-cluster/redis-cluster-statefulset.yaml
kubectl apply -f K8s/Yaml/redis-cluster/redis-cluster-svc.yaml
kubectl apply -f K8s/Yaml/redis-cluster/redis-cluster-headless-svc.yaml
# ⚠️ 初始化集群
kubectl apply -f K8s/Yaml/redis-cluster/redis-cluster-init-job.yaml
```

### 5. 部署 Nacos
```bash
kubectl apply -f K8s/Yaml/nacos/nacos-cluster-conf.yaml
kubectl apply -f K8s/Yaml/nacos/nacos-custom-properties.yaml
kubectl apply -f K8s/Yaml/nacos/statefulset.yaml
kubectl apply -f K8s/Yaml/nacos/service.yaml
kubectl apply -f K8s/Yaml/nacos/nacos-pdb.yaml
```
- ⚠️ 首次部署前，需初始化 MySQL 数据库：
  ```bash
  # 登录MySQL后执行
  source K8s/Yaml/nacos/mysql-schema.sql
  ```
- 验证集群状态：
  ```bash
  kubectl exec -it nacos-0 -n game-server -- curl http://localhost:8848/nacos/v1/ns/operator/cluster/state
  ```

---

## 3. 监控组件部署

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

## 4. 业务服务部署

### 1. Gateway 服务
```bash
kubectl apply -f K8s/Project/gateway/deployment.yaml
kubectl apply -f K8s/Project/gateway/service.yaml
```

### 2. Social Service
```bash
kubectl apply -f K8s/Project/social-service/deployment.yaml
kubectl apply -f K8s/Project/social-service/service.yaml
```

### 3. User Service
```bash
kubectl apply -f K8s/Project/user-service/deployment.yaml
kubectl apply -f K8s/Project/user-service/service.yaml
```

### 4. Leaderboard Service
```bash
kubectl apply -f K8s/Project/leaderboard-service/deployment.yaml
kubectl apply -f K8s/Project/leaderboard-service/service.yaml
```

---

## 5. Ingress 配置
```bash
kubectl apply -f K8s/ingress/gateway-ingress.yaml
kubectl apply -f K8s/ingress/nacos-ingress.yaml
kubectl apply -f K8s/ingress/monitoring-ingress.yaml
kubectl apply -f K8s/ingress/kafka-ui-ingress.yaml
```
- ⚠️ 请确保已安装 Ingress Controller，并配置好域名解析。

---

## 6. 验证与访问

### 1. 检查资源状态
```bash
kubectl get all,pvc,pv -n game-server
```

### 2. 检查服务日志
```bash
kubectl logs -f <pod-name> -n game-server
```

### 3. 通过 Ingress 访问服务（生产推荐）
- Nacos: http://nacos.yourdomain.com/nacos
- Prometheus: http://monitoring.yourdomain.com/prometheus
- Grafana: http://monitoring.yourdomain.com/grafana
- Gateway: http://gateway.yourdomain.com/ (HTTP API)
- Gateway WebSocket: ws://gateway.yourdomain.com/ws
- ⚠️ 请将 yourdomain.com 替换为实际域名

### 4. 通过 Port-forward 访问（开发调试）
```bash
kubectl port-forward svc/mysql 3306:3306 -n game-server
kubectl port-forward svc/redis 6379:6379 -n game-server
kubectl port-forward svc/gateway 8080:8080 -n game-server
```

---

## 7. 注意事项与常见问题
1. Nacos 集群建议至少 3 个节点，节点间需网络互通。
2. 所有服务均部署在 game-server 命名空间下。
3. 敏感信息存储在 Secret 中，secret.yaml 请勿提交到版本控制。
4. Redis 集群需先执行 init-job。
5. 监控组件需额外配置数据源和仪表盘。
6. 持久化数据均存储在 NFS 服务器。
7. 如遇资源未就绪、Pod CrashLoopBackOff 等问题，建议先查看日志和事件：
   ```bash
   kubectl describe pod <pod-name> -n game-server
   kubectl logs <pod-name> -n game-server
   ```

---

如需扩展服务或自定义参数，请参考各目录下的 yaml 文件及注释。 