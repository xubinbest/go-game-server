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
	indexes   map[string]map[int64]any // 索引: tableName -> id -> item (O(1)查询)
	mutex     sync.RWMutex
	tables    []Tables
	csvParser *csvparser.CSVParser
}

func NewDesignConfigManager(reg registry.Registry, tables []Tables) *DesignConfigManager {
	return &DesignConfigManager{
		registry:  reg,
		cache:     make(map[string]any),
		indexes:   make(map[string]map[int64]any),
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

		data := slicePtr.Elem().Interface()
		dcm.cache[table.TableName] = data

		// 构建索引以支持O(1)查询
		dcm.buildIndex(table.TableName, data)

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

		configData := slicePtr.Elem().Interface()
		dcm.cache[table.TableName] = configData

		// 重建索引
		dcm.buildIndex(table.TableName, configData)
	})
}

func (dcm *DesignConfigManager) GetConfig(tableName string) any {
	dcm.mutex.RLock()
	defer dcm.mutex.RUnlock()
	return dcm.cache[tableName]
}

// buildIndex 构建索引以支持O(1)查询
func (dcm *DesignConfigManager) buildIndex(tableName string, data any) {
	// 使用反射遍历切片，构建id->item的映射
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() != reflect.Slice {
		return
	}

	indexMap := make(map[int64]any, dataValue.Len())
	for i := 0; i < dataValue.Len(); i++ {
		item := dataValue.Index(i).Interface()
		itemValue := reflect.ValueOf(item)

		// 尝试获取ID字段
		idField := itemValue.FieldByName("ID")
		if !idField.IsValid() {
			continue
		}

		// 将ID转换为int64
		var id int64
		switch idField.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			id = idField.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			id = int64(idField.Uint())
		default:
			continue
		}

		indexMap[id] = item
	}

	dcm.indexes[tableName] = indexMap
}

// GetByID 根据ID获取配置项 (O(1)查询)
func (dcm *DesignConfigManager) GetByID(tableName string, id int64) (any, bool) {
	dcm.mutex.RLock()
	defer dcm.mutex.RUnlock()

	indexMap, exists := dcm.indexes[tableName]
	if !exists {
		return nil, false
	}

	item, found := indexMap[id]
	return item, found
}

// 便捷方法：根据类型获取配置项
func (dcm *DesignConfigManager) GetItemByID(id int64) (*ItemData, error) {
	item, found := dcm.GetByID("item", id)
	if !found {
		return nil, fmt.Errorf("item not found: %d", id)
	}
	itemData, ok := item.(ItemData)
	if !ok {
		return nil, fmt.Errorf("item type assertion failed for id: %d", id)
	}
	return &itemData, nil
}

func (dcm *DesignConfigManager) GetCardByID(id int64) (*CardData, error) {
	item, found := dcm.GetByID("card", id)
	if !found {
		return nil, fmt.Errorf("card not found: %d", id)
	}
	cardData, ok := item.(CardData)
	if !ok {
		return nil, fmt.Errorf("card type assertion failed for id: %d", id)
	}
	return &cardData, nil
}

func (dcm *DesignConfigManager) GetPetByID(id int64) (*PetData, error) {
	item, found := dcm.GetByID("pet", id)
	if !found {
		return nil, fmt.Errorf("pet not found: %d", id)
	}
	petData, ok := item.(PetData)
	if !ok {
		return nil, fmt.Errorf("pet type assertion failed for id: %d", id)
	}
	return &petData, nil
}

func (dcm *DesignConfigManager) GetEquipmentByID(id int64) (*EquipmentData, error) {
	item, found := dcm.GetByID("equip", id)
	if !found {
		return nil, fmt.Errorf("equipment not found: %d", id)
	}
	equipData, ok := item.(EquipmentData)
	if !ok {
		return nil, fmt.Errorf("equipment type assertion failed for id: %d", id)
	}
	return &equipData, nil
}
