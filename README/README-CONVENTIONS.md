# SocialPredict Coding Conventions

The following is a documentation of various coding conventions used to keep SocialPredict clean, maintainable and secure.

## Seperation of Concerns

* The application is separated into a backend and a frontend so that people who specialize in different areas can easily contribute seperately.
* The backend is written in Golang to do the processing of data, transactions, hold models and so fourth. We use Golang because it is a highly performant language with a wide adoption rate. [In some scenarios, with the proper optimizations, Golang in terms of raw processing power can beat equivalent Rust code](https://www.reddit.com/r/golang/comments/16q78g2/re_golang_code_3x_faster_than_rust_equivalent/). (Though factoring in I/O this might not be the case).
* The frontend is built with ReactJS, similarly because as a platform it has a high adoption rate.

## Backend

### Ab Initio

* The entire program should be as stateless as possible, bets should be ideally made in one ledger and all reporting from that point on are calculated in a stateless manner.
* User balances may be cached, though in an ideal world user balance should also be calculated ab initio, and compared to the cached amount for verification.

### Points Start with Integers

* All transactions must start as integer amounts rather than float amounts. There are mathematical conventions which later transparently account for drops in points throughout market mechanisms.

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

### Definition of Handlers

* From the standpoint of both golang packages and the actual function names, "handlers," means something specific, e.g. something responds to an HTTP request. Following the conventon of the Golang http/net package, it specifically means [responds to an HTTP request](https://pkg.go.dev/net/http#Handler).
* So don't just call any old function a handler, make sure it has something to do with an http request, it is likely a hierarchically top level function that draws down from the API, e.g. it is the first function that is called when someone hits `api/v0/whatever`.

#### Handler Response Types

* Ideally, every handler should have its own type which pre-designates what is included in the response.

### Database Connection Pooling

* We should use database connection pooling, e.g. starting likely mostly from a handler, we should set up a database connection such as:

```
db := util.GetDB()
```

* Then moving down from there, we should try to pass db into subsequent functions so that each query being done is using the same connection, rather than running `db := util.GetDB()` again and again within each subsequent function.