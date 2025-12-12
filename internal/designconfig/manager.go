package designconfig

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.xubinbest.com/go-game-server/internal/registry"
	"github.xubinbest.com/go-game-server/internal/utils"
	"github.xubinbest.com/go-game-server/internal/utils/csvparser"

	"go.uber.org/zap"
)

type DesignConfigManager struct {
	registry  registry.Registry
	cache     map[string]any
	mutex     sync.RWMutex
	tables    []Tables
	csvParser *csvparser.CSVParser
}

func NewDesignConfigManager(reg registry.Registry, tables []Tables) *DesignConfigManager {
	return &DesignConfigManager{
		registry:  reg,
		cache:     make(map[string]any),
		tables:    tables,
		csvParser: csvparser.NewCSVParser(),
	}
}

func (dcm *DesignConfigManager) Start() error {
	for _, table := range dcm.tables {
		content, err := dcm.registry.LoadConfig(context.Background(), table.DataId, table.Group)
		if err != nil {
			return fmt.Errorf("failed to load config %s: %v", table.DataId, err)
		}

		sliceType := reflect.SliceOf(table.DataType)
		slicePtr := reflect.New(sliceType)

		if err = dcm.csvParser.UnmarshalString(content, slicePtr.Interface()); err != nil {
			return fmt.Errorf("failed to unmarshal config %s: %v", table.DataId, err)
		}
		dcm.cache[table.TableName] = slicePtr.Elem().Interface()
		dcm.watchDesignConfig(table)
	}
	return nil
}

func (dcm *DesignConfigManager) watchDesignConfig(table Tables) {
	dcm.registry.WatchConfig(table.DataId, table.Group, func(namespace, group, dataId, data string) {
		dcm.mutex.Lock()
		defer dcm.mutex.Unlock()
		sliceType := reflect.SliceOf(table.DataType)
		slicePtr := reflect.New(sliceType)
		if err := dcm.csvParser.UnmarshalString(data, slicePtr.Interface()); err != nil {
			utils.Error("failed to unmarshal config", zap.String("dataId", dataId), zap.Error(err))
			return
		}
		dcm.cache[table.TableName] = slicePtr.Elem().Interface()
	})
}

func (dcm *DesignConfigManager) GetConfig(tableName string) any {
	dcm.mutex.RLock()
	defer dcm.mutex.RUnlock()
	return dcm.cache[tableName]
}
