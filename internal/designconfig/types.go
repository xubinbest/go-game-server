// 解析策划配置表
// 支持基础数据类型
// 自定义数据结构需要在配置表中配置成json格式，可以是struct或者slice
// 比如 attribute Attribute 或者 cost []BaseItemCost
package designconfig

import "reflect"

var BaseGroup = "PLAN_CONFIG"

// 用于配置服务需要哪些配置表
type Tables struct {
	DataId    string
	TableName string
	DataType  reflect.Type
	Group     string // 配置表分组
}

// 基础物品消耗
type BaseItemCost struct {
	ItemId int `json:"itemId"`
	Count  int `json:"count"`
}

// 属性
type Attribute struct {
	Atk   int `json:"atk"`
	Def   int `json:"def"`
	HpMax int `json:"hpMax"`
}

// 道具配置表
type ItemData struct {
	ID      int    `csv:"id"`
	Name    string `csv:"name"`
	Type    int    `csv:"type"`
	Subtype int    `csv:"subtype"`
	Color   int    `csv:"color"`
	Stack   int    `csv:"stack"`
}

// 等级配置表
type LevelData struct {
	Level     int       `csv:"level"`
	Exp       int       `csv:"exp"`
	Attribute Attribute `csv:"attribute"`
}

// 装备配置表
type EquipmentData struct {
	ID        int       `csv:"id"`
	Name      string    `csv:"name"`
	Attribute Attribute `csv:"attribute"`
}

// 宠物配置表
type PetData struct {
	ID    int    `csv:"id"`
	Name  string `csv:"name"`
	Color int    `csv:"color"`
}

// 宠物等级表
type PetLevelData struct {
	ID        int       `csv:"id"`
	PetId     int       `csv:"pet_id"`
	Level     int       `csv:"level"`
	Exp       int       `csv:"exp"`
	Attribute Attribute `csv:"attribute"`
}

// 卡牌配置表
type CardData struct {
	ID        int            `csv:"id"`
	Name      string         `csv:"name"`
	Color     int            `csv:"color"`
	Attribute Attribute      `csv:"attribute"`
	Cost      []BaseItemCost `csv:"cost"` // 激活卡牌消耗
}

// 卡牌星级配置表
type CardStarData struct {
	ID        int            `csv:"id"`
	CardId    int            `csv:"card_id"`
	Star      int            `csv:"star"`
	Attribute Attribute      `csv:"attribute"`
	Cost      []BaseItemCost `csv:"cost"` // 升星消耗
}

// 卡牌星级配置表
type CardLevelData struct {
	ID        int            `csv:"id"`
	CardId    int            `csv:"card_id"`
	Level     int            `csv:"level"`
	Attribute Attribute      `csv:"attribute"`
	Cost      []BaseItemCost `csv:"cost"` // 升级消耗
}

// 月签到配置表
type MonthlySignData struct {
	ID     int            `csv:"id"`
	Reward []BaseItemCost `csv:"reward"` // 奖励
}

// 月累计签到配置表
type MonthlySignCumulativeData struct {
	ID     int            `csv:"id"`
	Reward []BaseItemCost `csv:"reward"` // 奖励
}
