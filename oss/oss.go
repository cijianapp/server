package oss

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/disintegration/gift"
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

// PutImageResize upload all files to aliyun oss after resize and crops
func PutImageResize(icon string, imageType string) (bool, string) {
	objectName, iconString := objectNameAndData(icon)
	if objectName == "" {
		return false, objectName
	}

	// iconByte, _ := base64.StdEncoding.DecodeString(iconString)

	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(iconString))

	m, _, err := image.Decode(reader)

	dst := m

	if imageType == "avatar" {
		dst = resizeAvatar(m)
	}

	if imageType == "cover" {
		dst = resizeCover(m)
	}

	buff := new(bytes.Buffer)
	err = png.Encode(buff, dst)

	err = bucket.PutObject(objectName, bytes.NewReader(buff.Bytes()))
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

func resizeAvatar(src image.Image) image.Image {
	g := gift.New(
		gift.ResizeToFill(512, 512, gift.NearestNeighborResampling, gift.CenterAnchor),
	)

	dst := image.NewRGBA(g.Bounds(src.Bounds()))

	g.Draw(dst, src)

	return dst
}

func resizeCover(src image.Image) image.Image {
	g := gift.New(
		gift.ResizeToFill(1920, 1440, gift.NearestNeighborResampling, gift.CenterAnchor),
	)

	dst := image.NewRGBA(g.Bounds(src.Bounds()))

	g.Draw(dst, src)

	return dst
}

func saveImage(filename string, img image.Image) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("os.Create failed: %v", err)
	}
	defer f.Close()
	err = png.Encode(f, img)
	if err != nil {
		log.Fatalf("png.Encode failed: %v", err)
	}
}
