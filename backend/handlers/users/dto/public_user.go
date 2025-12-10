package dto

// PublicUserResponse represents the public-facing user data returned by HTTP handlers.
type PublicUserResponse struct {
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
}
