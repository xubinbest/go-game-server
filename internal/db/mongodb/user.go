package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/interfaces"
	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/snowflake"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBUserDatabase 实现 UserDatabase 接口
type MongoDBUserDatabase struct {
	client   *mongo.Client
	database string
	sf       *snowflake.Snowflake
}

// NewMongoDBUserDatabase 创建 MongoDBUserDatabase 实例
func NewMongoDBUserDatabase(client *mongo.Client, database string, sf *snowflake.Snowflake) interfaces.UserDatabase {
	return &MongoDBUserDatabase{
		client:   client,
		database: database,
		sf:       sf,
	}
}

// 获取用户集合
func (m *MongoDBUserDatabase) collection() *mongo.Collection {
	return m.client.Database(m.database).Collection("users")
}

// GetUserByUsername 根据用户名获取用户
func (m *MongoDBUserDatabase) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	filter := bson.M{"username": username}

	var user models.User
	err := m.collection().FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // 用户不存在
		}
		return nil, err
	}

	return &user, nil
}

// GetUser 根据ID获取用户
func (m *MongoDBUserDatabase) GetUser(ctx context.Context, userID int64) (*models.User, error) {
	filter := bson.M{"_id": userID}

	var user models.User
	err := m.collection().FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // 用户不存在
		}
		return nil, err
	}

	return &user, nil
}

// CreateUser 创建新用户
func (m *MongoDBUserDatabase) CreateUser(ctx context.Context, user *models.User) error {
	if user.ID == 0 {
		var err error
		user.ID, err = m.sf.NextID()
		if err != nil {
			return fmt.Errorf("failed to generate ID: %w", err)
		}
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	// MongoDB使用_id作为主键
	document := bson.M{
		"_id":           user.ID,
		"username":      user.Username,
		"level":         user.Level,
		"exp":           user.Exp,
		"email":         user.Email,
		"password_hash": user.PasswordHash,
		"salt":          user.Salt,
		"role":          user.Role,
		"created_at":    user.CreatedAt,
		"update_at":     user.UpdatedAt,
	}

	_, err := m.collection().InsertOne(ctx, document)
	if err != nil {
		// 检查用户名是否已存在
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("username already exists")
		}
		return err
	}

	return nil
}

// UpdateUser 更新用户信息
func (m *MongoDBUserDatabase) UpdateUser(ctx context.Context, user *models.User) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": bson.M{
			"username":      user.Username,
			"level":         user.Level,
			"exp":           user.Exp,
			"email":         user.Email,
			"password_hash": user.PasswordHash,
			"salt":          user.Salt,
			"role":          user.Role,
			"update_at":     user.UpdatedAt,
		},
	}

	result, err := m.collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// DeleteUser 删除用户
func (m *MongoDBUserDatabase) DeleteUser(ctx context.Context, userID int64) error {
	filter := bson.M{"_id": userID}

	result, err := m.collection().DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// 获取月签到集合
func (m *MongoDBUserDatabase) monthlySignCollection() *mongo.Collection {
	return m.client.Database(m.database).Collection("monthly_signs")
}

// 获取月签到奖励集合
func (m *MongoDBUserDatabase) monthlySignRewardCollection() *mongo.Collection {
	return m.client.Database(m.database).Collection("monthly_sign_rewards")
}

// GetMonthlySign 获取用户月签到信息
func (m *MongoDBUserDatabase) GetMonthlySign(ctx context.Context, userID int64) (*models.MonthlySign, error) {
	filter := bson.M{
		"_id": userID,
	}

	var sign models.MonthlySign
	err := m.monthlySignCollection().FindOne(ctx, filter).Decode(&sign)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // 没有签到记录
		}
		return nil, err
	}

	return &sign, nil
}

// CreateOrUpdateMonthlySign 创建或更新月签到信息
func (m *MongoDBUserDatabase) CreateOrUpdateMonthlySign(ctx context.Context, sign *models.MonthlySign) error {
	if sign.CreatedAt.IsZero() {
		sign.CreatedAt = time.Now()
	}
	sign.UpdatedAt = time.Now()

	filter := bson.M{
		"_id": sign.UserID,
	}

	update := bson.M{
		"$set": bson.M{
			"year":           sign.Year,
			"month":          sign.Month,
			"sign_days":      sign.SignDays,
			"last_sign_time": sign.LastSignTime,
			"updated_at":     sign.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"_id":        sign.UserID,
			"created_at": sign.CreatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := m.monthlySignCollection().UpdateOne(ctx, filter, update, opts)
	return err
}

// GetMonthlySignReward 获取用户月签到累计奖励记录
func (m *MongoDBUserDatabase) GetMonthlySignReward(ctx context.Context, userID int64) (*models.MonthlySignReward, error) {
	filter := bson.M{
		"_id": userID,
	}

	var reward models.MonthlySignReward
	err := m.monthlySignRewardCollection().FindOne(ctx, filter).Decode(&reward)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // 没有奖励记录
		}
		return nil, err
	}

	return &reward, nil
}

// CreateOrUpdateMonthlySignReward 创建或更新月签到累计奖励记录
func (m *MongoDBUserDatabase) CreateOrUpdateMonthlySignReward(ctx context.Context, reward *models.MonthlySignReward) error {
	if reward.CreatedAt.IsZero() {
		reward.CreatedAt = time.Now()
	}
	reward.UpdatedAt = time.Now()

	filter := bson.M{
		"_id": reward.UserID,
	}

	update := bson.M{
		"$set": bson.M{
			"year":        reward.Year,
			"month":       reward.Month,
			"reward_days": reward.RewardDays,
			"updated_at":  reward.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"_id":        reward.UserID,
			"created_at": reward.CreatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := m.monthlySignRewardCollection().UpdateOne(ctx, filter, update, opts)
	return err
}

// CreateMonthlySign 创建月签到记录（用于首次签到）
func (m *MongoDBUserDatabase) CreateMonthlySign(ctx context.Context, sign *models.MonthlySign) error {
	if sign.CreatedAt.IsZero() {
		sign.CreatedAt = time.Now()
	}
	sign.UpdatedAt = time.Now()

	_, err := m.monthlySignCollection().InsertOne(ctx, sign)
	return err
}

// 其他用户相关方法实现...
