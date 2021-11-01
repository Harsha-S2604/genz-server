package users

import (
	"time"
)

// use promoted fields if possible
type UserVerificationCode struct {
	VerficationCode string 		`json: "userVerificationCode"`
	CreatedAt		time.Time 	`json: "createdAt"`
	email			string		`json: "email"`	
}

func NewVerificationCode(verificationCode string, createdAt time.Time, email string) UserVerificationCode {
	return UserVerificationCode{verificationCode, createdAt, email}
}