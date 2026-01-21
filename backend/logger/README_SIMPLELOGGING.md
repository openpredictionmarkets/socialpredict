### SocialPredict Logging — simplelogging.go Conventions & Usage Overview

Purpose: provide consistent INFO, WARN, and ERROR lines that include the original caller’s file:line plus your message.

Output format (stdout): YYYY/MM/DD HH:MM:SS LEVEL file.go:line logger.(*CustomLogger).Level() - Context - Function: Message

Where to view (Docker): run and watch backend logs
docker logs -f socialpredict-backend-container

Current Logger with levels: (backend/logger/simplelogging.go)

```
// LogInfo is a convenience function to log informational messages with context.
func LogInfo(context, function, message string) {
logger := NewCustomLogger(os.Stdout, "", log.LstdFlags)
logger.Info(fmt.Sprintf("%s - %s: %s", context, function, message))
}

// LogWarn is a convenience function to log warning messages with context.
func LogWarn(context, function, message string) {
logger := NewCustomLogger(os.Stdout, "", log.LstdFlags)
logger.Warn(fmt.Sprintf("%s - %s: %s", context, function, message))
}

// LogError is a convenience function to log errors with context.
func LogError(context, function string, err error) {
logger := NewCustomLogger(os.Stdout, "", log.LstdFlags)
logger.Error(fmt.Sprintf("%s - %s: %v", context, function, err))
}
```

#### Usage in code (handlers or core)

```
Import the logger package
import "socialpredict/logger"

Log at key points in your function

logger.LogInfo("ChangePassword", "ChangePassword", "ChangePassword handler called")

securityService := security.NewSecurityService()
db := util.GetDB()

user, httperr := auth.ValidateTokenAndGetUser(r, usersSvc)
if httperr != nil {
http.Error(w, "Invalid token: "+httperr.Error(), http.StatusUnauthorized)
logger.LogError("ChangePassword", "ValidateTokenAndGetUser", httperr)
return
}

if _, err := securityService.Sanitizer.SanitizePassword(req.NewPassword); err != nil {
http.Error(w, err.Error(), http.StatusBadRequest)
logger.LogError("ChangePassword", "ValidateNewPasswordStrength", err)
return
}
```

#### Conventions for message content

Format: Context - Function: Message

Context examples: ChangePassword, BuyPosition, ResolveMarket

Function examples: DecodeRequestBody, ValidateInputFields, UpdatePasswordInDB

Message: concise, human-readable, and safe

Do not include passwords, raw tokens, or full request bodies

Example outputs

```
2025/10/06 11:36:13 INFO changepassword.go:25 logger.(*CustomLogger).Info() - ChangePassword - ChangePassword: ChangePassword handler called

2025/10/06 11:36:13 ERROR changepassword.go:54 logger.(*CustomLogger).Error() - ChangePassword - ValidateInputFields: New password is required
```
