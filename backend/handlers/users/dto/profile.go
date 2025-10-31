package dto

// ChangeDescriptionRequest represents the incoming payload when updating a profile description.
type ChangeDescriptionRequest struct {
	Description string `json:"description"`
}

// ChangeDisplayNameRequest represents the incoming payload when updating a display name.
type ChangeDisplayNameRequest struct {
	DisplayName string `json:"displayName"`
}

// ChangeEmojiRequest represents the incoming payload when updating a personal emoji.
type ChangeEmojiRequest struct {
	Emoji string `json:"emoji"`
}

// ChangePersonalLinksRequest represents the incoming payload when updating personal links.
type ChangePersonalLinksRequest struct {
	PersonalLink1 string `json:"personalLink1"`
	PersonalLink2 string `json:"personalLink2"`
	PersonalLink3 string `json:"personalLink3"`
	PersonalLink4 string `json:"personalLink4"`
}

// ChangePasswordRequest represents the incoming payload when updating a password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// PrivateUserResponse represents the shape returned by profile mutation endpoints.
type PrivateUserResponse struct {
	ID                    int64  `json:"id"`
	Username              string `json:"username"`
	DisplayName           string `json:"displayname"`
	UserType              string `json:"usertype"`
	InitialAccountBalance int64  `json:"initialAccountBalance"`
	AccountBalance        int64  `json:"accountBalance"`
	PersonalEmoji         string `json:"personalEmoji,omitempty"`
	Description           string `json:"description,omitempty"`
	PersonalLink1         string `json:"personalink1,omitempty"`
	PersonalLink2         string `json:"personalink2,omitempty"`
	PersonalLink3         string `json:"personalink3,omitempty"`
	PersonalLink4         string `json:"personalink4,omitempty"`
	Email                 string `json:"email"`
	APIKey                string `json:"apiKey,omitempty"`
	MustChangePassword    bool   `json:"mustChangePassword"`
}
