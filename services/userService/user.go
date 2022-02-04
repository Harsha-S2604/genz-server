package userService

import (
	"log"
	"database/sql"
	"net/http"
	"net/smtp"
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

type AboutYouStruct struct {
	AboutYou string	`json: "aboutYou"`
	UserId string 	`json: "userId"`
}

type RegisterStruct struct {
	UserId 				string `json: "userId"`
	Name				string `json: "name"`
	Email 				string `json: "email"`
	Password 			string `json: "password"`
	Profile				string `json: "profile"`
	VerificationCode 	string `jsin: "verificationCode"`
}

type ChangePasswordStruct struct {
	UserId 			string	`json: userId`
	OldPasswd		string	`json: oldPasswd`
	NewPasswd		string	`json: newPasswd`
	ConfirmPasswd	string	`json: confirmPasswd`
}

var (
	X_GENZ_TOKEN = "4439EA5BDBA8B179722265789D029477"
)

func ValidateUserLoginHandler(genzDB *sql.DB) gin.HandlerFunc {
	ValidateUserLogin := func(ctx *gin.Context) {
		isValidToken, msg := verification.VerifyInternalKey(ctx)
		if !isValidToken {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": msg,
			})
			return
		}

		var userFromRequest users.User
		user := new(users.User);
		resultsCount := 0		
		ctx.ShouldBindJSON(&userFromRequest)
		// validate email and hash the password
		isValidEmail, isValidEmailErr := validations.ValidateUserEmail(userFromRequest.Email)
		hashedPassword := hashing.HashUserPassword(userFromRequest.Password)
		if !isValidEmail && isValidEmailErr != nil {
			log.Println("ERROR Function ValidateUserLogin: ", isValidEmailErr.Error(), userFromRequest.Email)
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Email is not in a valid format.",
			})
			return
		}
		
		// Execute the query
		results, qryErr := genzDB.Query("SELECT email, password FROM users WHERE email = ?;", userFromRequest.Email)
		if qryErr != nil {
			log.Println("ERROR Function ValidateUserLogin: "+qryErr.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"success": false,
				"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
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
					"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
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
				"message": "You have not registered yet. Please register.",
			})
			return
		}
		// check if user credentials match or not and send appropriate messages
		if user.Email == userFromRequest.Email && user.Password == hashedPassword {
			var userData users.User
			log.Println("User credentials match",userFromRequest.Email)
			loggedinUser, qryMatchErr := genzDB.Query("SELECT user_id, name, email FROM users WHERE email = ?;", userFromRequest.Email)
			if qryMatchErr != nil {
				log.Println("ERROR Function ValidateUserLogin: "+qryMatchErr.Error())
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code": http.StatusInternalServerError,
					"success": false,
					"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
				})
				return
			}
			for loggedinUser.Next() {
				loggedinUserErr := loggedinUser.Scan(&userData.UserId, &userData.Name, &userData.Email)
				if loggedinUserErr != nil {
					log.Println("ERROR Function ValidateUserLogin: "+loggedinUserErr.Error())
					ctx.JSON(http.StatusInternalServerError, gin.H{
						"code": http.StatusInternalServerError,
						"success": false,
						"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
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
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"success": false,
			"message": "Incorrect email or password.",
		})
	}

	return gin.HandlerFunc(ValidateUserLogin)
}

func UserRegisterHandler(genzDB *sql.DB) gin.HandlerFunc {
	
	UserRegister := func(ctx *gin.Context) {
		isValidToken, msg := verification.VerifyInternalKey(ctx)
		if !isValidToken {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": msg,
			})
			return
		}
		var userFromRequest RegisterStruct
		ctx.ShouldBindJSON(&userFromRequest)
		isCodeVerified, message := verifyCode(genzDB, userFromRequest.Email, userFromRequest.VerificationCode)
		if !isCodeVerified {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": message,
			})
			return
		}

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
			return
		}
		hashedPassword := hashing.HashUserPassword(userFromRequest.Password)
		if !isValidEmail && isValidEmailErr != nil {
			log.Println("ERROR Function UserRegister: ", isValidEmailErr.Error(), userFromRequest.Email)
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Email is not in a valid format.",
			})
			return
		}
		userId, generateIdErr := generateUserId(genzDB)
		if generateIdErr != nil {
			log.Println("ERROR Function UserRegister: USER ID generation error", generateIdErr.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"success": false,
				"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
			})
		}
		// Execute the query					
		insertResults, qryErr := genzDB.ExecContext(ctx, "INSERT INTO users VALUES(?, ?, ?, ?, ?);", userId, userFromRequest.Name, 
		userFromRequest.Email, hashedPassword, "{\"about\": \"\", \"contact\": \"{}\"}")
		if qryErr != nil {
			log.Println("ERROR Function UserRegister: "+qryErr.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"success": false,
				"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
			})
			return
		}

		rowsAffected, rowsAffectedErr := insertResults.RowsAffected()
		if rowsAffectedErr != nil {
			log.Fatal("ERROR Function UserRegister: ", rowsAffectedErr.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"success": false,
				"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
			})
			return
		}
		if rowsAffected > 0 {
			log.Println("User registration successfull", userFromRequest.Email)
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": true,
				"message": "User registration successfull.",
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"success": false,
			"message": "User registration failed.",
		})

	}

	return gin.HandlerFunc(UserRegister)

}

