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
