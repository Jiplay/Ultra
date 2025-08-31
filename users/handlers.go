package users

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"ultra/database"
)

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	collection := database.MongoDB.Collection("users")
	var existingUser User
	err := collection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	user := User{
		Email:     req.Email,
		Name:      req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	user.ID = result.InsertedID.(primitive.ObjectID)

	response := UserResponse{
		ID:        user.ID.Hex(),
		Email:     user.Email,
		Name:      user.Name,
		Weight:    user.Weight,
		Height:    user.Height,
		Age:       user.Age,
		Picture:   user.Picture,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	collection := database.MongoDB.Collection("users")
	var user User
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	response := UserResponse{
		ID:        user.ID.Hex(),
		Email:     user.Email,
		Name:      user.Name,
		Weight:    user.Weight,
		Height:    user.Height,
		Age:       user.Age,
		Picture:   user.Picture,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	updateFields := bson.M{
		"updated_at": time.Now(),
	}

	if req.Name != nil {
		updateFields["name"] = *req.Name
	}
	if req.Email != nil {
		updateFields["email"] = *req.Email
	}
	if req.Weight != nil {
		updateFields["weight"] = *req.Weight
	}
	if req.Height != nil {
		updateFields["height"] = *req.Height
	}
	if req.Age != nil {
		updateFields["age"] = *req.Age
	}
	if req.Picture != nil {
		updateFields["picture"] = *req.Picture
	}

	collection := database.MongoDB.Collection("users")
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": updateFields},
	)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	// Get updated user
	var user User
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		http.Error(w, "Failed to get updated user", http.StatusInternalServerError)
		return
	}

	response := UserResponse{
		ID:        user.ID.Hex(),
		Email:     user.Email,
		Name:      user.Name,
		Weight:    user.Weight,
		Height:    user.Height,
		Age:       user.Age,
		Picture:   user.Picture,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	collection := database.MongoDB.Collection("users")
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func RegisterRoutes(router *mux.Router) {
	userRouter := router.PathPrefix("/users").Subrouter()
	
	userRouter.HandleFunc("", CreateUserHandler).Methods("POST")
	userRouter.HandleFunc("/{id}", GetUserHandler).Methods("GET")
	userRouter.HandleFunc("/{id}", UpdateProfileHandler).Methods("PUT")
	userRouter.HandleFunc("/{id}", DeleteUserHandler).Methods("DELETE")
}