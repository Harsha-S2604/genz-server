package routes

import (
	"database/sql"

	"github.com/Harsha-S2604/genz-server/services/userService"
	"github.com/Harsha-S2604/genz-server/services/blogService"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

)

func SetupRouter(genzDB *sql.DB) *gin.Engine{
	
	router := gin.Default()
	config := cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowHeaders:     []string{"Origin", "x-genz-token", "content-type"},
	}
	router.Use(cors.New(config))

	userAPIRouter := router.Group("genz-server/user-api") 
	{
		// validate user login
		userAPIRouter.POST("/login", userService.ValidateUserLoginHandler(genzDB))

		// register user
		userAPIRouter.POST("/register", userService.UserRegisterHandler(genzDB))

		// get user details by id
		userAPIRouter.GET("/fetch-profile", userService.GetUserByIdHandler(genzDB))

	}

	blogAPIRouter := router.Group("genz-server/blog-api")
	{
		blogAPIRouter.POST("/add-blog", blogService.AddBlogHandler(genzDB))
	}

	return router
}