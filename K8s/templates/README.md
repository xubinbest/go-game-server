# K8s 配置模板使用指南

## 📋 概述

本目录包含Kubernetes资源配置模板，用于快速创建新的服务配置。

## 📦 可用模板

### 1. Deployment 模板
**文件**: `deployment-template.yaml`

**占位符**:
- `{{SERVICE_NAME}}` - 服务名称
- `{{IMAGE_NAME}}` - 镜像名称
- `{{REPLICAS}}` - 副本数
- `{{GRPC_PORT}}` - gRPC端口
- `{{CPU_REQUEST}}` - CPU请求
- `{{CPU_LIMIT}}` - CPU限制
- `{{MEMORY_REQUEST}}` - 内存请求
- `{{MEMORY_LIMIT}}` - 内存限制

**使用示例**:
```bash
# 复制模板
cp deployment-template.yaml my-service/deployment.yaml

# 使用sed替换（Linux/Mac）
sed -i 's/{{SERVICE_NAME}}/my-service/g' my-service/deployment.yaml
sed -i 's/{{IMAGE_NAME}}/my-service:v1.0.0/g' my-service/deployment.yaml
# ... 其他替换

# 或手动编辑替换所有占位符
```

---

### 2. Service 模板
**文件**: `service-template.yaml`

**占位符**:
- `{{SERVICE_NAME}}` - 服务名称
- `{{GRPC_PORT}}` - gRPC端口

---

### 3. PodDisruptionBudget 模板
**文件**: `pdb-template.yaml`

**占位符**:
- `{{SERVICE_NAME}}` - 服务名称
- `{{MIN_AVAILABLE}}` - 最少可用Pod数

---

### 4. HorizontalPodAutoscaler 模板
**文件**: `hpa-template.yaml`

**占位符**:
- `{{SERVICE_NAME}}` - 服务名称
- `{{MIN_REPLICAS}}` - 最小副本数
- `{{MAX_REPLICAS}}` - 最大副本数
- `{{CPU_TARGET}}` - CPU目标使用率
- `{{MEMORY_TARGET}}` - 内存目标使用率

---

### 5. NetworkPolicy 模板
**文件**: `network-policy-template.yaml`

**占位符**:
- `{{SERVICE_NAME}}` - 服务名称
- `{{GRPC_PORT}}` - gRPC端口

---

## 🚀 快速创建新服务配置

### 方法1: 手动替换

1. 复制模板文件到目标目录
2. 使用文本编辑器替换所有占位符
3. 验证配置

### 方法2: 使用脚本（推荐）

创建脚本 `create-service-config.sh`:

```bash
#!/bin/bash

SERVICE_NAME=$1
IMAGE_NAME=$2
GRPC_PORT=$3
REPLICAS=${4:-3}

# 创建目录
mkdir -p K8s/Project/$SERVICE_NAME

# 创建Deployment
sed -e "s/{{SERVICE_NAME}}/$SERVICE_NAME/g" \
    -e "s/{{IMAGE_NAME}}/$IMAGE_NAME/g" \
    -e "s/{{GRPC_PORT}}/$GRPC_PORT/g" \
    -e "s/{{REPLICAS}}/$REPLICAS/g" \
    -e "s/{{CPU_REQUEST}}/250m/g" \
    -e "s/{{CPU_LIMIT}}/500m/g" \
    -e "s/{{MEMORY_REQUEST}}/512Mi/g" \
    -e "s/{{MEMORY_LIMIT}}/1Gi/g" \
    K8s/templates/deployment-template.yaml > K8s/Project/$SERVICE_NAME/deployment.yaml

# 创建Service
sed -e "s/{{SERVICE_NAME}}/$SERVICE_NAME/g" \
    -e "s/{{GRPC_PORT}}/$GRPC_PORT/g" \
    K8s/templates/service-template.yaml > K8s/Project/$SERVICE_NAME/service.yaml

# 创建PDB
sed -e "s/{{SERVICE_NAME}}/$SERVICE_NAME/g" \
    -e "s/{{MIN_AVAILABLE}}/2/g" \
    K8s/templates/pdb-template.yaml > K8s/Project/$SERVICE_NAME/pdb.yaml

echo "Configuration files created for $SERVICE_NAME"
```

**使用方法**:
```bash
chmod +x create-service-config.sh
./create-service-config.sh my-service my-service:v1.0.0 50057 3
```

---

## 📝 模板最佳实践

所有模板都包含以下最佳实践：

1. ✅ **更新策略**: RollingUpdate with maxUnavailable: 0
2. ✅ **优雅关闭**: terminationGracePeriodSeconds: 30
3. ✅ **资源限制**: 明确的requests和limits
4. ✅ **健康检查**: gRPC健康检查（对于gRPC服务）
5. ✅ **环境变量**: 完整的Nacos配置
6. ✅ **命名空间**: 统一使用game-server

---

## ⚠️ 注意事项

1. **占位符替换**: 确保替换所有占位符，否则配置无效
2. **端口冲突**: 确保GRPC端口不与其他服务冲突
3. **资源限制**: 根据实际需求调整CPU和内存限制
4. **验证配置**: 使用验证脚本检查配置是否正确

---
