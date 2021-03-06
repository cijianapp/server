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

// OssURL is the base url of oss
const OssURL = "http://cijiantest01.oss-us-east-1.aliyuncs.com/"

func ossConfig() *oss.Bucket {
	endpoint := "http://oss-us-east-1.aliyuncs.com"
	accessKeyID := "LTAI4FhCjD***TkQYUShUdgKZgT"
	accessKeySecret := "kSu78QuaDX5Xz***yCXaTZ9nqJrNY5aoM"
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
