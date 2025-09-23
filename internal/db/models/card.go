package models

// Card 卡牌模型
type Card struct {
	ID         int64 `json:"id" bson:"id" gorm:"primaryKey;autoIncrement:false"`
	UserID     int64 `json:"user_id" bson:"user_id" gorm:"type:bigint;not null;index"`
	TemplateID int64 `json:"template_id" bson:"template_id" gorm:"type:bigint;not null"`
	Level      int32 `json:"level" bson:"level" gorm:"type:int;default:1;not null"`
	Star       int32 `json:"star" bson:"star" gorm:"type:int;default:1;not null"`
	CreatedAt  int64 `json:"created_at" bson:"created_at" gorm:"type:bigint;not null"`
	UpdatedAt  int64 `json:"updated_at" bson:"updated_at" gorm:"type:bigint;not null"`
}

func (Card) TableName() string {
	return "user_cards"
}

// CardCollection 定义用户卡牌集合数据模型
type CardCollection struct {
	UserID int64   `json:"user_id" bson:"user_id,omitempty"`
	Cards  []*Card `json:"cards" bson:"cards,omitempty"`
}
