package userService

import (
	"log"
	"database/sql"
	"net/http"
	"errors"
	"strings"
	"strconv"
	"time"

	"github.com/Harsha-S2604/genz-server/models/users"
	"github.com/Harsha-S2604/genz-server/utilities/validations"
	"github.com/Harsha-S2604/genz-server/utilities/hashing"
	"github.com/Harsha-S2604/genz-server/utilities/verification"
	"github.com/gin-gonic/gin"
)

var (
	X_GENZ_TOKEN = "4439EA5BDBA8B179722265789D029477"
)

func ValidateUserLoginHandler(genzDB *sql.DB) gin.HandlerFunc {
	ValidateUserLogin := func(ctx *gin.Context) {
		var userFromRequest users.User
		var xGenzToken string 
		user := new(users.User);
		resultsCount := 0
		xGenzTokenArr, ok := ctx.Request.Header["X-Genz-Token"];

		if !ok {
			log.Println("Token not exists")
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry my friend, Invalid request. We'll fix it ASAP. Please refresh the page or try again later.",
			})
			return
		} else {
			xGenzToken = xGenzTokenArr[0]
			ctx.ShouldBindJSON(&userFromRequest)
		} 

		if X_GENZ_TOKEN != xGenzToken {
			log.Println("ERROR Function ValidateUserLogin: Invalid security key", userFromRequest.Email)
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry my friend, Invalid security key. We'll fix it ASAP. Please refresh the page or try again later.",
			})
			return
		}
		

		// validate email and hash the password
		isValidEmail, err := validations.ValidateUserEmail(userFromRequest.Email)
		hashedPassword := hashing.HashUserPassword(userFromRequest.Password)
		if err != nil {
			log.Println("ERROR Function ValidateUserLogin: ", err.Error(), userFromRequest.Email)
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry my friend, Email is not in a valid format.",
			})
			return
		} else {
			log.Println("is valid email", isValidEmail, userFromRequest.Email)
		}
		
		// Execute the query
		results, qryErr := genzDB.Query("SELECT email, password FROM users WHERE email = ?;", userFromRequest.Email)
		if qryErr != nil {
			log.Println("ERROR Function ValidateUserLogin: "+qryErr.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"success": false,
				"message": "Sorry my friend, something went wrong on our side. Our team is working on it. Please refresh the page or try again later.",
			})
			return
		}


		// bind the query results with user variable declared above
		for results.Next() {
			resultsCount += 1
			resultErr := results.Scan(&user.Email, &user.Password)
			if resultErr != nil {
				log.Println("ERROR Function ValidateUserLogin: "+resultErr.Error())
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code": http.StatusInternalServerError,
					"success": false,
					"message": "Sorry my friend, something went wrong on our side. Our team is working on it. Please refresh the page or try again later.",
				})
				return
			}
		}

		// return not registered if user not exists.
		if resultsCount == 0 {
			log.Println("User not found "+userFromRequest.Email)
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Hi friend! looks like you have not registered yet. Please register.",
			})
		} else {
			// check if user credentials match or not and send appropriate messages
			if user.Email == userFromRequest.Email && user.Password == hashedPassword {
				var userData users.User
				log.Println("User credentials match",userFromRequest.Email)
				loggedinUser, qryMatchErr := genzDB.Query("SELECT user_id, name, email, is_email_verified FROM users WHERE email = ?;", userFromRequest.Email)
				if qryMatchErr != nil {
					log.Println("ERROR Function ValidateUserLogin: "+qryMatchErr.Error())
					ctx.JSON(http.StatusInternalServerError, gin.H{
						"code": http.StatusInternalServerError,
						"success": false,
						"message": "Sorry my friend, something went wrong on our side. Our team is working on it. Please refresh the page or try again later.",
					})
					return
				}
				for loggedinUser.Next() {
					loggedinUserErr := loggedinUser.Scan(&userData.UserId, &userData.Name, &userData.Email, &userData.IsEmailVerified)
					if loggedinUserErr != nil {
						log.Println("ERROR Function ValidateUserLogin: "+loggedinUserErr.Error())
						ctx.JSON(http.StatusInternalServerError, gin.H{
							"code": http.StatusInternalServerError,
							"success": false,
							"message": "Sorry my friend, something went wrong on our side. Our team is working on it. Please refresh the page or try again later.",
						})
						return
					}
				}
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": true,
					"data": userData,
					"message": "You are now logged in. You will soon be redirected.",
				})
			} else {
				log.Println("User credentials don't match",userFromRequest.Email)
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Incorrect email or password.",
				})
			}
		}
	}

	return gin.HandlerFunc(ValidateUserLogin)
}

