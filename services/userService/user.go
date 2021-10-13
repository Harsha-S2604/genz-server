package userService

import (
	"log"
	"database/sql"
	"net/http"

	"github.com/Harsha-S2604/genz-server/models/users"
	"github.com/Harsha-S2604/genz-server/validations"
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
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"success": false,
				"message": "Sorry my friend, Invalid security key. Please refresh the page or try again later.",
			})
			return
		} else {
			xGenzToken = xGenzTokenArr[0]
		} 

		if X_GENZ_TOKEN != xGenzToken {
			log.Println("ERROR Function ValidateUserLogin: Invalid security key", userFromRequest.Email)
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"success": false,
				"message": "Sorry my friend, Invalid security key. Please refresh the page or try again later.",
			})
			return
		}
		ctx.ShouldBindJSON(&userFromRequest)

		isValidEmail, err := validations.ValidateUserEmail(userFromRequest.Email)

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
			if user.Email == userFromRequest.Email && user.Password == userFromRequest.Password {
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

// func UserRegisterHandler(genzDB *sql.DB) gin.HandlerFunc {
	
// 	UserRegister := func(ctx *gin.Context) {

// 	}
// }