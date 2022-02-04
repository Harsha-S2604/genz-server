package blogService

import (
	"log"
	"database/sql"
	"net/http"
	"time"
	"strconv"
	"mime/multipart"

	"github.com/Harsha-S2604/genz-server/models/blogs"
	"github.com/Harsha-S2604/genz-server/utilities/verification"
	"github.com/Harsha-S2604/genz-server/services/cloudservice/aws/s3"


	"github.com/gin-gonic/gin"
)

type ImageForm struct {
    Image *multipart.FileHeader `form:"upload" binding:"required"`
}


func AddBlogHandler(genzDB *sql.DB) gin.HandlerFunc {

	AddBlog := func(ctx *gin.Context) {
		var blog blogs.Blog
		isValidToken, msg := verification.VerifyInternalKey(ctx)
		if !isValidToken {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": msg,
			})
			return
		}
			
		bindErr := ctx.ShouldBindJSON(&blog)
		if(bindErr != nil) {
			log.Println("Bind error:", bindErr.Error())
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry, Something went wrong. Our team is working on it. Please refresh the page or try again later.",
			})
			return
		}
		
		var id int
		tx, beginErr := genzDB.Begin()
		if beginErr != nil {
			log.Println("Begin error:", beginErr.Error())
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry, Something went wrong. Our team is working on it. Please refresh the page or try again later.",
			})
			return
		}

		{
			timeNow := time.Now()
			stmt, stmtErr := tx.Prepare("INSERT INTO blog(blog_title, blog_description, blog_content, blog_created_at, blog_last_updated_at, blog_is_draft, blog_total_views, blog_total_likes, blog_image, email) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING blog_id;")
			if stmtErr != nil {
				log.Println("Statement error:", stmtErr.Error())
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry, Something went wrong. Our team is working on it. Please refresh the page or try again later.",
				})
				return
			}
			
			defer stmt.Close()

			stmtErr = stmt.QueryRow(
				blog.BlogTitle, 
				blog.BlogDescription, 
				blog.BlogContent, 
				timeNow, 
				timeNow, 
				blog.BlogIsDraft, 
				blog.BlogTotalViews, 
				blog.BlogTotalLikes,
				blog.BlogImage, 
				blog.User.Email,
			).Scan(&id)

			if stmtErr != nil {
				log.Println("Statement error:", stmtErr.Error())
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry, Something went wrong. Our team is working on it. Please refresh the page or try again later.",
				})
				return
			}
		}

		{

			commitErr := tx.Commit()
			if commitErr != nil {
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry, Something went wrong. Our team is working on it. Please refresh the page or try again later.",
				})
				return
			}

		}

		if id > 0 {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": true,
				"blog_id": id,
				"message": "Blog added successfully",
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"success": false,
			"message": "Failed to publish the blog. Something went wrong. Our team is working on it. Please try again later.",
		})
		

	}
	return gin.HandlerFunc(AddBlog)
}

func GetBlogByIDHandler(genzDB *sql.DB) gin.HandlerFunc {

	GetBlogByID := func(ctx *gin.Context) {
		
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
		blogIdFromReq := queryParams["blogId"][0]
		isGetDraft := queryParams["get_draft"][0]
		var blog blogs.Blog

		getResultQuery := genzDB.QueryRow("SELECT * FROM blog WHERE blog_id=? AND blog_is_draft=?", blogIdFromReq, isGetDraft).Scan(&blog.BlogID, &blog.BlogTitle, &blog.BlogDescription, &blog.BlogContent, &blog.BlogCreatedAt, &blog.BlogLastUpdatedAt, &blog.BlogIsDraft, &blog.BlogTotalViews, &blog.BlogTotalLikes, &blog.BlogImage, &blog.User.Email)
		switch getResultQuery {
			case sql.ErrNoRows:
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Blog not found",
				})
			case nil:
				blogIdStr := strconv.Itoa(blog.BlogID)
				blogImageObj, blogImageObjErr := s3.GetObjectFromS3(blogIdStr)
				// fmt.Println(blogImageObjErr.Error())
				if blogImageObjErr == nil {
					blog.BlogImage = blogImageObj
				}
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

	return gin.HandlerFunc(GetBlogByID)

}

func GetAllBlogsHandler(genzDB *sql.DB) gin.HandlerFunc {

	GetAllBlog := func(ctx *gin.Context) {
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
		emailFromReq := queryParams["email"][0]
		isGetDraft, _ := strconv.ParseBool(queryParams["get_draft"][0])
		var blogsArr []blogs.Blog

		blogRows, blogRowsErr := genzDB.Query("SELECT * FROM blog WHERE email=? AND blog_is_draft=?", emailFromReq, isGetDraft)
		if blogRowsErr != nil {
			log.Println("ERROR function GetAllBlog: ", blogRowsErr.Error())
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
			})
			return
		} 
		for blogRows.Next() {
			var blogObj blogs.Blog
			if blogErr := blogRows.Scan(&blogObj.BlogID, &blogObj.BlogTitle, &blogObj.BlogDescription, &blogObj.BlogContent, &blogObj.BlogCreatedAt, &blogObj.BlogLastUpdatedAt, &blogObj.BlogIsDraft, &blogObj.BlogTotalViews, &blogObj.BlogTotalLikes, &blogObj.BlogImage, &blogObj.User.Email); blogErr != nil {
				log.Println("ERROR function GetAllBlog blogErr:", blogErr.Error())
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
				})
				return
			}
			blogIdStr := strconv.Itoa(blogObj.BlogID)
			blogImageObj, blogImageObjErr := s3.GetObjectFromS3(blogIdStr)
			if blogImageObjErr == nil {
				blogObj.BlogImage = blogImageObj
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

	return gin.HandlerFunc(GetAllBlog)

}

