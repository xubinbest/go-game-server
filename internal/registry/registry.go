package registry

import (
	"context"
	"errors"
	"os"
	"strconv"
)

var (
	ErrServiceNotFound = errors.New("service not found")
)

type ServiceInstance struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Metadata    map[string]string `json:"metadata"`
	Ip          string            `json:"ip"`
	ServiceHost string            `json:"service_host,omitempty"`
	Port        int               `json:"port"`
	GroupName   string            `json:"group_name,omitempty"`
	ClusterName string            `json:"cluster_name,omitempty"`
}

type RegistryConfig struct {
	Type  string       `json:"type"`
	ETCD  *ETCDConfig  `json:"etcd,omitempty"`
	Nacos *NacosConfig `json:"nacos,omitempty"`
}

type ETCDConfig struct {
	Endpoints   []string `json:"endpoints"`
	DialTimeout uint64   `json:"dial_timeout"`
}

type NacosConfig struct {
	NameSpace string `json:"namespace"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Host      string `json:"host"`
	Port      uint64 `json:"port"`
	GroupName string `json:"group_name"`
	Timeout   uint64 `json:"timeout"`
}

// ConfigChangeCallback 配置变更回调函数
type ConfigChangeCallback func(namespace, group, dataId, data string)

type Registry interface {
	Register(ctx context.Context, instance *ServiceInstance) error
	Deregister(ctx context.Context, instance *ServiceInstance) error
	Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
	Watch(ctx context.Context, serviceName string) (chan []*ServiceInstance, error)
	WatchConfig(dataId, group string, callback ConfigChangeCallback)
	LoadConfig(ctx context.Context, dataId, group string) (string, error)
	Close() error
}

func NewRegistry() (Registry, error) {
	cfg, err := GetRegistryConfig()
	if err != nil {
		return nil, err
	}
	switch cfg.Type {
	case "etcd":
		return NewEtcdRegistry(cfg.ETCD)
	case "nacos":
		return NewNacosRegistry(cfg.Nacos)
	default:
		return nil, errors.New("unsupported registry type")
	}
}

func GetRegistryConfig() (*RegistryConfig, error) {
	registryType := os.Getenv("REGISTRY_TYPE")
	if registryType == "" {
		return nil, errors.New("registry type not set")
	}

	registryConfig := &RegistryConfig{
		Type: registryType,
	}

	switch registryType {
	case "etcd":
		endpoints, err := getEnvString("ETCD_ENDPOINTS")
		if err != nil {
			return nil, err
		}
		dialTimeout, err := getEnvUint64("ETCD_DIAL_TIMEOUT")
		if err != nil {
			return nil, err
		}
		registryConfig.ETCD = &ETCDConfig{
			Endpoints:   []string{endpoints},
			DialTimeout: dialTimeout,
		}
	case "nacos":
		namespace, err := getEnvString("NACOS_NAMESPACE")
		if err != nil {
			return nil, err
		}
		host, err := getEnvString("NACOS_SERVER")
		if err != nil {
			return nil, err
		}
		port, err := getEnvUint64("NACOS_PORT")
		if err != nil {
			return nil, err
		}
		groupName, err := getEnvString("NACOS_GROUP")
		if err != nil {
			return nil, err
		}
		timeout, err := getEnvUint64("NACOS_TIMEOUT")
		if err != nil {
			return nil, errors.New("NACOS_TIMEOUT not set")
		}
		registryConfig.Nacos = &NacosConfig{
			NameSpace: namespace,
			Username:  "",
			Password:  "",
			Host:      host,
			Port:      uint64(port),
			GroupName: groupName,
			Timeout:   uint64(timeout),
		}
	default:
		return nil, errors.New("unsupported registry type")
	}

	return registryConfig, nil
}

func getEnvString(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", errors.New("environment variable " + key + " not set")
	}
	return value, nil
}

func getEnvUint64(key string) (uint64, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return 0, errors.New("environment variable " + key + " not set")
	}
	value, err := strconv.ParseUint(valueStr, 10, 64)
	if err != nil {
		return 0, errors.New("invalid value for environment variable " + key + "valueStr: " + valueStr)
	}
	return value, nil
}
