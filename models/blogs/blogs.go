package blogs

import (

	"github.com/Harsha-S2604/genz-server/models/users"
)

type Blog struct {
	BlogID 			  int64 		`json: blogId`
	BlogTitle 		  string 		`json: blogTitle`
	BlogDescription   string 		`json: blogDescription`
	BlogContent 	  string 		`json: blogContent`
	BlogCreatedAt	  []uint8 		`json: blogCreatedAt`
	BlogLastUpdatedAt []uint8 		`json: blogUpdatedAt`
	BlogIsDraft		  bool			`json: blogIsDraft`
	BlogTotalViews	  int64			`json: blogTotalViews`
	BlogTotalLikes	  int64			`json: blogTotalLikes`
	User			  users.User	`json: user`
}

type SavedBlogs struct {
	Blog	Blog 	   `json: blog`
	User	users.User `json: user`
}