package bets

import (
	"context"
	"errors"

	dbets "socialpredict/internal/domain/bets"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ dbets.PlaceUnitOfWork = (*GormRepository)(nil)

// PlaceBetTransaction commits the user debit and bet insert as one unit of work.
func (r *GormRepository) PlaceBetTransaction(ctx context.Context, fn dbets.PlaceTransactionFunc) error {
	// The transaction begins here and commits only after the callback returns nil.
	// Any callback error rolls back every tx-scoped repository write.
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, NewGormRepository(tx), newPlaceUserService(tx))
	})
}

func newPlaceUserService(tx *gorm.DB) *dusers.Service {
	users := placeUserRepository{db: tx}
	return dusers.NewServiceWithDependencies(dusers.ServiceDependencies{
		Reader:      users,
		BalanceRepo: users,
	}, nil, nil)
}

type placeUserRepository struct {
	db *gorm.DB
}

func (r placeUserRepository) GetByUsername(ctx context.Context, username string) (*dusers.User, error) {
	var dbUser models.User
	query := r.db.WithContext(ctx).Where("username = ?", username)
	if r.db.Dialector.Name() == "postgres" {
		// SocialPredict's primary DB is Postgres; row-level UPDATE locking
		// protects balance checks and debits from overlapping placements.
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dusers.ErrUserNotFound
		}
		return nil, err
	}
	return modelUserToDomain(&dbUser), nil
}

func (r placeUserRepository) UpdateBalance(ctx context.Context, username string, newBalance int64) error {
	result := r.db.WithContext(ctx).Model(&models.User{}).
		Where("username = ?", username).
		Update("account_balance", newBalance)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return dusers.ErrUserNotFound
	}
	return nil
}

func modelUserToDomain(user *models.User) *dusers.User {
	if user == nil {
		return nil
	}
	return &dusers.User{
		ID:                    user.ID,
		Username:              user.Username,
		DisplayName:           user.DisplayName,
		Email:                 user.Email,
		PasswordHash:          user.Password,
		UserType:              user.UserType,
		InitialAccountBalance: user.InitialAccountBalance,
		AccountBalance:        user.AccountBalance,
		PersonalEmoji:         user.PersonalEmoji,
		Description:           user.Description,
		PersonalLink1:         user.PersonalLink1,
		PersonalLink2:         user.PersonalLink2,
		PersonalLink3:         user.PersonalLink3,
		PersonalLink4:         user.PersonalLink4,
		APIKey:                user.APIKey,
		MustChangePassword:    user.MustChangePassword,
		CreatedAt:             user.CreatedAt,
		UpdatedAt:             user.UpdatedAt,
	}
}
