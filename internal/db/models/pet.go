package models

import (
	"time"
)

// Pet 宠物模型
type Pet struct {
	ID         int64     `json:"id" bson:"id" gorm:"primaryKey;autoIncrement:false"`
	UserID     int64     `json:"user_id" bson:"user_id" gorm:"type:bigint;not null;index"`
	TemplateID int64     `json:"template_id" bson:"template_id" gorm:"type:bigint;not null"`
	Name       string    `json:"name" bson:"name" gorm:"type:varchar(50);not null"`
	Level      int32     `json:"level" bson:"level" gorm:"type:int;default:1;not null"`
	Exp        int32     `json:"exp" bson:"exp" gorm:"type:int;default:0;not null"`
	IsBattle   bool      `json:"is_battle" bson:"is_battle" gorm:"type:boolean;default:false;not null"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at" gorm:"autoUpdateTime"`
}

func (Pet) TableName() string {
	return "pets"
}
