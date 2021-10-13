package validations

import (
	"regexp"
	"errors"
)

func ValidateUserEmail(email string) (bool, error) {
	var re = regexp.MustCompile(`^[a-zA-Z0-9+_.-]+@[a-zA-Z0-9.-]+$`)
	if re.MatchString(email) {
		return true, nil
	} 
	return false, errors.New("Not a valid email.")
}