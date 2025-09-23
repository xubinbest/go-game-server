// 策划配置表csv解析器
// 基础数据类型以及自定义类型
// 当前自定义类型只支持json格式转结构体

package csvparser

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// CSVParser 自定义CSV解析器
type CSVParser struct{}

// NewCSVParser 创建新的CSV解析器实例
func NewCSVParser() *CSVParser {
	return &CSVParser{}
}

// UnmarshalString 解析CSV字符串到结构体切片
func (p *CSVParser) UnmarshalString(csvData string, dest interface{}) error {
	// 检查dest是否为指向切片的指针
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to slice")
	}

	// 获取切片元素类型
	sliceType := destValue.Elem().Type()
	elemType := sliceType.Elem()

	// 解析CSV
	reader := csv.NewReader(strings.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %v", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("empty CSV data")
	}

	// 第一行是表头
	headers := records[0]

	// 创建字段映射
	fieldMap, err := p.createFieldMap(elemType, headers)
	if err != nil {
		return fmt.Errorf("failed to create field map: %v", err)
	}

	// 创建结果切片
	resultSlice := reflect.MakeSlice(sliceType, 0, len(records)-1)

	// 解析每一行数据
	for i := 1; i < len(records); i++ {
		record := records[i]

		// 创建新的结构体实例
		elemValue := reflect.New(elemType).Elem()

		// 填充字段值
		err := p.fillStructFields(elemValue, fieldMap, headers, record)
		if err != nil {
			return fmt.Errorf("error parsing row %d: %v", i, err)
		}

		// 添加到结果切片
		resultSlice = reflect.Append(resultSlice, elemValue)
	}

	// 设置结果
	destValue.Elem().Set(resultSlice)
	return nil
}

// createFieldMap 创建字段映射
func (p *CSVParser) createFieldMap(structType reflect.Type, headers []string) (map[string]reflect.StructField, error) {
	fieldMap := make(map[string]reflect.StructField)

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// 跳过未导出的字段
		if !field.IsExported() {
			continue
		}

		csvTag := field.Tag.Get("csv")
		if csvTag == "" {
			csvTag = strings.ToLower(field.Name)
		}

		// 支持忽略字段
		if csvTag == "-" {
			continue
		}

		fieldMap[csvTag] = field
	}

	return fieldMap, nil
}

// fillStructFields 填充结构体字段
func (p *CSVParser) fillStructFields(structValue reflect.Value, fieldMap map[string]reflect.StructField, headers []string, record []string) error {
	for i, header := range headers {
		if i >= len(record) {
			continue
		}

		field, exists := fieldMap[header]
		if !exists {
			// 跳过未知字段
			continue
		}

		fieldValue := structValue.FieldByName(field.Name)
		if !fieldValue.CanSet() {
			continue
		}

		err := p.setFieldValue(fieldValue, record[i], field.Type)
		if err != nil {
			return fmt.Errorf("error setting field %s: %v", field.Name, err)
		}
	}

	return nil
}

// setFieldValue 设置字段值
func (p *CSVParser) setFieldValue(fieldValue reflect.Value, csvValue string, fieldType reflect.Type) error {
	// 处理空值
	if csvValue == "" {
		return nil
	}

	switch fieldType.Kind() {
	case reflect.String:
		fieldValue.SetString(csvValue)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(csvValue, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse int: %v", err)
		}
		fieldValue.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(csvValue, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse uint: %v", err)
		}
		fieldValue.SetUint(uintVal)

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(csvValue, 64)
		if err != nil {
			return fmt.Errorf("failed to parse float: %v", err)
		}
		fieldValue.SetFloat(floatVal)

	case reflect.Bool:
		boolVal, err := strconv.ParseBool(csvValue)
		if err != nil {
			return fmt.Errorf("failed to parse bool: %v", err)
		}
		fieldValue.SetBool(boolVal)

	case reflect.Struct:
		// 处理自定义结构体，尝试JSON解析
		return p.parseJSONToStruct(fieldValue, csvValue)

	case reflect.Slice:
		// 处理切片类型，尝试JSON解析
		return p.parseJSONToSlice(fieldValue, csvValue, fieldType)

	case reflect.Map:
		// 处理映射类型，尝试JSON解析
		return p.parseJSONToMap(fieldValue, csvValue, fieldType)

	case reflect.Ptr:
		// 处理指针类型
		return p.parsePointerField(fieldValue, csvValue, fieldType)

	default:
		return fmt.Errorf("unsupported field type: %v", fieldType.Kind())
	}

	return nil
}

// parseJSONToStruct 解析JSON到结构体
func (p *CSVParser) parseJSONToStruct(fieldValue reflect.Value, jsonStr string) error {
	// 创建结构体指针用于JSON解析
	structPtr := reflect.New(fieldValue.Type())

	err := json.Unmarshal([]byte(jsonStr), structPtr.Interface())
	if err != nil {
		return fmt.Errorf("failed to parse JSON to struct: %v", err)
	}

	// 设置值
	fieldValue.Set(structPtr.Elem())
	return nil
}

// parseJSONToSlice 解析JSON到切片
func (p *CSVParser) parseJSONToSlice(fieldValue reflect.Value, jsonStr string, fieldType reflect.Type) error {
	slicePtr := reflect.New(fieldType)

	err := json.Unmarshal([]byte(jsonStr), slicePtr.Interface())
	if err != nil {
		return fmt.Errorf("failed to parse JSON to slice: %v", err)
	}

	fieldValue.Set(slicePtr.Elem())
	return nil
}

// parseJSONToMap 解析JSON到映射
func (p *CSVParser) parseJSONToMap(fieldValue reflect.Value, jsonStr string, fieldType reflect.Type) error {
	mapPtr := reflect.New(fieldType)

	err := json.Unmarshal([]byte(jsonStr), mapPtr.Interface())
	if err != nil {
		return fmt.Errorf("failed to parse JSON to map: %v", err)
	}

	fieldValue.Set(mapPtr.Elem())
	return nil
}

// parsePointerField 解析指针字段
func (p *CSVParser) parsePointerField(fieldValue reflect.Value, csvValue string, fieldType reflect.Type) error {
	// 创建指针指向的类型的实例
	elemType := fieldType.Elem()
	elemPtr := reflect.New(elemType)

	// 递归解析
	err := p.setFieldValue(elemPtr.Elem(), csvValue, elemType)
	if err != nil {
		return err
	}

	fieldValue.Set(elemPtr)
	return nil
}