func UserRegisterHandler(genzDB *sql.DB) gin.HandlerFunc {
	
	UserRegister := func(ctx *gin.Context) {

		var userFromRequest users.User
		var xGenzToken string 
		// user := new(users.User)
		xGenzTokenArr, ok := ctx.Request.Header["X-Genz-Token"];

		if !ok {
			log.Println("Token not exists.")
			ctx.JSON(http.StatusOK, gin.H {
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry my friend, Invalid request. We'll fix it ASAP. Please refresh the page or try again later.",
			})
		} else {
			xGenzToken = xGenzTokenArr[0]
			ctx.ShouldBindJSON(&userFromRequest)

			if X_GENZ_TOKEN != xGenzToken {
				log.Println("ERROR Function UserRegister: Invalid security key", userFromRequest.Email)
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry my friend, Invalid security key. We'll fix it ASAP. Please refresh the page or try again later.",
				})
			} else {

				// validate email and hash the password
				isValidEmail, isValidEmailErr := validations.ValidateUserEmail(userFromRequest.Email)
				isUserExists := checkUserExists(genzDB, userFromRequest.Email)
				if isUserExists {
					log.Println("User already exists", userFromRequest.Email)
					ctx.JSON(http.StatusOK, gin.H{
						"code": http.StatusOK,
						"success": false,
						"message": "User already exists.",
					})
				} else {
					log.Println("Registering... user", userFromRequest.Email)
					hashedPassword := hashing.HashUserPassword(userFromRequest.Password)
					if isValidEmailErr != nil {
						log.Println("ERROR Function UserRegister: ", isValidEmailErr.Error(), userFromRequest.Email)
						ctx.JSON(http.StatusOK, gin.H{
							"code": http.StatusOK,
							"success": false,
							"message": "Sorry my friend, Email is not in a valid format.",
						})
					} else {
						log.Println("is valid email", isValidEmail, userFromRequest.Email)
						userId, generateIdErr := generateUserId(genzDB)
						if generateIdErr != nil {
							log.Println("ERROR Function UserRegister: USER ID generation error", generateIdErr.Error())
							ctx.JSON(http.StatusInternalServerError, gin.H{
								"code": http.StatusInternalServerError,
								"success": false,
								"message": "Sorry my friend, something went wrong on our side. Our team is working on it. Please refresh the page or try again later.",
							})
						} else {
							log.Println("Generated userid: ", userId)

							// generate the verification code
							verificationCode, verificationCodeErr := verification.GenerateSixDigitCode()
							if verificationCodeErr != nil {
								ctx.JSON(http.StatusOK, gin.H{
									"code": http.StatusOK,
									"success": false,
									"message": "User registration failed. Please try again later.",
								})
								return
							}
							log.Println("verification code", verificationCode)
							// Execute the query
							insertResults, qryErr := genzDB.ExecContext(ctx, "INSERT INTO users VALUES(?, ?, ?, ?, ?);", userId, userFromRequest.Name, 
							userFromRequest.Email,userFromRequest.IsEmailVerified, hashedPassword)
							if qryErr != nil {
								log.Println("ERROR Function UserRegister: "+qryErr.Error())
								ctx.JSON(http.StatusInternalServerError, gin.H{
									"code": http.StatusInternalServerError,
									"success": false,
									"message": "Sorry my friend, something went wrong on our side. Our team is working on it. Please refresh the page or try again later.",
								})
							}
							rowsAffected, rowsAffectedErr := insertResults.RowsAffected()
							if rowsAffectedErr != nil {
								log.Fatal("ERROR Function UserRegister: ", rowsAffectedErr.Error())
								ctx.JSON(http.StatusInternalServerError, gin.H{
									"code": http.StatusInternalServerError,
									"success": false,
									"message": "Sorry my friend, something went wrong on our side. Our team is working on it. Please refresh the page or try again later.",
								})
							} else {
								if rowsAffected > 0 {
									log.Println("User registration successfull", userFromRequest.Email)
									timeNow := time.Now()
									_, _ = genzDB.ExecContext(ctx, "INSERT INTO user_verification_code VALUES(?, ?, ?);", verificationCode, userFromRequest.Email, timeNow)
									ctx.JSON(http.StatusOK, gin.H{
										"code": http.StatusOK,
										"success": true,
										"message": "User registration successfull.",
										"data": timeNow,
									})
								} else {
									ctx.JSON(http.StatusOK, gin.H{
										"code": http.StatusOK,
										"success": false,
										"message": "User registration failed.",
									})
								}
							}
						}
						
					}
				}
			}
		}

		

	}

	return gin.HandlerFunc(UserRegister)

}

