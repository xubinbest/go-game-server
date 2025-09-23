package registry

import (
	"context"
	"errors"
	"fmt"

	"github.xubinbest.com/go-game-server/internal/utils"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"go.uber.org/zap"
)

type nacosRegistry struct {
	groupName    string
	namingClient naming_client.INamingClient
	configClient config_client.IConfigClient
}

func NewNacosRegistry(cfg *NacosConfig) (Registry, error) {
	clientConfig := constant.ClientConfig{
		NamespaceId:         cfg.NameSpace,
		TimeoutMs:           cfg.Timeout,
		BeatInterval:        1000,
		NotLoadCacheAtStart: true,
		Username:            cfg.Username,
		Password:            cfg.Password,
		AccessKey:           "",
		SecretKey:           "",
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: cfg.Host,
			Port:   cfg.Port,
		},
	}

	nacosClientParam := vo.NacosClientParam{
		ClientConfig:  &clientConfig,
		ServerConfigs: serverConfigs,
	}

	namingClient, err := clients.NewNamingClient(nacosClientParam)

	if err != nil {
		panic(fmt.Sprintf("failed to create Nacos naming client: %v", err))
	}

	configClient, err := clients.NewConfigClient(nacosClientParam)

	if err != nil {
		panic(fmt.Sprintf("failed to create Nacos config client: %v", err))
	}

	return &nacosRegistry{
		groupName:    cfg.GroupName,
		namingClient: namingClient,
		configClient: configClient,
	}, nil
}

func (n *nacosRegistry) Close() error {
	n.namingClient.CloseClient()
	n.configClient.CloseClient()
	return nil
}

// LoadConfig 加载应用配置（原有功能）
func (n *nacosRegistry) LoadConfig(ctx context.Context, dataId, group string) (string, error) {
	content, err := n.configClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})

	if err != nil {
		utils.Error("Failed to load config from Nacos", zap.Error(err))
		return "", err
	}
	return content, nil
}

func (n *nacosRegistry) Register(ctx context.Context, instance *ServiceInstance) error {
	_, err := n.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          instance.Ip,
		Port:        uint64(instance.Port),
		ServiceName: instance.Name,
		GroupName:   n.groupName,
		ClusterName: instance.ClusterName,
		Metadata:    instance.Metadata,
		Ephemeral:   true,
		Healthy:     true,
		Enable:      true,
		Weight:      1.0,
	})
	return err
}

func (n *nacosRegistry) Deregister(ctx context.Context, instance *ServiceInstance) error {
	_, err := n.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          instance.Ip,
		Port:        uint64(instance.Port), // Nacos默认端口
		ServiceName: instance.Name,
		GroupName:   n.groupName,
		Ephemeral:   true,
	})
	return err
}

func (n *nacosRegistry) Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	// 判断serviceName是否以-service结尾,如果不是，则添加-service后缀
	if len(serviceName) < 8 || serviceName[len(serviceName)-8:] != "-service" {
		serviceName = serviceName + "-service"
	}

	instances, err := n.namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		GroupName:   n.groupName,
		HealthyOnly: true,
	})

	if err != nil {
		return nil, err
	}

	var result []*ServiceInstance
	for _, instance := range instances {
		result = append(result, &ServiceInstance{
			ID:       fmt.Sprintf("%s:%d", instance.Ip, instance.Port),
			Name:     instance.ServiceName,
			Metadata: instance.Metadata,
			Ip:       instance.Ip,
			Port:     int(instance.Port),
		})
	}

	if len(result) == 0 {
		return nil, errors.New("service not found")
	}

	return result, nil
}

func (n *nacosRegistry) Watch(ctx context.Context, serviceName string) (chan []*ServiceInstance, error) {
	watchChan := make(chan []*ServiceInstance, 10)

	err := n.namingClient.Subscribe(&vo.SubscribeParam{
		ServiceName: serviceName,
		GroupName:   n.groupName,
		SubscribeCallback: func(services []model.Instance, err error) {
			// 处理错误
			if err != nil {
				return
			}
			var instances []*ServiceInstance
			for _, s := range services {
				instances = append(instances, &ServiceInstance{
					ID:       fmt.Sprintf("%s:%d", s.Ip, s.Port),
					Name:     s.ServiceName,
					Metadata: s.Metadata,
					Ip:       s.Ip,
					Port:     int(s.Port),
				})
			}
			if len(instances) > 0 {
				watchChan <- instances
			}
		},
	})

	return watchChan, err
}

func (n *nacosRegistry) WatchConfig(dataId, group string, callback ConfigChangeCallback) {
	n.configClient.ListenConfig(vo.ConfigParam{
		DataId:   dataId,
		Group:    group,
		OnChange: callback,
	})
}
