package markets

import (
	"context"
	"errors"

	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ dmarkets.GroupedMarketUnitOfWork = (*GormRepository)(nil)

// GroupedMarketTransaction commits grouped-market structural writes, child
// market writes, and balance mutations as one unit of work.
func (r *GormRepository) GroupedMarketTransaction(ctx context.Context, fn dmarkets.GroupedMarketTransactionFunc) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, NewGormRepository(tx), newGroupedMarketUserService(tx))
	})
}

func newGroupedMarketUserService(tx *gorm.DB) *dusers.Service {
	users := groupedMarketUserRepository{db: tx}
	return dusers.NewServiceWithDependencies(dusers.ServiceDependencies{
		Reader:      users,
		BalanceRepo: users,
	}, nil, nil)
}

type groupedMarketUserRepository struct {
	db *gorm.DB
}

func (r groupedMarketUserRepository) GetByUsername(ctx context.Context, username string) (*dusers.User, error) {
	var dbUser models.User
	query := r.db.WithContext(ctx).Where("username = ?", username)
	if r.db.Dialector.Name() == "postgres" {
		// Balance checks and fee/payout mutations must serialize per user.
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dusers.ErrUserNotFound
		}
		return nil, err
	}
	return groupedMarketUserModelToDomain(&dbUser), nil
}

func (r groupedMarketUserRepository) UpdateBalance(ctx context.Context, username string, newBalance int64) error {
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

func groupedMarketUserModelToDomain(user *models.User) *dusers.User {
	if user == nil {
		return nil
	}
	out := &dusers.User{
		ID:                        int64(user.ID),
		Username:                  user.Username,
		DisplayName:               user.DisplayName,
		Email:                     user.Email,
		PasswordHash:              user.Password,
		UserType:                  user.UserType,
		InitialAccountBalance:     user.InitialAccountBalance,
		AccountBalance:            user.AccountBalance,
		PersonalEmoji:             user.PersonalEmoji,
		Description:               user.Description,
		PersonalLink1:             user.PersonalLink1,
		PersonalLink2:             user.PersonalLink2,
		PersonalLink3:             user.PersonalLink3,
		PersonalLink4:             user.PersonalLink4,
		APIKey:                    user.APIKey,
		MustChangePassword:        user.MustChangePassword,
		ModeratorStatus:           dusers.ModeratorStatus(user.ModeratorStatus),
		ModeratorSuspensionReason: user.ModeratorSuspensionReason,
		ModeratorSuspendedBy:      user.ModeratorSuspendedBy,
		ModeratorSuspendedAt:      cloneTimePtr(user.ModeratorSuspendedAt),
		CreatedAt:                 user.CreatedAt,
		UpdatedAt:                 user.UpdatedAt,
	}
	out.NormalizeRoleState()
	return out
}
