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

		// edit user profile
		userAPIRouter.POST("/edit-username", userService.EditUserNameHandler(genzDB))

		userAPIRouter.POST("/edit-aboutyou", userService.EditAboutYouHandler(genzDB))

		userAPIRouter.POST("/change-passwd", userService.ChangePasswordHandler(genzDB))

	}

	blogAPIRouter := router.Group("genz-server/blog-api")
	{
		// add blog
		blogAPIRouter.POST("/add-blog", blogService.AddBlogHandler(genzDB))

		// get blog details by id
		blogAPIRouter.GET("/fetch-blog", blogService.GetBlogByIDHandler(genzDB))

		// get all blogs
		blogAPIRouter.GET("/fetch-blogs", blogService.GetAllBlogsHandler(genzDB))

		// delete blog
		blogAPIRouter.POST("/remove-blog", blogService.DeleteBlogHandler(genzDB))

		// recent blog
		blogAPIRouter.GET("/recent-blogs", blogService.FetchRecentArticlesHandler(genzDB))

		// add favorites blog
		blogAPIRouter.GET("/saved-blog", blogService.AddSavedBlogsHandler(genzDB))

	}

	return router
}