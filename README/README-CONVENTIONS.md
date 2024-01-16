# SocialPredict Coding Conventions

The following is a documentation of various coding conventions used to keep SocialPredict clean, maintainable and secure.


## Backend

### Specific Data Structs for Private and Public Responses

* For certain sensitive databases such as Users, there is Private information and Public information. When information from the Users table is needed to be combined with other types of tables such as Bets or Markets, we must be sure to use a function which specifically only returns public information that we explicitly allow, to prevent private information or even private fields from leaking through into API responses.
* This provides a balance between creating a more secure application and reducing queries on the database or building confusing queries on the database side.
* If private information is needed, for example in a User's profile page, then a different struct and retrival function must be used.
* The below is an example and the packages or code may have changed since this was authored, but the principle remainds the same.

#### Example Type Setup and Retrieval Function

```
// PublicUserType is a struct for user data that is safe to send to the client for Profiles
type PublicUserType struct {
	Username              string  `json:"username"`
	DisplayName           string  `json:"displayname" gorm:"unique;not null"`
	UserType              string  `json:"usertype"`
	InitialAccountBalance float64 `json:"initialAccountBalance"`
	AccountBalance        float64 `json:"accountBalance"`
	PersonalEmoji         string  `json:"personalEmoji,omitempty"`
	Description           string  `json:"description,omitempty"`
	PersonalLink1         string  `json:"personalink1,omitempty"`
	PersonalLink2         string  `json:"personalink2,omitempty"`
	PersonalLink3         string  `json:"personalink3,omitempty"`
	PersonalLink4         string  `json:"personalink4,omitempty"`
}

...

// Function to get the Info From the Database
func GetPublicUserInfo(db *gorm.DB, username string) PublicUserType {

	var user models.User
	db.Where("username = ?", username).First(&user)

	return PublicUserType{
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
	}
}
```

#### Example Usage In Another Package

* Note that `publicCreator` is used as the value Creator: in the response struct.
* If we had alternatively just used, "user" from the database, private info would be leaked.
* If we had not explicitly built the custom struct above, certain sensitive fields such as, "email:" could have been leaked, even if they were blank. This is unprofessional.
* Using the custom struct from a single package promotes code reusability.

```
	// get market creator
	// Fetch the Creator's public information using utility function
	publicCreator := usersHandlers.GetPublicUserInfo(db, market.CreatorUsername)

	// Manually construct the response
	response := struct {
		Market             PublicResponseMarket                   `json:"market"`
		Creator            usersHandlers.PublicUserType           `json:"creator"`
		ProbabilityChanges []marketMathHandlers.ProbabilityChange `json:"probabilityChanges"`
		NumUsers           int                                    `json:"numUsers"`
		TotalVolume        float64                                `json:"totalVolume"`
	}{
		Market:             responseMarket,
		Creator:            publicCreator,
		ProbabilityChanges: probabilityChanges,
		NumUsers:           numUsers,
		TotalVolume:        marketVolume,
	}

```

### Time-Based Validations Occur on Server Side, Not Client Side

* While logic could be built on the client side that governs the display of buttons, we don't validate time-based actions based upon client time.
* Hypothetically a user could manipulate their browser time to be running in the past, so we don't rely on browsers to tell us what time it is for the purposes of rulemaking. If a user wants to fiddle with the interface and show themselves action that won't be taken, we don't care, but they can't take an action on the API.