# Kubernetes 部署总览

本项目支持两种 Kubernetes 部署方式：

- [YAML 文件部署指南](DEPLOYMENT_GUIDE_YAML.md)
- [Helm Chart 部署指南](DEPLOYMENT_GUIDE_HELM.md)

请根据实际需求选择合适的部署方式。

---

## 目录结构简要说明
- K8s/namespace.yaml         # 命名空间
- K8s/secret.yaml            # 敏感信息
- K8s/Yaml/                  # 各基础组件 YAML 配置
- K8s/Project/               # 业务服务 YAML/Helm 配置
- K8s/Helm/                  # 各基础组件 Helm Chart
- K8s/ingress/               # Ingress 配置

---

## 快速入口
- [YAML 文件部署详细流程](DEPLOYMENT_GUIDE_YAML.md)
- [Helm Chart 部署详细流程](DEPLOYMENT_GUIDE_HELM.md)
