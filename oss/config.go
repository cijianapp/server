package oss

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func handleError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
	}
}

var bucket = ossConfig()

func ossConfig() *oss.Bucket {
	endpoint := "http://oss-us-east-1.aliyuncs.com"
	accessKeyID := "LTAI4FhCjDTkQYUShUdgKZgT"
	accessKeySecret := "kSu78QuaDX5XzyCXaTZ9nqJrNY5aoM"
	bucketName := "cijiantest01"
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		handleError(err)
	}
	// 获取存储空间。
	bucket, err := client.Bucket(bucketName)
	handleError(err)
	return bucket
}