func GetUserByIdHandler(genzDB *sql.DB) gin.HandlerFunc {

	GetUserById := func(ctx *gin.Context) {
		isValidToken, msg := verification.VerifyInternalKey(ctx)
		if !isValidToken {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": msg,
			})
			return
		}

		queryParams := ctx.Request.URL.Query()
		userIdFromReq := queryParams["userId"][0]
		var user users.User 
		getUserResultsErr := genzDB.QueryRow("SELECT user_id, name, email, profile FROM users WHERE user_id=?;", userIdFromReq).Scan(&user.UserId, &user.Name, &user.Email, &user.Profile)
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
					"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
				})
		}
	}

	return gin.HandlerFunc(GetUserById)
}

func CheckUserByEmailHandler(genzDB *sql.DB) gin.HandlerFunc {

	CheckUserByEmail := func(ctx *gin.Context) {
		
		isValidToken, msg := verification.VerifyInternalKey(ctx)
		if !isValidToken {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": msg,
			})
			return
		}
		var userFromRequest users.User
		ctx.ShouldBindJSON(&userFromRequest) 
		isExists := checkUserExists(genzDB, userFromRequest.Email)
		if isExists {
			log.Println("User already exist")
			ctx.JSON(http.StatusOK, gin.H {
				"code": http.StatusOK,
				"success": false,
				"message": "User already exist. Please sign in to continue.",
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H {
			"code": http.StatusOK,
			"success": true,
			"message": "",
		})
				 

	}

	return gin.HandlerFunc(CheckUserByEmail)
}

func SendVerificationCodeHandler(genzDB *sql.DB) gin.HandlerFunc {

	sendVerificationCode := func(ctx *gin.Context) {
		isValidToken, msg := verification.VerifyInternalKey(ctx)
		if !isValidToken {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": msg,
			})
			return
		}
		var userFromRequest users.User
		ctx.ShouldBindJSON(&userFromRequest)

		isUserExist := checkUserExistsForVerification(genzDB, userFromRequest.Email)
		if isUserExist {
			var codeCount int
			getCountQry := "SELECT code_sent_count FROM user_verification_code WHERE email=?;"
			_ = genzDB.QueryRow(getCountQry, userFromRequest.Email).Scan(&codeCount)
			if codeCount >= 2 {
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "You have exceeded the maximum number of attempts. Please try again later.",
				})
				return
			}

			// generate the verification code
			verificationCode, verificationCodeErr := verification.GenerateSixDigitCode()
			if verificationCodeErr != nil {
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry, something went wrong. Our team is working on it. Please, try again later.",
				})
				return
			}
			timeNow := time.Now()
			updateVerificationCodeQry := "UPDATE user_verification_code SET verification_code=?, code_sent_count=?, created_at=? WHERE email=?;"
			_, updateVerificationCodeQryErr := genzDB.Exec(updateVerificationCodeQry, verificationCode, codeCount + 1, timeNow, userFromRequest.Email)
			if updateVerificationCodeQryErr != nil {
				log.Println("Update verification code error: ", updateVerificationCodeQryErr.Error())
				ctx.JSON(http.StatusOK, gin.H {
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry, something went wrong. Our team is working on it. Please, try again later.",
				})
				return
			}
			isSent := sendVerificationCode(verificationCode, userFromRequest.Email)
			if isSent {
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": true,
					"message": "A new verification code has been sent to your email. Please check it.",
					"data": codeCount,
				})
				return
			} 
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry something went wrong. Our team is working on it. Please try again later.",
			})
			return
			
		} 
		timeNow := time.Now()
		verificationCode, verificationCodeErr := verification.GenerateSixDigitCode()
		if verificationCodeErr != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry, something went wrong. Our team is working on it. Please, try again later.",
			})
			return
		}
		_, _ = genzDB.Exec("INSERT INTO user_verification_code VALUES(?, ?, ?, ?);", verificationCode, userFromRequest.Email, timeNow, 0)
		isSent := sendVerificationCode(verificationCode, userFromRequest.Email)
		if isSent {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": true,
				"message": "verification code has been sent to your account.",
				"data": timeNow,
			})
			return
		} 
		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"success": false,
			"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
		})
					
		
	}

	return gin.HandlerFunc(sendVerificationCode)

}

