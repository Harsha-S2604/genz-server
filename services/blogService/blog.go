package blogService

import (
	"log"
	"database/sql"
	"net/http"

	"github.com/Harsha-S2604/genz-server/models/blogs"

	"github.com/gin-gonic/gin"
)

var (
	X_GENZ_TOKEN = "4439EA5BDBA8B179722265789D029477"
)

func AddBlogHandler(genzDB *sql.DB) gin.HandlerFunc {

	AddBlog := func(ctx *gin.Context) {
		log.Println("ADDING THE BLOG...")
		var xGenzToken string
		var blog blogs.Blog
		xGenzTokenArr, ok := ctx.Request.Header["X-Genz-Token"];
		
		if !ok {
			log.Println("Token not exists.")
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry my friend, Invalid request. We'll fix it ASAP. Please refresh the page or try again later.",
			})
		} else {
			xGenzToken = xGenzTokenArr[0]
			if xGenzToken != X_GENZ_TOKEN {
				log.Println("INVALID TOKEN.")
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry my friend, Invalid security token. We'll fix it ASAP. Please refresh the page or try again later.",
				})
				return
			}
			ctx.ShouldBindJSON(&blog)
			insertQueryResult, insertQueryError := genzDB.ExecContext(ctx, "INSERT INTO blog(blog_title, blog_description, blog_content, email) VALUES(?, ?, ?, ?);", blog.BlogTitle, blog.BlogDescription, blog.BlogContent, blog.User.Email)
			if insertQueryError != nil {
				log.Println("ERROR function AddBlog:", insertQueryError.Error())
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry my friend, Something went wrong. We'll fix it ASAP. Please refresh the page or try again later.",
				})
			} else {
				rowsAffected, rowsAffectedErr := insertQueryResult.RowsAffected()
				if rowsAffectedErr != nil {
					log.Println("ERROR function AddBlog rowsAffectedErr:", rowsAffectedErr.Error())
					ctx.JSON(http.StatusOK, gin.H{
						"code": http.StatusOK,
						"success": false,
						"message": "Sorry my friend, Something went wrong. We'll fix it ASAP. Please refresh the page or try again later.",
					})
				} else {
					if rowsAffected > 0 {
						log.Println("Blog added successfully.")
						ctx.JSON(http.StatusOK, gin.H{
							"code": http.StatusOK,
							"success": true,
							"message": "Blog Posted.",
						})
					} else {
						log.Println("Failed to post the blog.")
						ctx.JSON(http.StatusOK, gin.H{
							"code": http.StatusOK,
							"success": false,
							"message": "Failed to post the blog. Please try again later or contact us.",
						})
					}
				}
			}
		}

	}
	return gin.HandlerFunc(AddBlog)
}