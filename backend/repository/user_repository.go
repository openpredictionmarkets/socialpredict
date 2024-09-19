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
	// Use Where to search by username column
	result := repo.db.Where("username = ?", username).First(&user)
	if err := result.Error(); err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *UserRepository) CountUsers() (int64, error) {
	var count int64
	if err := repo.db.Model(&models.User{}).Count(&count).Error(); err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *UserRepository) CountRegularUsers() (int64, error) {
	var count int64
	if err := repo.db.Model(&models.User{}).Where("user_type = ?", "REGULAR").Count(&count).Error(); err != nil {
		return 0, err
	}
	return count, nil
}
