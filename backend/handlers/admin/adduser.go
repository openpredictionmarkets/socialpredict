package adminhandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"socialpredict/handlers"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
	configsvc "socialpredict/internal/service/config"
	"socialpredict/logger"
	"socialpredict/models"
	"socialpredict/security"
	"strings"

	"github.com/brianvoe/gofakeit"
	"gorm.io/gorm"
)

func AddUserHandler(db *gorm.DB, configService configsvc.Service, auth authsvc.Authenticator) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		responseData, handlerErr := processAddUser(r, db, configService, auth)
		if handlerErr != nil {
			_ = handlers.WriteFailure(w, handlerErr.statusCode, handlerErr.reason)
			logAddUserFailure(handlerErr)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(responseData)
		if username, ok := responseData["username"].(string); ok {
			logger.LogInfo("AddUser", "ProcessAddUser", "Created user "+username)
		}
	}
}

type handlerError struct {
	message    string
	statusCode int
	logErr     error
	reason     handlers.FailureReason
}

func processAddUser(r *http.Request, db *gorm.DB, configService configsvc.Service, auth authsvc.Authenticator) (map[string]interface{}, *handlerError) {
	securityService := security.NewSecurityService()
	req, decodeErr := decodeAddUserRequest(r)
	if decodeErr != nil {
		return nil, &handlerError{
			message:    decodeErr.Error(),
			statusCode: http.StatusBadRequest,
			logErr:     decodeErr,
			reason:     handlers.ReasonInvalidRequest,
		}
	}

	if err := validateAddUserUsername(securityService, req.Username); err != nil {
		return nil, &handlerError{
			message:    "Invalid username: " + err.Error(),
			statusCode: http.StatusBadRequest,
			logErr:     err,
			reason:     handlers.ReasonValidationFailed,
		}
	}
	req.Username, _ = securityService.Sanitizer.SanitizeUsername(req.Username)

	if db == nil {
		return nil, &handlerError{message: "database unavailable", statusCode: http.StatusInternalServerError, reason: handlers.ReasonInternalError}
	}

	if auth == nil {
		return nil, &handlerError{message: "authentication service unavailable", statusCode: http.StatusInternalServerError, reason: handlers.ReasonInternalError}
	}
	if _, httpErr := auth.RequireAdmin(r); httpErr != nil {
		return nil, &handlerError{
			message:    httpErr.Message,
			statusCode: httpErr.StatusCode,
			logErr:     errors.New(httpErr.Message),
			reason:     handlers.AuthFailureReason(httpErr.StatusCode, httpErr.Message),
		}
	}

	if configService == nil {
		return nil, &handlerError{message: "configuration service unavailable", statusCode: http.StatusInternalServerError, reason: handlers.ReasonInternalError}
	}

	appConfig := configService.Current()
	user := buildNewUser(db, req.Username, appConfig)

	if err := checkUniqueFields(db, &user); err != nil {
		return nil, &handlerError{
			message:    err.Error(),
			statusCode: http.StatusBadRequest,
			logErr:     err,
			reason:     handlers.ReasonValidationFailed,
		}
	}

	password, err := generateAndHashPassword(&user)
	if err != nil {
		return nil, &handlerError{
			message:    err.Error(),
			statusCode: http.StatusInternalServerError,
			logErr:     err,
			reason:     handlers.ReasonInternalError,
		}
	}

	if result := db.Create(&user); result.Error != nil {
		return nil, &handlerError{
			message:    "Failed to create user",
			statusCode: http.StatusInternalServerError,
			logErr:     result.Error,
			reason:     handlers.ReasonInternalError,
		}
	}

	responseData := map[string]interface{}{
		"message":  "User created successfully",
		"username": user.Username,
		"password": password,
		"usertype": user.UserType,
	}
	return responseData, nil
}

type addUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
}

func decodeAddUserRequest(r *http.Request) (addUserRequest, error) {
	var req addUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return addUserRequest{}, fmt.Errorf("Error decoding request body")
	}
	return req, nil
}

func validateAddUserUsername(securityService *security.SecurityService, username string) error {
	if err := securityService.Validator.ValidateStruct(addUserRequest{Username: username}); err != nil {
		return err
	}
	_, err := securityService.Sanitizer.SanitizeUsername(username)
	return err
}

func buildNewUser(db *gorm.DB, username string, appConfig *configsvc.AppConfig) models.User {
	return models.User{
		PublicUser: models.PublicUser{
			Username:              username,
			DisplayName:           dusers.UniqueDisplayName(db),
			UserType:              "REGULAR",
			InitialAccountBalance: appConfig.Economics.User.InitialAccountBalance,
			AccountBalance:        appConfig.Economics.User.InitialAccountBalance,
			PersonalEmoji:         randomEmoji(),
		},
		PrivateUser: models.PrivateUser{
			Email:  dusers.UniqueEmail(db),
			APIKey: dusers.GenerateUniqueAPIKey(db),
		},
		MustChangePassword: true,
	}
}

func generateAndHashPassword(user *models.User) (string, error) {
	password := gofakeit.Password(true, true, true, false, false, 12)
	if err := user.HashPassword(password); err != nil {
		return "", fmt.Errorf("Failed to hash password")
	}
	return password, nil
}

func checkUniqueFields(db *gorm.DB, user *models.User) error {
	// Check for existing users with the same username, display name, email, or API key.
	var count int64
	db.Model(&models.User{}).Where(
		"username = ? OR display_name = ? OR email = ? OR api_key = ?",
		user.Username, user.DisplayName, user.Email, user.APIKey,
	).Count(&count)

	if count > 0 {
		return fmt.Errorf("username, display name, email, or API key already in use")
	}

	return nil
}

func randomEmoji() string {
	emojis := []string{"😀", "😃", "😄", "😁", "😆"}
	return emojis[rand.Intn(len(emojis))]
}

func logAddUserFailure(handlerErr *handlerError) {
	if handlerErr == nil || handlerErr.logErr == nil {
		return
	}

	if handlerErr.statusCode >= http.StatusInternalServerError {
		logger.LogError("AddUser", "ProcessAddUser", handlerErr.logErr)
		return
	}

	message := strings.TrimSpace(handlerErr.message)
	if message == "" {
		message = handlerErr.logErr.Error()
	}
	logger.LogWarn("AddUser", "ProcessAddUser", message)
}
