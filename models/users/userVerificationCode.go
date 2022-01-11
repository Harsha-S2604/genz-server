package users

import (
	"time"
)

// use promoted fields if possible
type UserVerificationCode struct {
	VerificationCode string 	`json: "verificationCode"`
	CreatedAt		time.Time 	`json: "createdAt"`
	Email			string		`json: "email"`
	CodeSentCount	int			`json: "codeSentCount"`	
}

func NewVerificationCode(verificationCode string, createdAt time.Time, email string, codeSentCount int) UserVerificationCode {
	return UserVerificationCode{verificationCode, createdAt, email, codeSentCount}
}