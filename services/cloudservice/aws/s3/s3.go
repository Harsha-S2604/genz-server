package s3

import (
	"context"
	"io/ioutil"
	"encoding/base64"
	"os"
	"mime/multipart"
	"bytes"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/aws"
)

var (
	bucketName = os.Getenv("DEV_BUCKET_NAME")
	parentObj = os.Getenv("PARENT_OBJECT")
)

func UploadImageToS3(img *multipart.FileHeader, userId string, blogId string)(bool, error) {

	// load the configuration file
	cfg, cfgErr := config.LoadDefaultConfig(context.TODO())
	if cfgErr != nil {
		return false, cfgErr
	}

	// create a AWS S3 client using "cfg" variable
	client := s3.NewFromConfig(cfg)



	f, fileErr := img.Open()
	if fileErr != nil {
		return false, fileErr
	}
	defer f.Close()

	size := img.Size
	buffer := make([]byte, size)
	key := parentObj+blogId
	f.Read(buffer)
	fileBytes := bytes.NewReader(buffer)
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   fileBytes,
	}

	_, responseErr := client.PutObject(context.TODO(), input)

	if responseErr != nil {
		return false, responseErr
	}

	return true, nil
}


func GetObjectFromS3(blogId string)(string, error) {
	// load the configuration file
	cfg, cfgErr := config.LoadDefaultConfig(context.TODO())
	if cfgErr != nil {
		return "", cfgErr
	}

	// create a AWS S3 client using "cfg" variable
	client := s3.NewFromConfig(cfg)

	input := &s3.GetObjectInput {
		Bucket: aws.String(bucketName),
		Key: aws.String(parentObj+"/"+blogId),
	}

	response, responseErr := client.GetObject(context.TODO(), input)
	if responseErr != nil {
		return "", responseErr
	}

	data, dataErr := ioutil.ReadAll(response.Body)
	if dataErr != nil {
		return "", dataErr 
	}	
	
	base64Enc := base64.StdEncoding.EncodeToString(data)
	return base64Enc, nil
}