func GetUserByIdHandler(genzDB *sql.DB) gin.HandlerFunc {

	GetUserById := func(ctx *gin.Context) {
		

		var xGenzToken string
		xGenzTokenArr, ok := ctx.Request.Header["X-Genz-Token"];
		if !ok {
			log.Println("Token not exists.")
			ctx.JSON(http.StatusOK, gin.H {
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry my friend, Invalid request. We'll fix it ASAP. Please refresh the page or try again later.",
			})
		} else {
			queryParams := ctx.Request.URL.Query()
			userIdFromReq := queryParams["userId"][0]
			xGenzToken = xGenzTokenArr[0]
			var user users.User
			if X_GENZ_TOKEN != xGenzToken {
				log.Println("ERROR Function GetUserById: Invalid security key", userIdFromReq)
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry my friend, Invalid security key. We'll fix it ASAP. Please refresh the page or try again later.",
				})
			} else {
				getUserResultsErr := genzDB.QueryRow("SELECT user_id, name, email, profile, is_email_verified FROM users WHERE user_id=?;", userIdFromReq).Scan(&user.UserId, &user.Name, &user.Email, &user.Profile, &user.IsEmailVerified)
				switch getUserResultsErr {
					case sql.ErrNoRows:
						log.Println("No rows were returned!", userIdFromReq)
						ctx.JSON(http.StatusOK, gin.H{
							"code": http.StatusOK,
							"success": false,
							"message": "User not found",
						})
					case nil:
						log.Println("User fetched for profile", userIdFromReq)
						ctx.JSON(http.StatusOK, gin.H{
							"code": http.StatusOK,
							"success": true,
							"data": user,
						})
					default:
						log.Println("ERROR Function GetUserById: "+getUserResultsErr.Error())
						ctx.JSON(http.StatusInternalServerError, gin.H{
							"code": http.StatusInternalServerError,
							"success": false,
							"message": "Sorry my friend, something went wrong on our side. Our team is working on it. Please refresh the page or try again later.",
						})
				}
				
			}
		}

	}

	return gin.HandlerFunc(GetUserById)
}

