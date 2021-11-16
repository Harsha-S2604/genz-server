package blogService

import (
	"log"
	"database/sql"
	"net/http"
	"time"
	"strconv"

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
			err := ctx.ShouldBindJSON(&blog)
			if(err != nil) {
				log.Println("ERROR FUNCTION ADD BLOG:", err.Error())
			}
			timeNow := time.Now()
			insertQueryResult, insertQueryError := genzDB.ExecContext(ctx, "INSERT INTO blog(blog_title, blog_description, blog_content, blog_created_at, blog_last_updated_at, blog_is_draft, blog_total_views, blog_total_likes, email) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?);", blog.BlogTitle, blog.BlogDescription, blog.BlogContent, timeNow, timeNow, blog.BlogIsDraft, blog.BlogTotalViews, blog.BlogTotalLikes, blog.User.Email)
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

func GetBlogByIDHandler(genzDB *sql.DB) gin.HandlerFunc {

	GetBlogByID := func(ctx *gin.Context) {
		var xGenzToken string
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
				log.Println("Invalid token.")
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry my friend, Invalid token. We'll fix it ASAP. Please refresh the page or try again later.",
				})
			} else {
				queryParams := ctx.Request.URL.Query()
				blogIdFromReq := queryParams["blogId"][0]
				isGetDraft := queryParams["get_draft"][0]
				var blog blogs.Blog

				getResultQuery := genzDB.QueryRow("SELECT * FROM blog WHERE blog_id=? AND blog_is_draft=?", blogIdFromReq, isGetDraft).Scan(&blog.BlogID, &blog.BlogTitle, &blog.BlogDescription, &blog.BlogContent, &blog.BlogCreatedAt, &blog.BlogLastUpdatedAt, &blog.BlogIsDraft, &blog.BlogTotalViews, &blog.BlogTotalLikes, &blog.User.Email)
				switch getResultQuery {
					case sql.ErrNoRows:
						log.Println("No rows were returned!", blogIdFromReq)
						ctx.JSON(http.StatusOK, gin.H{
							"code": http.StatusOK,
							"success": false,
							"message": "Blog not found",
						})
					case nil:
						log.Println("Blog fetched.", blogIdFromReq)
						ctx.JSON(http.StatusOK, gin.H{
							"code": http.StatusOK,
							"success": true,
							"data": blog,
						})
					default:
						log.Println("ERROR Function GetBlogById: "+getResultQuery.Error())
						ctx.JSON(http.StatusInternalServerError, gin.H{
							"code": http.StatusInternalServerError,
							"success": false,
							"message": "Sorry my friend, something went wrong on our side. Our team is working on it. Please refresh the page or try again later.",
						}) 
				}
			}
		}
	}

	return gin.HandlerFunc(GetBlogByID)

}

func GetAllBlogsHandler(genzDB *sql.DB) gin.HandlerFunc {

	GetAllBlog := func(ctx *gin.Context) {
		var xGenzToken string
		xGenzTokenArr, ok := ctx.Request.Header["X-Genz-Token"];

		if !ok {
			log.Println("TOKEN NOT EXISTS.")
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry my friend, Invalid request. We'll fix it ASAP. Please refresh the page or try again later.",
			})
		} else {
			xGenzToken = xGenzTokenArr[0]
			if xGenzToken != X_GENZ_TOKEN {
				log.Println("Invalid token.")
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry my friend, Invalid token. We'll fix it ASAP. Please refresh the page or try again later.",
				})
			} else {
				queryParams := ctx.Request.URL.Query()
				emailFromReq := queryParams["email"][0]
				isGetDraft, _ := strconv.ParseBool(queryParams["get_draft"][0])
				var blogsArr []blogs.Blog

				blogRows, blogRowsErr := genzDB.Query("SELECT * FROM blog WHERE email=? AND blog_is_draft=?", emailFromReq, isGetDraft)
				if blogRowsErr != nil {
					log.Println("ERROR function GetAllBlog: ", blogRowsErr.Error())
					ctx.JSON(http.StatusOK, gin.H{
						"code": http.StatusOK,
						"success": false,
						"message": "Sorry my friend, Something went wrong. We'll fix it ASAP. Please refresh the page or try again later.",
					})
				} else {
					for blogRows.Next() {
						var blogObj blogs.Blog
						if blogErr := blogRows.Scan(&blogObj.BlogID, &blogObj.BlogTitle, &blogObj.BlogDescription, &blogObj.BlogContent, &blogObj.BlogCreatedAt, &blogObj.BlogLastUpdatedAt, &blogObj.BlogIsDraft, &blogObj.BlogTotalViews, &blogObj.BlogTotalLikes, &blogObj.User.Email); blogErr != nil {
							log.Println("ERROR function GetAllBlog blogErr:", blogErr.Error())
							ctx.JSON(http.StatusOK, gin.H{
								"code": http.StatusOK,
								"success": false,
								"message": "Sorry my friend, Something went wrong. We'll fix it ASAP. Please refresh the page or try again later.",
							})
							return
						}

						blogsArr = append(blogsArr, blogObj)
					}

					ctx.JSON(http.StatusOK, gin.H{
						"code": http.StatusOK,
						"success": true,
						"message": "Blog fetched.",
						"data": blogsArr,
					})
				}


			}
		}
	}

	return gin.HandlerFunc(GetAllBlog)

}

