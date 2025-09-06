package users

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository interface {
	CreateUser(user *User) error
	GetUserByID(id primitive.ObjectID) (*User, error)
	GetUserByEmail(email string) (*User, error)
	UpdateUser(id primitive.ObjectID, updateFields bson.M) (*User, error)
	DeleteUser(id primitive.ObjectID) error
	UserExists(email string) (bool, error)
}

type MongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(collection *mongo.Collection) Repository {
	return &MongoRepository{collection: collection}
}

func (r *MongoRepository) CreateUser(user *User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	result, err := r.collection.InsertOne(context.Background(), user)
	if err != nil {
		return err
	}
	
	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *MongoRepository) GetUserByID(id primitive.ObjectID) (*User, error) {
	var user User
	err := r.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *MongoRepository) GetUserByEmail(email string) (*User, error) {
	var user User
	err := r.collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *MongoRepository) UpdateUser(id primitive.ObjectID, updateFields bson.M) (*User, error) {
	updateFields["updated_at"] = time.Now()
	
	_, err := r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		bson.M{"$set": updateFields},
	)
	if err != nil {
		return nil, err
	}
	
	return r.GetUserByID(id)
}

func (r *MongoRepository) DeleteUser(id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		return err
	}
	
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	
	return nil
}

func (r *MongoRepository) UserExists(email string) (bool, error) {
	var user User
	err := r.collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}