func EditUserNameHandler(genzDB *sql.DB) gin.HandlerFunc {
	
	EditUserName := func(ctx *gin.Context) {
		var xGenzToken string
		xGenzTokenArr, ok := ctx.Request.Header["X-Genz-Token"];
		var userFromRequest users.User
		if !ok {
			log.Println("Token not exists.")
			ctx.JSON(http.StatusOK, gin.H {
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry my friend, Invalid request. We'll fix it ASAP. Please refresh the page or try again later.",
			})
		} else {
			ctx.ShouldBindJSON(&userFromRequest)
			xGenzToken = xGenzTokenArr[0]
			var user users.User
			if X_GENZ_TOKEN != xGenzToken {
				log.Println("ERROR Function EditUserName: Invalid security key", userFromRequest.UserId)
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry my friend, Invalid security key. We'll fix it ASAP. Please refresh the page or try again later.",
				})
			} else {
				getUserResultsErr := genzDB.QueryRow("SELECT user_id FROM users WHERE user_id=?;", userFromRequest.UserId).Scan(&user.UserId)
				switch getUserResultsErr {
					case sql.ErrNoRows:
						log.Println("No rows were returned!", userFromRequest.UserId)
						ctx.JSON(http.StatusOK, gin.H{
							"code": http.StatusOK,
							"success": false,
							"message": "User not found",
						})
					case nil:
						sqlUpdateUserNameQuery := "UPDATE users SET name=? WHERE user_id=?;"
						_, updateQueryError := genzDB.Exec(sqlUpdateUserNameQuery, userFromRequest.Name, userFromRequest.UserId)
						if updateQueryError != nil {
							log.Println("ERROR function EditUserName: "+updateQueryError.Error())
							ctx.JSON(http.StatusOK, gin.H{
								"code": http.StatusOK,
								"success": true,
								"message": "Failed to update username.",
							})
						} else {
							ctx.JSON(http.StatusOK, gin.H{
								"code": http.StatusOK,
								"success": true,
								"message": "Username updated.",
							})
						}
					default:
						log.Println("ERROR Function EditUserName: "+getUserResultsErr.Error())
						ctx.JSON(http.StatusInternalServerError, gin.H{
							"code": http.StatusInternalServerError,
							"success": false,
							"message": "Sorry my friend, something went wrong on our side. Our team is working on it. Please refresh the page or try again later.",
						})
				}

			}

		}
	}

	return gin.HandlerFunc(EditUserName)
}

func generateUserId(genzDB *sql.DB) (string, error) {
	var user users.User
	var maxUserId int
	var finalUserId string

	query := "SELECT user_id FROM users;"
	queryResults, qryErr := genzDB.Query(query)
	if qryErr != nil {
		log.Println("ERROR Function generateUserId: "+qryErr.Error())
		return "", errors.New("Sorry my friend, something went wrong on our side. Our team is working on it. Please refresh the page or try again later.")
	}
	for queryResults.Next() {
		queryResultsErr := queryResults.Scan(&user.UserId)
		if queryResultsErr != nil {
			log.Println("ERROR Function generateUserId: "+queryResultsErr.Error())
			return "", errors.New("Sorry my friend, something went wrong on our side. Our team is working on it. Please refresh the page or try again later.")
		}
		userIdArr := strings.Split(user.UserId, "-")
		numUserId, atoiErr := strconv.Atoi(userIdArr[1])

		if atoiErr != nil {
			log.Println("ERROR Function generateUserId: "+atoiErr.Error())
			return "", errors.New("Sorry my friend, something went wrong on our side. Our team is working on it. Please refresh the page or try again later.")
		}

		if maxUserId <= numUserId {
			maxUserId = numUserId
		}

	}
	userIdString := strconv.Itoa(maxUserId+1)
	finalUserId = "GB-"+userIdString
	return finalUserId, nil
}


func checkUserExists(genzDB *sql.DB, email string) bool {
	var rowCount int
	query := "SELECT COUNT(*) FROM users WHERE email = ?"
	_ = genzDB.QueryRow(query, email).Scan(&rowCount)
	if rowCount > 0 {
		return true
	}
	return false
}