func DeleteBlogHandler(genzDB *sql.DB) gin.HandlerFunc {

	DeleteBlog := func(ctx *gin.Context) {
		var xGenzToken string
		xGenzTokenArr, ok := ctx.Request.Header["X-Genz-Token"]

		if !ok {
			log.Println("TOKEN NOT EXISTS.")
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry my friend, Invalid request. We'll fix it ASAP. Please refresh the page or try again later.",
			})
		} else {
			xGenzToken = xGenzTokenArr[0]
			if xGenzToken != X_GENZ_TOKEN {
				log.Println("Invalid token.")
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry my friend, Invalid token. We'll fix it ASAP. Please refresh the page or try again later.",
				})
			} else {
				queryParams := ctx.Request.URL.Query()
				blogIdFromReq := queryParams["blogId"][0]
				emailFromReq := queryParams["email"][0]

				delRes, delErr := genzDB.Exec("DELETE FROM blog WHERE blog_id=? AND email=?", blogIdFromReq, emailFromReq)
				if delErr != nil {
					log.Println("ERROR function DeleteBlog: ", delErr.Error())
					ctx.JSON(http.StatusOK, gin.H{
						"code": http.StatusOK,
						"success": false,
						"message": "Sorry my friend, Something went wrong. We'll fix it ASAP. Please refresh the page or try again later.",
					})
				} else {
					count, countErr := delRes.RowsAffected()
					if countErr != nil {
						log.Println("ERROR function DeleteBlog: ", countErr.Error())
						ctx.JSON(http.StatusOK, gin.H{
							"code": http.StatusOK,
							"success": false,
							"message": "Sorry my friend, Something went wrong. We'll fix it ASAP. Please refresh the page or try again later.",
						})
					} else {
						if count > 0 {
							log.Println("DELETED SUCCESSFULLY", blogIdFromReq)
							ctx.JSON(http.StatusOK, gin.H{
								"code": http.StatusOK,
								"success": true,
								"message": "Successfully deleted the blog.",
							})
						} else {
							log.Println("DELETION FAILED", blogIdFromReq)
							ctx.JSON(http.StatusOK, gin.H{
								"code": http.StatusOK,
								"success": false,
								"message": "Blog not found. Please refresh the page or try again later.",
							})
						}
					}
				}
			}
		}
	}

	return gin.HandlerFunc(DeleteBlog)

}

