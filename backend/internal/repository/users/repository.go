package users

import (
	"context"
	"errors"

	"socialpredict/internal/domain/auth"
	positionsmath "socialpredict/internal/domain/math/positions"
	dusers "socialpredict/internal/domain/users"
	usermodels "socialpredict/internal/domain/users/models"
	"socialpredict/models"

	"gorm.io/gorm"
)

// GormRepository implements the users domain repository interface using GORM
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM-based users repository
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// GetByUsername retrieves a user by username
func (r *GormRepository) GetByUsername(ctx context.Context, username string) (*dusers.User, error) {
	var dbUser models.User

	err := r.db.WithContext(ctx).Where("username = ?", username).First(&dbUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dusers.ErrUserNotFound
		}
		return nil, err
	}

	return r.modelToDomain(&dbUser), nil
}

// UpdateBalance updates a user's account balance
func (r *GormRepository) UpdateBalance(ctx context.Context, username string, newBalance int64) error {
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

// Create creates a new user in the database
func (r *GormRepository) Create(ctx context.Context, user *dusers.User) error {
	dbUser := r.domainToModel(user)
	dbUser.MustChangePassword = true

	result := r.db.WithContext(ctx).Create(&dbUser)
	if result.Error != nil {
		// Check for unique constraint violations
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return dusers.ErrUserAlreadyExists
		}
		return result.Error
	}

	// Update the domain model with the generated ID
	user.ID = dbUser.ID
	return nil
}

// Update updates a user in the database
func (r *GormRepository) Update(ctx context.Context, user *dusers.User) error {
	dbUser := r.domainToModel(user)

	result := r.db.WithContext(ctx).Omit("api_key", "must_change_password").Save(&dbUser)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return dusers.ErrUserNotFound
	}

	return nil
}

// Delete removes a user from the database
func (r *GormRepository) Delete(ctx context.Context, username string) error {
	result := r.db.WithContext(ctx).Where("username = ?", username).Delete(&models.User{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return dusers.ErrUserNotFound
	}

	return nil
}

// List retrieves users with the given filters
func (r *GormRepository) List(ctx context.Context, filters usermodels.ListFilters) ([]*dusers.User, error) {
	query := r.db.WithContext(ctx).Model(&models.User{})

	if filters.UserType != "" {
		query = query.Where("user_type = ?", filters.UserType)
	}

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	query = query.Order("created_at DESC")

	var dbUsers []models.User
	if err := query.Find(&dbUsers).Error; err != nil {
		return nil, err
	}

	users := make([]*dusers.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = r.modelToDomain(&dbUser)
	}

	return users, nil
}

// ListUserBets retrieves all bets placed by the specified user ordered by placement time descending.
func (r *GormRepository) ListUserBets(ctx context.Context, username string) ([]*dusers.UserBet, error) {
	var bets []models.Bet
	if err := r.db.WithContext(ctx).
		Where("username = ?", username).
		Order("placed_at DESC").
		Find(&bets).Error; err != nil {
		return nil, err
	}

	result := make([]*dusers.UserBet, len(bets))
	for i, bet := range bets {
		result[i] = &dusers.UserBet{
			MarketID: bet.MarketID,
			PlacedAt: bet.PlacedAt,
		}
	}
	return result, nil
}

// GetMarketQuestion retrieves the question title for the specified market.
func (r *GormRepository) GetMarketQuestion(ctx context.Context, marketID uint) (string, error) {
	var market models.Market
	if err := r.db.WithContext(ctx).Select("question_title").Where("id = ?", marketID).First(&market).Error; err != nil {
		return "", err
	}
	return market.QuestionTitle, nil
}

// GetUserPositionInMarket calculates the user's position within the specified market.
func (r *GormRepository) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dusers.MarketUserPosition, error) {
	var market models.Market
	if err := r.db.WithContext(ctx).First(&market, marketID).Error; err != nil {
		return nil, err
	}

	var bets []models.Bet
	if err := r.db.WithContext(ctx).
		Where("market_id = ?", marketID).
		Order("placed_at ASC").
		Find(&bets).Error; err != nil {
		return nil, err
	}

	snapshot := positionsmath.MarketSnapshot{
		ID:               int64(market.ID),
		CreatedAt:        market.CreatedAt,
		IsResolved:       market.IsResolved,
		ResolutionResult: market.ResolutionResult,
	}

	position, err := positionsmath.CalculateMarketPositionForUser_WPAM_DBPM(snapshot, bets, username)
	if err != nil {
		return nil, err
	}

	return &dusers.MarketUserPosition{
		YesSharesOwned: position.YesSharesOwned,
		NoSharesOwned:  position.NoSharesOwned,
	}, nil
}

