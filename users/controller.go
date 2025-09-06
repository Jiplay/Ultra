package users

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Controller struct {
	repo Repository
}

func NewController(repo Repository) *Controller {
	return &Controller{repo: repo}
}

func (c *Controller) CreateUser(req *CreateUserRequest) (*UserResponse, error) {
	if req.Email == "" {
		return nil, errors.New("email is required")
	}
	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	exists, err := c.repo.UserExists(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user already exists")
	}

	user := &User{
		Email: req.Email,
		Name:  req.Name,
	}

	err = c.repo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return c.userToResponse(user), nil
}

func (c *Controller) GetUserByID(userID string) (*UserResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	user, err := c.repo.GetUserByID(objectID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return c.userToResponse(user), nil
}

func (c *Controller) UpdateProfile(userID string, req *UpdateProfileRequest) (*UserResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	updateFields := bson.M{}

	if req.Name != nil {
		updateFields["name"] = *req.Name
	}
	if req.Email != nil {
		if *req.Email != "" {
			exists, err := c.repo.UserExists(*req.Email)
			if err != nil {
				return nil, err
			}
			if exists {
				existing, err := c.repo.GetUserByEmail(*req.Email)
				if err != nil {
					return nil, err
				}
				if existing.ID != objectID {
					return nil, errors.New("email already in use")
				}
			}
		}
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

	if len(updateFields) == 0 {
		return nil, errors.New("no fields to update")
	}

	user, err := c.repo.UpdateUser(objectID, updateFields)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return c.userToResponse(user), nil
}

func (c *Controller) DeleteUser(userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	err = c.repo.DeleteUser(objectID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("user not found")
		}
		return err
	}

	return nil
}

func (c *Controller) userToResponse(user *User) *UserResponse {
	return &UserResponse{
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
}