package users

import (
	"net/http"
	
	"go.mongodb.org/mongo-driver/mongo"
)

func Setup(db *mongo.Database, mux *http.ServeMux) {
	collection := db.Collection("users")
	repo := NewMongoRepository(collection)
	controller := NewController(repo)
	handler := NewHandler(controller)
	
	RegisterRoutes(mux, handler)
}