func DeleteBlogHandler(genzDB *sql.DB) gin.HandlerFunc {

	DeleteBlog := func(ctx *gin.Context) {
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
		blogIdFromReq := queryParams["blogId"][0]
		emailFromReq := queryParams["email"][0]

		delRes, delErr := genzDB.Exec("DELETE FROM blog WHERE blog_id=? AND email=?", blogIdFromReq, emailFromReq)
		if delErr != nil {
			log.Println("ERROR function DeleteBlog: ", delErr.Error())
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
			})
			return
		}
		count, countErr := delRes.RowsAffected()
		if countErr != nil {
			log.Println("ERROR function DeleteBlog: ", countErr.Error())
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
			})
			return
		} 
		if count > 0 {
			log.Println("DELETED SUCCESSFULLY", blogIdFromReq)
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": true,
				"message": "Successfully deleted the blog.",
			})
			return
		}
		log.Println("DELETION FAILED", blogIdFromReq)
		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"success": false,
			"message": "Blog not found. Please refresh the page or try again later.",
		})
		
		
		
			
		
	}

	return gin.HandlerFunc(DeleteBlog)

}

func FetchRecentArticlesHandler(genzDB *sql.DB) gin.HandlerFunc {

	FetchRecentArticles := func(ctx *gin.Context) {
		isValidToken, msg := verification.VerifyInternalKey(ctx)
		if !isValidToken {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": msg,
			})
			return
		}

		recentArticlesRows, recentArticlesRowsErr := genzDB.Query("SELECT * FROM blog WHERE blog_is_draft=false ORDER BY blog_created_at DESC LIMIT 5")
		if recentArticlesRowsErr != nil {
			log.Println("ERROR function FetchRecentArticles: ", recentArticlesRowsErr.Error())
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
			})
			return
		}

		var recentArticlesArr []blogs.Blog
		for recentArticlesRows.Next() {
			var blogObj blogs.Blog
			if blogErr := recentArticlesRows.Scan(&blogObj.BlogID, &blogObj.BlogTitle, &blogObj.BlogDescription, &blogObj.BlogContent, &blogObj.BlogCreatedAt, &blogObj.BlogLastUpdatedAt, &blogObj.BlogIsDraft, &blogObj.BlogTotalViews, &blogObj.BlogTotalLikes, &blogObj.BlogImage, &blogObj.User.Email); blogErr != nil {
				log.Println("ERROR function FetchRecentArticles:", blogErr.Error())
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
				})
				return
			}
			blogIdStr := strconv.Itoa(blogObj.BlogID)
			blogImageObj, blogImageObjErr := s3.GetObjectFromS3(blogIdStr)
			if blogImageObjErr == nil {
				blogObj.BlogImage = blogImageObj
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

	return gin.HandlerFunc(FetchRecentArticles)
}

func AddSavedBlogsHandler(genzDB *sql.DB) gin.HandlerFunc {

	AddSavedBlogs := func(ctx *gin.Context) {

		isValidToken, msg := verification.VerifyInternalKey(ctx)
		if !isValidToken {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": msg,
			})
			return
		}

		var savedBlogs blogs.SavedBlogs
		savedBlogBindingErr := ctx.ShouldBindJSON(&savedBlogs)
		if savedBlogBindingErr != nil {
			log.Fatal("ERROR function AddFavorites:", savedBlogBindingErr.Error())
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
			})
			return
		}
		blogId, email := savedBlogs.Blog.BlogID, savedBlogs.User.Email
		query := "INSERT INTO saved_blogs VALUE(?, ?)"
		insertQueryResult, insertQueryError := genzDB.ExecContext(ctx, query, blogId, email)
		if insertQueryError != nil {
			log.Fatal("ERROR function AddFavorites", insertQueryError)
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
			})
			return	
		}

		rowsAffected, rowsAffectedErr := insertQueryResult.RowsAffected()
		if rowsAffectedErr != nil {
			log.Fatal("ERROR function AddFavorites", rowsAffectedErr.Error())
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
			})
			return
		}

		if rowsAffected > 0 {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": true,
				"message": "Blog added to favorites.",
			})
			return
		}

		log.Fatal("Failed to add the blog to favorites.")
		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"success": false,
			"message": "Sorry, Something went wrong. Our team is working on it. Please try again later.",
		})
		return
	}

	return gin.HandlerFunc(AddSavedBlogs)
}

func UploadStoryImageHandler(genzDB *sql.DB) gin.HandlerFunc {

	uploadStoryImage := func(ctx *gin.Context) {

		isValidToken, msg := verification.VerifyInternalKey(ctx)
		if !isValidToken {
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": msg,
			})
			return
		}

		form, formFileErr := ctx.MultipartForm()
		if formFileErr != nil {
			log.Println("Invalid Form File", formFileErr.Error())
			ctx.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"success": false,
				"message": "invalid file type",
			})
			return	
		}
		formFile := form.File["image"]
		blogId := form.Value["blogId"]
		userId := form.Value["userId"]
		if !(len(formFile) == 0) {
			isUploadedToS3, isUploadedToS3Err := s3.UploadImageToS3(formFile[0], userId[0], blogId[0])
			if isUploadedToS3Err != nil {
				log.Println("Failed To Upload to S3", isUploadedToS3Err.Error())
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": false,
					"message": "Sorry something went wrong. Our team is working on it. Please try again later.",
				})
				return
			}
			if isUploadedToS3 {
				ctx.JSON(http.StatusOK, gin.H{
					"code": http.StatusOK,
					"success": true,
					"message": "Successfully uploaded",
				})
				return
			}
		}

		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"success": false,
			"message": "Please provide the story image",
		})
			

	}

	return gin.HandlerFunc(uploadStoryImage)
}