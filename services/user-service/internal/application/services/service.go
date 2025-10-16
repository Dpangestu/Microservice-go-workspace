package services

import (
	"bkc_microservice/services/user-service/internal/domain/entities"
	"bkc_microservice/services/user-service/internal/domain/repositories"
)

type UserService struct {
	UserRepo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) *UserService {
	return &UserService{UserRepo: repo}
}

func (s *UserService) GetUserProfile(id string) (*entities.User, error) {
	return s.UserRepo.FindByID(id)
}

func (s *UserService) CreateUser(user *entities.User) error {
	return s.UserRepo.Create(user)
}
