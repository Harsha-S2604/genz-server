package blogs

import (
	"github.com/Harsha-S2604/genz-server/models/users"
)

type Blog struct {
	BlogID 			int64 		`json: blogId`
	BlogTitle 		string 		`json: blogTitle`
	BlogDescription string 		`json: blogDescription`
	BlogContent 	string 		`json: blogContent`
	User			users.User	`json: user`
}