func EditUserNameHandler(genzDB *sql.DB) gin.HandlerFunc {
	
	EditUserName := func(ctx *gin.Context) {
		isValidToken, msg := verification.VerifyInternalKey(ctx)
		if !isValidToken {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": msg,
			})
			return
		}
		var userFromRequest users.User
		ctx.ShouldBindJSON(&userFromRequest)
		var user users.User
			
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
					return
				} 
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": true,
					"message": "Username updated.",
				})
				
			default:
				log.Println("ERROR Function EditUserName: "+getUserResultsErr.Error())
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code": http.StatusInternalServerError,
					"success": false,
					"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
				})
		}		
	}

	return gin.HandlerFunc(EditUserName)
}


func EditAboutYouHandler(genzDB *sql.DB) gin.HandlerFunc {
	
	EditAboutYou := func(ctx *gin.Context) {

		isValidToken, msg := verification.VerifyInternalKey(ctx)
		if !isValidToken {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": msg,
			})
			return
		}

		var dataFromRequest AboutYouStruct
		ctx.ShouldBindJSON(&dataFromRequest)
		var user users.User
			

		getUserResultsErr := genzDB.QueryRow("SELECT user_id FROM users WHERE user_id=?;", dataFromRequest.UserId).Scan(&user.UserId)
		switch getUserResultsErr {
			case sql.ErrNoRows:
				log.Println("No rows were returned!", dataFromRequest.UserId)
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "User not found",
				})
			case nil:
				sqlUpdateAboutYouQuery := "UPDATE users SET profile=JSON_SET(profile, \"$.about\", ?) WHERE user_id=?"
				_, updateQueryError := genzDB.Exec(sqlUpdateAboutYouQuery, dataFromRequest.AboutYou, dataFromRequest.UserId)
				if updateQueryError != nil {
					log.Println("ERROR function EditAboutYou: "+updateQueryError.Error())
					ctx.JSON(http.StatusOK, gin.H{
						"code": http.StatusOK,
						"success": true,
						"message": "Failed to update.",
					})
					return
				}

				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": true,
					"message": "Updated successfully.",
				})
				
			default:
				log.Println("ERROR Function EditAboutYou: "+getUserResultsErr.Error())
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code": http.StatusInternalServerError,
					"success": false,
					"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
				})
		}

	}

	return gin.HandlerFunc(EditAboutYou)
}