// ListUserMarkets returns markets the user has participated in ordered by last bet time.
func (r *GormRepository) ListUserMarkets(ctx context.Context, userID int64) ([]*dusers.UserMarket, error) {
	var dbMarkets []models.Market

	query := r.db.WithContext(ctx).Table("markets").
		Joins("join bets on bets.market_id = markets.id").
		Where("bets.user_id = ?", userID).
		Order("bets.created_at DESC").
		Distinct("markets.*").
		Find(&dbMarkets)

	if query.Error != nil {
		return nil, query.Error
	}

	markets := make([]*dusers.UserMarket, len(dbMarkets))
	for i, m := range dbMarkets {
		markets[i] = &dusers.UserMarket{
			ID:                      m.ID,
			QuestionTitle:           m.QuestionTitle,
			Description:             m.Description,
			OutcomeType:             m.OutcomeType,
			ResolutionDateTime:      m.ResolutionDateTime,
			FinalResolutionDateTime: m.FinalResolutionDateTime,
			UTCOffset:               m.UTCOffset,
			IsResolved:              m.IsResolved,
			ResolutionResult:        m.ResolutionResult,
			InitialProbability:      m.InitialProbability,
			YesLabel:                m.YesLabel,
			NoLabel:                 m.NoLabel,
			CreatorUsername:         m.CreatorUsername,
			CreatedAt:               m.CreatedAt,
			UpdatedAt:               m.UpdatedAt,
		}
	}

	return markets, nil
}

// GetCredentials returns the hashed password and password-change flag for the specified user.
func (r *GormRepository) GetCredentials(ctx context.Context, username string) (*auth.Credentials, error) {
	var user models.User
	if err := r.db.WithContext(ctx).
		Select("id", "password", "must_change_password").
		Where("username = ?", username).
		Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dusers.ErrUserNotFound
		}
		return nil, err
	}

	return &auth.Credentials{
		UserID:             user.ID,
		PasswordHash:       user.Password,
		MustChangePassword: user.MustChangePassword,
	}, nil
}

// UpdatePassword persists a new password hash and updates the must-change flag.
func (r *GormRepository) UpdatePassword(ctx context.Context, username string, hashedPassword string, mustChange bool) error {
	result := r.db.WithContext(ctx).Model(&models.User{}).
		Where("username = ?", username).
		Updates(map[string]any{
			"password":             hashedPassword,
			"must_change_password": mustChange,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return dusers.ErrUserNotFound
	}
	return nil
}

// GetAPIKey returns the API key for the specified user.
func (r *GormRepository) GetAPIKey(ctx context.Context, username string) (string, error) {
	var user models.User
	if err := r.db.WithContext(ctx).
		Select("api_key").
		Where("username = ?", username).
		Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", dusers.ErrUserNotFound
		}
		return "", err
	}
	return user.APIKey, nil
}

// domainToModel converts a domain user to a GORM model.
// Auth fields (APIKey, MustChangePassword) are not mapped here; they are
// managed through dedicated repository methods.
func (r *GormRepository) domainToModel(user *dusers.User) models.User {
	return models.User{
		ID: int64(user.ID),
		Model: gorm.Model{
			ID:        uint(user.ID),
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		PublicUser: models.PublicUser{
			Username:              user.Username,
			DisplayName:           user.DisplayName,
			UserType:              user.UserType,
			InitialAccountBalance: user.InitialAccountBalance,
			AccountBalance:        user.AccountBalance,
			PersonalEmoji:         user.PersonalEmoji,
			Description:           user.Description,
			PersonalLink1:         user.PersonalLink1,
			PersonalLink2:         user.PersonalLink2,
			PersonalLink3:         user.PersonalLink3,
			PersonalLink4:         user.PersonalLink4,
		},
		PrivateUser: models.PrivateUser{
			Email: user.Email,
		},
	}
}

// modelToDomain converts a GORM model to a domain user.
// Auth fields (APIKey, MustChangePassword) are not mapped; they live in the
// auth domain and are accessed through GetCredentials / GetAPIKey.
func (r *GormRepository) modelToDomain(dbUser *models.User) *dusers.User {
	return &dusers.User{
		ID:                    int64(dbUser.ID),
		Username:              dbUser.Username,
		DisplayName:           dbUser.DisplayName,
		Email:                 dbUser.Email,
		UserType:              dbUser.UserType,
		InitialAccountBalance: dbUser.InitialAccountBalance,
		AccountBalance:        dbUser.AccountBalance,
		PersonalEmoji:         dbUser.PersonalEmoji,
		Description:           dbUser.Description,
		PersonalLink1:         dbUser.PersonalLink1,
		PersonalLink2:         dbUser.PersonalLink2,
		PersonalLink3:         dbUser.PersonalLink3,
		PersonalLink4:         dbUser.PersonalLink4,
		CreatedAt:             dbUser.CreatedAt,
		UpdatedAt:             dbUser.UpdatedAt,
	}
}
