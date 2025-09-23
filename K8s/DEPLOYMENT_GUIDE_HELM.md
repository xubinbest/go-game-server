# Kubernetes Helm Chart 部署指南

## 1. 前提条件
1. 已安装并配置 Helm 工具（ https://helm.sh/zh/docs/intro/install/ ）
2. 已配置 Kubernetes 集群访问权限
3. 已创建命名空间和 Secret（可复用YAML方式的K8s/namespace.yaml、K8s/secret.yaml）

---

## 2. 目录结构说明
- K8s/Helm/game-monitoring/         # 监控相关Chart（Prometheus、Grafana等）
- K8s/Helm/kafka/                   # Kafka UI等
- K8s/Helm/kafka-cluster/           # Kafka集群
- K8s/Helm/mysql-cluster/           # MySQL集群
- K8s/Helm/nacos-cluster/           # Nacos集群
- K8s/Helm/nfs-subdir-external-provisioner/ # NFS动态存储
- K8s/Helm/redis-cluster/           # Redis集群
- K8s/Project/                      # 业务服务可自定义Helm Chart

---

## 3. 安装/升级/卸载示例

### 1. 安装 NFS Provisioner（如需动态存储）
```bash
cd K8s/Helm/nfs-subdir-external-provisioner
helm install nfs-provisioner . -n game-server
```

### 2. 安装 MySQL 集群
```bash
cd K8s/Helm/mysql-cluster
helm install mysql-cluster . -n game-server
```

### 3. 安装 Redis 集群
```bash
cd K8s/Helm/redis-cluster
helm install redis-cluster . -n game-server
```

### 4. 安装 Nacos 集群
```bash
cd K8s/Helm/nacos-cluster
helm install nacos-cluster . -n game-server
```

### 5. 安装监控组件（Prometheus/Grafana）
```bash
cd K8s/Helm/game-monitoring
helm install game-monitoring . -n game-server
```

### 6. 安装业务服务（如有Helm Chart）
```bash
cd K8s/Project/gateway
helm install gateway . -n game-server
# 其他服务同理
```

### 7. 升级/卸载
升级：
```bash
helm upgrade nacos-cluster . -n game-server
```
卸载：
```bash
helm uninstall nacos-cluster -n game-server
```

---

## 4. values.yaml 配置
- 各Chart目录下均有 `values.yaml`，可根据实际需求自定义参数（如副本数、资源限制、持久化等）。
- 建议先阅读各Chart下的 `README.md` 或注释。

---

## 5. 注意事项
1. Helm部署前需先创建命名空间和Secret。
2. values.yaml 支持自定义参数，详见各Chart目录下说明。
3. 其余注意事项同YAML方案。
4. 推荐使用 `helm list -n game-server` 查看已安装的Release。

---

如需一键部署所有服务，可编写脚本或使用 Helmfile/ArgoCD 等工具。

如有问题，欢迎补充反馈！ 