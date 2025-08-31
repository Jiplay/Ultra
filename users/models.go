package users

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string             `bson:"email" json:"email"`
	Name      string             `bson:"name" json:"name"`
	Weight    float64            `bson:"weight" json:"weight"`
	Height    float64            `bson:"height" json:"height"`
	Age       int                `bson:"age" json:"age"`
	Picture   string             `bson:"picture" json:"picture"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateUserRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required"`
}

type UpdateProfileRequest struct {
	Name    *string  `json:"name,omitempty"`
	Email   *string  `json:"email,omitempty"`
	Weight  *float64 `json:"weight,omitempty"`
	Height  *float64 `json:"height,omitempty"`
	Age     *int     `json:"age,omitempty"`
	Picture *string  `json:"picture,omitempty"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Weight    float64   `json:"weight"`
	Height    float64   `json:"height"`
	Age       int       `json:"age"`
	Picture   string    `json:"picture"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}