func FetchRecentArticlesHandler(genzDB *sql.DB) gin.HandlerFunc {

	FetchRecentArticles := func(ctx *gin.Context) {
		var xGenzToken string
		xGenzTokenArr, ok := ctx.Request.Header["X-Genz-Token"]

		if !ok {
			log.Println("TOKEN NOT EXISTS.")
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry my friend, Invalid request. We'll fix it ASAP. Please refresh the page or try again later.",
			})
		} else {
			xGenzToken = xGenzTokenArr[0]
			if xGenzToken != X_GENZ_TOKEN {
				log.Println("Invalid token.")
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry my friend, Invalid token. We'll fix it ASAP. Please refresh the page or try again later.",
				})
			} else {
				recentArticlesRows, recentArticlesRowsErr := genzDB.Query("SELECT * FROM blog WHERE blog_is_draft=false ORDER BY blog_created_at DESC LIMIT 5")
				if recentArticlesRowsErr != nil {
					log.Println("ERROR function FetchRecentArticles: ", recentArticlesRowsErr.Error())
					ctx.JSON(http.StatusOK, gin.H{
						"code": http.StatusOK,
						"success": false,
						"message": "Sorry my friend, Something went wrong. We'll fix it ASAP. Please refresh the page or try again later.",
					})
				} else {
					var recentArticlesArr []blogs.Blog
					for recentArticlesRows.Next() {
						var blogObj blogs.Blog
						if blogErr := recentArticlesRows.Scan(&blogObj.BlogID, &blogObj.BlogTitle, &blogObj.BlogDescription, &blogObj.BlogContent, &blogObj.BlogCreatedAt, &blogObj.BlogLastUpdatedAt, &blogObj.BlogIsDraft, &blogObj.BlogTotalViews, &blogObj.BlogTotalLikes, &blogObj.User.Email); blogErr != nil {
							log.Println("ERROR function FetchRecentArticles:", blogErr.Error())
							ctx.JSON(http.StatusOK, gin.H{
								"code": http.StatusOK,
								"success": false,
								"message": "Sorry my friend, Something went wrong. We'll fix it ASAP. Please refresh the page or try again later.",
							})
							return
						}
						recentArticlesArr = append(recentArticlesArr, blogObj)
					}

					ctx.JSON(http.StatusOK, gin.H{
						"code": http.StatusOK,
						"success": true,
						"message": "recent articles fetched.",
						"data": recentArticlesArr,
					})

				}
			}
		}
	}

	return gin.HandlerFunc(FetchRecentArticles)
}

func AddSavedBlogsHandler(genzDB *sql.DB) gin.HandlerFunc {

	AddSavedBlogs := func(ctx *gin.Context) {

		var xGenzToken string
		xGenzTokenArr, ok := ctx.Request.Header["X-Genz-Token"]

		if !ok {
			log.Println("TOKEN NOT EXISTS.")
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry my friend, Invalid request. We'll fix it ASAP. Please refresh the page or try again later.",
			})
		} else {
			xGenzToken = xGenzTokenArr[0]
			if xGenzToken != X_GENZ_TOKEN {
				log.Println("Invalid token.")
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry my friend, Invalid token. We'll fix it ASAP. Please refresh the page or try again later.",
				})
			} else {
				var savedBlogs blogs.SavedBlogs
				savedBlogBindingErr := ctx.ShouldBindJSON(&savedBlogs)
				if savedBlogBindingErr != nil {
					log.Fatal("ERROR function AddFavorites:", savedBlogBindingErr.Error())
					ctx.JSON(http.StatusOK, gin.H{
						"code": http.StatusOK,
						"success": false,
						"message": "Sorry my friend, Something went wrong. We'll fix it ASAP. Please refresh the page or try again later.",
					})
				} else {
					blogId, email := savedBlogs.Blog.BlogID, savedBlogs.User.Email
					query := "INSERT INTO saved_blogs VALUE(?, ?)"
					insertQueryResult, insertQueryError := genzDB.ExecContext(ctx, query, blogId, email)
					if insertQueryError != nil {
						log.Fatal("ERROR function AddFavorites", insertQueryError)
						ctx.JSON(http.StatusOK, gin.H{
							"code": http.StatusOK,
							"success": false,
							"message": "Sorry my friend, Error saving the blogs to the favorites. Please refresh the page or try again later.",
						})	
					} else {
						rowsAffected, rowsAffectedErr := insertQueryResult.RowsAffected()
						if rowsAffectedErr != nil {
							log.Fatal("ERROR function AddFavorites", rowsAffectedErr.Error())
							ctx.JSON(http.StatusOK, gin.H{
								"code": http.StatusOK,
								"success": false,
								"message": "Sorry my friend, Something went wrong. We'll fix it ASAP. Please refresh the page or try again later.",
							})
						} else {
							if rowsAffected > 0 {
								ctx.JSON(http.StatusOK, gin.H{
									"code": http.StatusOK,
									"success": true,
									"message": "Blog added to favorites.",
								})
							} else {
								log.Fatal("Failed to add the blog to favorites.")
								ctx.JSON(http.StatusOK, gin.H{
									"code": http.StatusOK,
									"success": false,
									"message": "Failed to add the blog to favorites. Please refresh the page or try again later.",
								})
							}
						}
					}
					
				}
			}
		}

	}

	return gin.HandlerFunc(AddSavedBlogs)
}