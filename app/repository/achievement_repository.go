// repository/achievement_repository.go (PERBAIKAN)
package repository

import (
	"context"
	"time"

	"achievement-backend/app/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AchievementRepository interface {
	// Basic CRUD
	Create(ctx context.Context, achievement *models.Achievement) (primitive.ObjectID, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Achievement, error)
	FindByStudentID(ctx context.Context, studentID uuid.UUID, page, limit int) ([]*models.Achievement, int, error)
	Update(ctx context.Context, id primitive.ObjectID, achievement *models.Achievement) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	
	// Advanced Queries
	FindWithFilter(ctx context.Context, filter bson.M, page, limit int) ([]*models.Achievement, int, error)
	FindByStudentIDs(ctx context.Context, studentIDs []uuid.UUID, page, limit int) ([]*models.Achievement, int, error)
	
	// Status-related
	FindByStatus(ctx context.Context, studentID uuid.UUID, status string, page, limit int) ([]*models.Achievement, int, error)
	
	// Statistics
	GetStatisticsByStudent(ctx context.Context, studentID uuid.UUID) (*models.AchievementStatistics, error)
	GetStatisticsByAdvisor(ctx context.Context, advisorID uuid.UUID) (*models.AchievementStatistics, error)
	
	// Aggregation
	CountByType(ctx context.Context, studentID uuid.UUID) (map[string]int, error)
	CountByPeriod(ctx context.Context, studentID uuid.UUID, startDate, endDate time.Time) (map[string]int, error)
}

type achievementRepo struct {
	collection *mongo.Collection
}

func NewAchievementRepository(db *mongo.Database) AchievementRepository {
	return &achievementRepo{
		collection: db.Collection("achievements"),
	}
}

func (r *achievementRepo) Create(ctx context.Context, achievement *models.Achievement) (primitive.ObjectID, error) {
	achievement.CreatedAt = time.Now()
	achievement.UpdatedAt = time.Now()
	
	if achievement.ID.IsZero() {
		achievement.ID = primitive.NewObjectID()
	}
	
	result, err := r.collection.InsertOne(ctx, achievement)
	if err != nil {
		return primitive.NilObjectID, err
	}
	
	return result.InsertedID.(primitive.ObjectID), nil
}

func (r *achievementRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Achievement, error) {
	var achievement models.Achievement
	filter := bson.M{"_id": id}
	
	err := r.collection.FindOne(ctx, filter).Decode(&achievement)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	
	return &achievement, nil
}

func (r *achievementRepo) FindByStudentID(ctx context.Context, studentID uuid.UUID, page, limit int) ([]*models.Achievement, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	skip := (page - 1) * limit
	
	filter := bson.M{"studentId": studentID.String()}
	
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})
	
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	
	var achievements []*models.Achievement
	if err := cursor.All(ctx, &achievements); err != nil {
		return nil, 0, err
	}
	
	return achievements, int(total), nil
}

func (r *achievementRepo) Update(ctx context.Context, id primitive.ObjectID, achievement *models.Achievement) error {
	achievement.UpdatedAt = time.Now()
	achievement.ID = id
	
	filter := bson.M{"_id": id}
	update := bson.M{"$set": achievement}
	
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	
	return nil
}

func (r *achievementRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	
	return nil
}

func (r *achievementRepo) FindWithFilter(ctx context.Context, filter bson.M, page, limit int) ([]*models.Achievement, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	skip := (page - 1) * limit
	
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})
	
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	
	var achievements []*models.Achievement
	if err := cursor.All(ctx, &achievements); err != nil {
		return nil, 0, err
	}
	
	return achievements, int(total), nil
}

func (r *achievementRepo) FindByStudentIDs(ctx context.Context, studentIDs []uuid.UUID, page, limit int) ([]*models.Achievement, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	skip := (page - 1) * limit
	
	var studentIDStrs []string
	for _, id := range studentIDs {
		studentIDStrs = append(studentIDStrs, id.String())
	}
	
	filter := bson.M{"studentId": bson.M{"$in": studentIDStrs}}
	
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})
	
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	
	var achievements []*models.Achievement
	if err := cursor.All(ctx, &achievements); err != nil {
		return nil, 0, err
	}
	
	return achievements, int(total), nil
}

func (r *achievementRepo) FindByStatus(ctx context.Context, studentID uuid.UUID, status string, page, limit int) ([]*models.Achievement, int, error) {
	// Note: Status ada di PostgreSQL, method ini placeholder saja
	return r.FindByStudentID(ctx, studentID, page, limit)
}

func (r *achievementRepo) GetStatisticsByStudent(ctx context.Context, studentID uuid.UUID) (*models.AchievementStatistics, error) {
	stats := &models.AchievementStatistics{
		ByType:   make(map[string]int),
		ByPeriod: make(map[string]int),
		TopTags:  []string{},
	}
	
	// Total achievements
	filter := bson.M{"studentId": studentID.String()}
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}
	stats.TotalAchievements = int(total)
	
	// Count by type
	typeCounts, err := r.CountByType(ctx, studentID)
	if err != nil {
		return nil, err
	}
	stats.ByType = typeCounts
	
	// Count by period (last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)
	periodCounts, err := r.CountByPeriod(ctx, studentID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	stats.ByPeriod = periodCounts
	
	// Average points
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "studentId", Value: studentID.String()}}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "avgPoints", Value: bson.D{{Key: "$avg", Value: "$points"}}},
			{Key: "totalPoints", Value: bson.D{{Key: "$sum", Value: "$points"}}},
		}}},
	}
	
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err == nil {
		defer cursor.Close(ctx)
		if cursor.Next(ctx) {
			var result struct {
				AvgPoints   float64 `bson:"avgPoints"`
				TotalPoints int     `bson:"totalPoints"`
			}
			if err := cursor.Decode(&result); err == nil {
				stats.AveragePoints = result.AvgPoints
			}
		}
	}
	
	return stats, nil
}

func (r *achievementRepo) GetStatisticsByAdvisor(ctx context.Context, advisorID uuid.UUID) (*models.AchievementStatistics, error) {
	// Ini perlu data dari PostgreSQL (student IDs per advisor)
	// Untuk sekarang, return empty
	return &models.AchievementStatistics{
		ByType:   make(map[string]int),
		ByPeriod: make(map[string]int),
		TopTags:  []string{},
	}, nil
}

func (r *achievementRepo) CountByType(ctx context.Context, studentID uuid.UUID) (map[string]int, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "studentId", Value: studentID.String()}}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$achievementType"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}
	
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	results := make(map[string]int)
	for cursor.Next(ctx) {
		var result struct {
			Type  string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		results[result.Type] = result.Count
	}
	
	return results, nil
}

func (r *achievementRepo) CountByPeriod(ctx context.Context, studentID uuid.UUID, startDate, endDate time.Time) (map[string]int, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "studentId", Value: studentID.String()},
			{Key: "createdAt", Value: bson.D{
				{Key: "$gte", Value: startDate},
				{Key: "$lte", Value: endDate},
			}},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "$dateToString", Value: bson.D{
				{Key: "format", Value: "%Y-%m"},
				{Key: "date", Value: "$createdAt"},
			}}}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}
	
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	results := make(map[string]int)
	for cursor.Next(ctx) {
		var result struct {
			Period string `bson:"_id"`
			Count  int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		results[result.Period] = result.Count
	}
	
	return results, nil
}