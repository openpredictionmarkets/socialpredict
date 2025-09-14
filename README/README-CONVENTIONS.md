# SocialPredict Coding Conventions

The following is a documentation of various coding conventions used to keep SocialPredict clean, maintainable and secure.

Further conventions will be added within the codebase within README files and will be assigned uuid's. To search for convention, simply search, "Convention," in the codebase.

## Seperation of Concerns

* The application is separated into a backend and a frontend so that people who specialize in different areas can easily contribute seperately.
* The backend is written in Golang to do the processing of data, transactions, hold models and so fourth. We use Golang because it is a highly performant language with a wide adoption rate. [In some scenarios, with the proper optimizations, Golang in terms of raw processing power can beat equivalent Rust code](https://www.reddit.com/r/golang/comments/16q78g2/re_golang_code_3x_faster_than_rust_equivalent/). (Though factoring in I/O this might not be the case).
* The frontend is built with ReactJS, similarly because as a platform it has a high adoption rate.

## Backend

### Follow All Golang Conventions First

* Follow Golang conventions per the standard Golang documentation and do not deviate. For any disagreements, refer to this documentation first, in order of the documentation's statements on what to refer to.

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


### Usage of Higher Order Functions

* Subject to Golang conventions, the use of higher order functions and polymorphism in general is encouraged to help with making things more testable and refactorable.

#### When To Use Higher Order Functions

- When you need to inject dependencies (like setup.EconConfigLoader) into an HTTP handler.
- When you want to create reusable, customizable behavior.
- When you want better testability (mocking dependencies easily).
- When you want clean separation between configuration/setup and actual execution.

#### When Not To Use Higher Order Functions

- When there is no meaningful dependency that needs to be passed.
- When it makes the code unnecessarily complex.
- When simple handler functions would be more readable and sufficient.

### 32-Bit Platform Compatibility

**Convention UUID: CONV-32BIT-001**

* When parsing string values to uint64 and then converting to platform-specific uint types, always validate that the value fits within the platform's uint size to prevent overflow on 32-bit systems.
* This convention addresses security concerns identified by GitHub's CodeQL analysis and ensures cross-platform compatibility.

#### The Problem

* On 64-bit platforms, `uint` is 64 bits and can hold any `uint64` value safely.
* On 32-bit platforms, `uint` is only 32 bits, so large `uint64` values will overflow when converted to `uint`.
* Direct conversion without validation can lead to unexpected behavior and potential security issues.

#### Recommended Implementation

**Option A: Direct Maximum Value Comparison (Recommended for Most Cases)**

This is the simplest and most readable approach:

```go
// 32-bit platform compatibility check (Convention CONV-32BIT-001)
// Ensure valueUint64 fits in a uint before casting
if valueUint64 > uint64(^uint(0)) {
    return errors.New("value exceeds allowed range for uint platform type")
}
valueUint := uint(valueUint64)
```

**Option B: Named Constants with Platform Detection (For Educational/Complex Cases)**

Use this when you need to understand or document the platform detection mechanism:

```go
// 32-bit platform compatibility check (Convention CONV-32BIT-001)
// Platform detection constants for 32-bit compatibility check
const (
    bitsInByte = 8
    bytesInUint32 = 4
    rightShiftFor64BitDetection = 63
    baseBitWidth = 32
)

// Detect platform bit width using named constants
maxUintValue := ^uint(0)
platformBitWidth := baseBitWidth << (maxUintValue >> rightShiftFor64BitDetection)
isPlatform32Bit := platformBitWidth == baseBitWidth

// Validate that the uint64 value fits in platform uint
if isPlatform32Bit && valueUint64 > math.MaxUint32 {
    http.Error(w, "Value out of range for platform", http.StatusBadRequest)
    return
}
valueUint := uint(valueUint64)
```

**When to Use Each Approach:**
- **Option A**: Use for most cases - simpler, more readable, equally effective
- **Option B**: Use when the platform detection logic needs to be explicit for educational purposes or when working with security-critical code where the mechanism should be transparent

#### What NOT To Do (Avoid Magic Numbers)

```go
// BAD: Uses magic numbers without explanation
if uintSize := 32 << (^uint(0) >> 63); uintSize == 32 && valueUint64 > math.MaxUint32 {
    // This is unclear and unmaintainable
}
```

#### How The Platform Detection Works

* `^uint(0)` creates the maximum value for the platform's uint type (all bits set to 1)
* On 64-bit platforms: `^uint(0)` has 64 ones, so `>> 63` yields 1, making `32 << 1 = 64`
* On 32-bit platforms: `^uint(0)` has 32 ones, so `>> 63` yields 0, making `32 << 0 = 32`
* This allows runtime detection of whether we're on a 32-bit or 64-bit platform

#### Usage Pattern

Apply this pattern whenever:
- Parsing string IDs to uint64 and then converting to uint
- Handling user input that gets converted to platform-specific types
- Working with numeric values that might exceed 32-bit limits

To find all implementations of this convention, search for "Convention CONV-32BIT-001" in the codebase.