func ChangePasswordHandler(genzDB *sql.DB) gin.HandlerFunc {

	ChangePassword := func(ctx *gin.Context) {
		isValidToken, msg := verification.VerifyInternalKey(ctx)
		if !isValidToken {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": msg,
			})
			return
		}

		var dataFromRequest ChangePasswordStruct
		ctx.ShouldBindJSON(&dataFromRequest)
		var user users.User

		getUserResultsErr := genzDB.QueryRow("SELECT password FROM users WHERE user_id=?;", dataFromRequest.UserId).Scan(&user.Password)
		switch getUserResultsErr {
			case sql.ErrNoRows:
				log.Println("No rows were returned!", dataFromRequest.UserId)
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "User not found",
				})
				return
			
			case nil:
				hashedOldPassword := hashing.HashUserPassword(dataFromRequest.OldPasswd)
				if hashedOldPassword != user.Password {
					log.Println("Invalid old password", dataFromRequest.UserId)
					ctx.JSON(http.StatusOK, gin.H{
						"code": http.StatusOK,
						"success": false,
						"message": "invalid old password",
					})
				} else if dataFromRequest.NewPasswd != dataFromRequest.ConfirmPasswd {
					log.Println("Password do not match", dataFromRequest.UserId)
					ctx.JSON(http.StatusOK, gin.H{
						"code": http.StatusOK,
						"success": false,
						"message": "Password do not match",
					})
					return
				}
				hashedNewPasswd := hashing.HashUserPassword(dataFromRequest.NewPasswd)
				updatePasswdQry := "UPDATE users SET password=? WHERE user_id=?"
				_, updatePasswdQryErr := genzDB.Exec(updatePasswdQry, hashedNewPasswd, dataFromRequest.UserId)
				if updatePasswdQryErr != nil {
					log.Println("Failed to update password", dataFromRequest.UserId)
					ctx.JSON(http.StatusOK, gin.H{
						"code": http.StatusOK,
						"success": false,
						"message": "Failed to update password. Please, try again later.",
					})
					return
				}

				log.Println("Password updated successfully", dataFromRequest.UserId)
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": true,
					"message": "Password updated successfully",
				})
					
				
			}
		}

	return gin.HandlerFunc(ChangePassword)
}

	
func generateUserId(genzDB *sql.DB) (string, error) {
	var user users.User
	var maxUserId int
	var finalUserId string

	query := "SELECT user_id FROM users;"
	queryResults, qryErr := genzDB.Query(query)
	if qryErr != nil {
		log.Println("ERROR Function generateUserId: "+qryErr.Error())
		return "", errors.New("Sorry, Something went wrong. Our team is working on it. Please try again later.")
	}
	for queryResults.Next() {
		queryResultsErr := queryResults.Scan(&user.UserId)
		if queryResultsErr != nil {
			log.Println("ERROR Function generateUserId: "+queryResultsErr.Error())
			return "", errors.New("Sorry, Something went wrong. Our team is working on it. Please try again later.")
		}
		userIdArr := strings.Split(user.UserId, "-")
		numUserId, atoiErr := strconv.Atoi(userIdArr[1])

		if atoiErr != nil {
			log.Println("ERROR Function generateUserId: "+atoiErr.Error())
			return "", errors.New("Sorry, Something went wrong. Our team is working on it. Please try again later.")
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


func checkUserExistsForVerification(genzDB *sql.DB, email string) bool {
	var rowCount int
	query := "SELECT COUNT(*) FROM user_verification_code WHERE email = ?"
	_ = genzDB.QueryRow(query, email).Scan(&rowCount)
	if rowCount > 0 {
		return true
	}
	return false
}


func sendVerificationCode(verificationCode string, email string) bool {
	from := "arix2604@gmail.com"
	fromPassword := "Zxc890@bnm123h4s26us20"
	toList := []string{email}
	host := "smtp.gmail.com"
	port := "587"

	msg :="To: "+email+"\r\n" +
	"Subject: Verify your email address!\r\n" +
	"\r\n" +
	"Hello " + 
	",\n\n We are happy that you are signed up for GenZ BlogZ. To Start exploring, please confirm your email address.\n\n Please use this verification code to complete your sign in: \n" +
	verificationCode + "\n\n Welcome to GenZ BlogZ!\n The Genz BlogZ Team"
	body := []byte(msg)
	auth := smtp.PlainAuth("", from, fromPassword, host)
	log.Println("Sending mail to", email)

	emailErr := smtp.SendMail(host+":"+port, auth, from, toList, body)
	if emailErr != nil {
		log.Println("Sending email error", emailErr.Error())
		return false
	}

	return true
}

func verifyCode(genzDB *sql.DB, email string, verificationCode string) (bool, string) {
	var userVerificationCode users.UserVerificationCode
	getUserResultsErr := genzDB.QueryRow("SELECT * FROM user_verification_code WHERE email=?;", email).Scan(&userVerificationCode.VerificationCode, &userVerificationCode.Email, &userVerificationCode.CreatedAt, &userVerificationCode.CodeSentCount)
	timeNow := time.Now()
	timeMinus := timeNow.Add(-1 * time.Hour)
	if getUserResultsErr != nil {
		return false, "Sorry, something went wrong. Our team is working on it. Please try again later."
	} else if userVerificationCode.VerificationCode == verificationCode {
		createdAt := userVerificationCode.CreatedAt
		if !(createdAt.After(timeMinus)) {
			return false, "verification code expired"
		}
	} else if userVerificationCode.VerificationCode != verificationCode{
		return false, "invalid verification code"
	}

	return true, ""
}