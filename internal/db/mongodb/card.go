package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/interfaces"
	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/snowflake"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDBCardDatabase struct {
	client   *mongo.Client
	dbName   string
	sf       *snowflake.Snowflake
	cardColl *mongo.Collection
}

func NewMongoDBCardDatabase(client *mongo.Client, dbName string, sf *snowflake.Snowflake) interfaces.CardDatabase {
	return &MongoDBCardDatabase{
		client:   client,
		dbName:   dbName,
		sf:       sf,
		cardColl: client.Database(dbName).Collection("user_cards"),
	}
}

func (m *MongoDBCardDatabase) GetUserCards(ctx context.Context, userID int64) ([]*models.Card, error) {
	filter := bson.M{"user_id": userID}

	cursor, err := m.cardColl.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get user cards: %w", err)
	}
	defer cursor.Close(ctx)

	var cards []*models.Card
	if err = cursor.All(ctx, &cards); err != nil {
		return nil, fmt.Errorf("failed to decode cards: %w", err)
	}

	return cards, nil
}

func (m *MongoDBCardDatabase) GetUserCard(ctx context.Context, userID int64, cardID int64) (*models.Card, error) {
	filter := bson.M{"_id": cardID, "user_id": userID}

	var card models.Card
	err := m.cardColl.FindOne(ctx, filter).Decode(&card)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("card not found")
		}
		return nil, fmt.Errorf("failed to get user card: %w", err)
	}

	return &card, nil
}

func (m *MongoDBCardDatabase) CreateCard(ctx context.Context, card *models.Card) error {
	now := time.Now().Unix()
	cardID, err := m.sf.NextID()
	if err != nil {
		return fmt.Errorf("failed to generate card ID: %w", err)
	}
	card.ID = cardID
	card.CreatedAt = now
	card.UpdatedAt = now

	_, err = m.cardColl.InsertOne(ctx, card)
	if err != nil {
		return fmt.Errorf("failed to create card: %w", err)
	}

	return nil
}

func (m *MongoDBCardDatabase) UpdateCard(ctx context.Context, card *models.Card) error {
	card.UpdatedAt = time.Now().Unix()

	filter := bson.M{"_id": card.ID, "user_id": card.UserID}
	update := bson.M{
		"$set": bson.M{
			"template_id": card.TemplateID,
			"level":       card.Level,
			"star":        card.Star,
			"updated_at":  card.UpdatedAt,
		},
	}

	result, err := m.cardColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("card not found")
	}

	return nil
}

func (m *MongoDBCardDatabase) UpgradeCard(ctx context.Context, userID int64, cardID int64, newLevel int32) error {
	filter := bson.M{"_id": cardID, "user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"level":      newLevel,
			"updated_at": time.Now().Unix(),
		},
	}

	result, err := m.cardColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to upgrade card: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("card not found")
	}

	return nil
}

func (m *MongoDBCardDatabase) UpgradeCardStar(ctx context.Context, userID int64, cardID int64, newStar int32) error {
	filter := bson.M{"_id": cardID, "user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"star":       newStar,
			"updated_at": time.Now().Unix(),
		},
	}

	result, err := m.cardColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to upgrade card star: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("card not found")
	}

	return nil
}

func (m *MongoDBCardDatabase) CardExists(ctx context.Context, userID int64, templateID int64) (bool, error) {
	filter := bson.M{"template_id": templateID, "user_id": userID}

	count, err := m.cardColl.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check card exists: %w", err)
	}

	return count > 0, nil
}
