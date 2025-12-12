package common

import (
	"fmt"
)

// ValidateHandlerDependencies 验证Handler依赖项是否为空
func ValidateHandlerDependencies(
	dbClient interface{},
	cacheClient interface{},
	cacheManager interface{},
	cfg interface{},
	configManager interface{},
) error {
	if dbClient == nil {
		return fmt.Errorf("dbClient cannot be nil")
	}
	if cacheClient == nil {
		return fmt.Errorf("cacheClient cannot be nil")
	}
	if cacheManager == nil {
		return fmt.Errorf("cacheManager cannot be nil")
	}
	if cfg == nil {
		return fmt.Errorf("cfg cannot be nil")
	}
	if configManager == nil {
		return fmt.Errorf("configManager cannot be nil")
	}
	return nil
}
