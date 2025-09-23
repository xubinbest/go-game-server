package models

// Inventory 定义用户背包数据模型
type Inventory struct {
	UserID   int64            `json:"user_id" bson:"user_id,omitempty"`
	Items    []*InventoryItem `json:"items" bson:"items,omitempty"`
	Capacity int32            `json:"capacity" bson:"capacity"` // 背包容量
}

// InventoryItem 背包物品模型
type InventoryItem struct {
	ID         int64 `json:"id" bson:"id" gorm:"primaryKey;autoIncrement:false"`
	UserID     int64 `json:"user_id" bson:"user_id" gorm:"type:bigint;not null;index"`
	TemplateID int64 `json:"template_id" bson:"template_id" gorm:"type:bigint;not null"`
	Count      int32 `json:"count" bson:"count" gorm:"type:int;default:1;not null"`
	Equipped   bool  `json:"equipped" bson:"equipped" gorm:"type:boolean;default:false;not null"`
	CreatedAt  int64 `json:"created_at" bson:"created_at" gorm:"type:bigint;not null"`
	UpdatedAt  int64 `json:"updated_at" bson:"updated_at" gorm:"type:bigint;not null"`
}

func (InventoryItem) TableName() string {
	return "inventory_items"
}

// Equipment 装备模型
type Equipment struct {
	ID         int64 `json:"id" bson:"id" gorm:"primaryKey;autoIncrement:false"`
	UserID     int64 `json:"user_id" bson:"user_id" gorm:"type:bigint;not null;index"`
	TemplateID int64 `json:"template_id" bson:"template_id" gorm:"type:bigint;not null"`
	Slot       int32 `json:"slot" bson:"slot" gorm:"type:int;not null"`
	CreatedAt  int64 `json:"created_at" bson:"created_at" gorm:"type:bigint;not null"`
	UpdatedAt  int64 `json:"updated_at" bson:"updated_at" gorm:"type:bigint;not null"`
}

func (Equipment) TableName() string {
	return "equipments"
}
