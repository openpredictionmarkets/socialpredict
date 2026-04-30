package adminhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
	configsvc "socialpredict/internal/service/config"
	"socialpredict/logger"
	"socialpredict/security"
	"strings"
)

type adminUserCreator interface {
	CreateAdminManagedUser(ctx context.Context, req dusers.AdminManagedUserCreateRequest) (*dusers.AdminManagedUserCreateResult, error)
}

func AddUserHandler(userCreator adminUserCreator, configService configsvc.Service, auth authsvc.Authenticator) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		responseData, handlerErr := processAddUser(r, userCreator, configService, auth)
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

func processAddUser(r *http.Request, userCreator adminUserCreator, configService configsvc.Service, auth authsvc.Authenticator) (map[string]interface{}, *handlerError) {
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

	if userCreator == nil {
		return nil, &handlerError{message: "user service unavailable", statusCode: http.StatusInternalServerError, reason: handlers.ReasonInternalError}
	}

	if auth == nil {
		return nil, &handlerError{message: "authentication service unavailable", statusCode: http.StatusInternalServerError, reason: handlers.ReasonInternalError}
	}
	if _, authErr := auth.RequireAdmin(r); authErr != nil {
		return nil, &handlerError{
			message:    authErr.Message,
			statusCode: authhttp.StatusCode(authErr),
			logErr:     errors.New(authErr.Message),
			reason:     authhttp.FailureReason(authErr),
		}
	}

	if configService == nil {
		return nil, &handlerError{message: "configuration service unavailable", statusCode: http.StatusInternalServerError, reason: handlers.ReasonInternalError}
	}

	appConfig := configService.Current()
	result, err := userCreator.CreateAdminManagedUser(r.Context(), dusers.AdminManagedUserCreateRequest{
		Username:              req.Username,
		InitialAccountBalance: appConfig.Economics.User.InitialAccountBalance,
	})
	if err != nil {
		return nil, addUserCreateError(err)
	}

	responseData := map[string]interface{}{
		"message":  "User created successfully",
		"username": result.Username,
		"password": result.Password,
		"usertype": result.UserType,
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

func addUserCreateError(err error) *handlerError {
	if errors.Is(err, dusers.ErrUserAlreadyExists) {
		validationErr := fmt.Errorf("username, display name, email, or API key already in use")
		return &handlerError{
			message:    validationErr.Error(),
			statusCode: http.StatusBadRequest,
			logErr:     validationErr,
			reason:     handlers.ReasonValidationFailed,
		}
	}

	return &handlerError{
		message:    "Failed to create user",
		statusCode: http.StatusInternalServerError,
		logErr:     err,
		reason:     handlers.ReasonInternalError,
	}
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
