package verification

import (
	"crypto/rand"
	"math/big"
	"log"
)

func GenerateSixDigitCode() (string, error) {

	randomSixDigitNumber, randomSixDigitNumberErr := rand.Int(rand.Reader, big.NewInt(999999 - 100000 + 1))

	if randomSixDigitNumberErr != nil {
		log.Println("Function GenerateSixDigitCode Err:", randomSixDigitNumberErr.Error())
		return "", randomSixDigitNumberErr
	}

	return randomSixDigitNumber.String(), nil

}