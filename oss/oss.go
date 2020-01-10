package oss

import (
	"bytes"
	"encoding/base64"
)

// PutObject upload all files to aliyun oss
func PutObject(icon string) (bool, string) {
	objectName, iconString := objectNameAndData(icon)
	if objectName == "" {
		return false, objectName
	}

	iconByte, _ := base64.StdEncoding.DecodeString(iconString)
	err := bucket.PutObject(objectName, bytes.NewReader(iconByte))
	if err != nil {
		handleError(err)
	}

	return true, objectName
}

//PutObjectFromFile upload all files to aliyun oss
func PutObjectFromFile(fileID string, fileType string, fileName string) {
	bucket := ossConfig()
	objectName := fileType + "/" + fileID + "/" + fileName
	localFileName := "tmp/" + fileID + ".mp4"
	err := bucket.PutObjectFromFile(objectName, localFileName)
	if err != nil {
		handleError(err)
	}
}

// GetObject download a file from aliyun oss
func GetObject() {
	objectName := "image/23.png"
	downloadedFileName := "23.png"
	err := bucket.GetObjectToFile(objectName, downloadedFileName)
	if err != nil {
		handleError(err)
	}
}
