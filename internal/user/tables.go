package user

import (
	"reflect"

	"github.xubinbest.com/go-game-server/internal/designconfig"
)

// 所有用到的配置表
var Tables = []designconfig.Tables{
	{
		DataId:    "item.csv",
		TableName: "item",
		DataType:  reflect.TypeOf(designconfig.ItemData{}),
		Group:     designconfig.BaseGroup,
	},
	{
		DataId:    "level.csv",
		TableName: "level",
		DataType:  reflect.TypeOf(designconfig.LevelData{}),
		Group:     designconfig.BaseGroup,
	},
	{
		DataId:    "equipment.csv",
		TableName: "equipment",
		DataType:  reflect.TypeOf(designconfig.EquipmentData{}),
		Group:     designconfig.BaseGroup,
	},
	{
		DataId:    "pet.csv",
		TableName: "pet",
		DataType:  reflect.TypeOf(designconfig.PetData{}),
		Group:     designconfig.BaseGroup,
	},
	{
		DataId:    "pet_level.csv",
		TableName: "pet_level",
		DataType:  reflect.TypeOf(designconfig.PetLevelData{}),
		Group:     designconfig.BaseGroup,
	},
	{
		DataId:    "card.csv",
		TableName: "card",
		DataType:  reflect.TypeOf(designconfig.CardData{}),
		Group:     designconfig.BaseGroup,
	},
	{
		DataId:    "card_star.csv",
		TableName: "card_star",
		DataType:  reflect.TypeOf(designconfig.CardStarData{}),
		Group:     designconfig.BaseGroup,
	},
	{
		DataId:    "card_level.csv",
		TableName: "card_level",
		DataType:  reflect.TypeOf(designconfig.CardLevelData{}),
		Group:     designconfig.BaseGroup,
	},
	{
		DataId:    "monthly_sign.csv",
		TableName: "monthly_sign",
		DataType:  reflect.TypeOf(designconfig.MonthlySignData{}),
		Group:     designconfig.BaseGroup,
	},
	{
		DataId:    "monthly_sign_cumulative.csv",
		TableName: "monthly_sign_cumulative",
		DataType:  reflect.TypeOf(designconfig.MonthlySignCumulativeData{}),
		Group:     designconfig.BaseGroup,
	},
}
