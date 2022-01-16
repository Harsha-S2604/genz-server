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

	userAPIRouter := router.Group("api/v1/users") 
	{
		// validate user login
		userAPIRouter.POST("/login", userService.ValidateUserLoginHandler(genzDB))

		// register user
		userAPIRouter.POST("/register", userService.UserRegisterHandler(genzDB))

		// get user by email
		userAPIRouter.POST("/check/email", userService.CheckUserByEmailHandler(genzDB))

		// get user details by id
		userAPIRouter.GET("/profile", userService.GetUserByIdHandler(genzDB))

		// edit user profile
		userAPIRouter.POST("/edit/username", userService.EditUserNameHandler(genzDB))

		userAPIRouter.POST("/edit/aboutyou", userService.EditAboutYouHandler(genzDB))

		userAPIRouter.POST("/change-passwd", userService.ChangePasswordHandler(genzDB))

		// send verification code
		userAPIRouter.POST("/verify/send", userService.SendVerificationCodeHandler(genzDB))

	}

	blogAPIRouter := router.Group("api/v1/blogs")
	{
		// add blog
		blogAPIRouter.POST("/add", blogService.AddBlogHandler(genzDB))

		// get blog details by id
		blogAPIRouter.GET("/blog", blogService.GetBlogByIDHandler(genzDB))

		// get all blogs
		blogAPIRouter.GET("/allBlogs", blogService.GetAllBlogsHandler(genzDB))

		// delete blog
		blogAPIRouter.POST("/remove", blogService.DeleteBlogHandler(genzDB))

		// recent blog
		blogAPIRouter.GET("/recent", blogService.FetchRecentArticlesHandler(genzDB))

		// add favorites blog
		blogAPIRouter.GET("/saved", blogService.AddSavedBlogsHandler(genzDB))

	}

	return router
}
