package repository

import (
	"socialpredict/models"
)

type UserRepository struct {
	db Database
}

func NewUserRepository(db Database) *UserRepository {
	return &UserRepository{db: db}
}

func (repo *UserRepository) GetAllUsers() ([]models.User, error) {
	var users []models.User
	result := repo.db.Find(&users)
	if err := result.Error(); err != nil {
		return nil, err
	}
	return users, nil
}

func (repo *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	result := repo.db.First(&user, username)
	if err := result.Error(); err != nil {
		return nil, err
	}
	return &user, nil
}
