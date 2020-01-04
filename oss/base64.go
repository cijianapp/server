package oss

import (
	"github.com/segmentio/ksuid"

	"encoding/base64"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"strings"
)

func objectNameAndData(icon string) (string, string) {

	objectName := ""

	s := strings.Split(icon, ",")

	if len(s) < 2 {
		return objectName, ""
	}

	fileName := ksuid.New().String()

	if strings.Contains(s[0], "jpeg") {
		objectName = "image/" + fileName + ".jpg"
	}

	if strings.Contains(s[0], "png") {
		objectName = "image/" + fileName + ".png"
	}

	if strings.Contains(s[0], "gif") {
		objectName = "image/" + fileName + ".gif"
	}

	return objectName, s[1]

}

func base2image() {
	// reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(s[1]))
}

func base2jpeg(reader io.Reader, fileName string) {

	file, err := jpeg.Decode(reader)
	handleError(err)

	opt := jpeg.Options{
		Quality: 100,
	}

	out, err := os.Create("tmp/" + fileName + ".jpg")

	err = jpeg.Encode(out, file, &opt) // put quality to 80%
	handleError(err)
}

func base2png(reader io.Reader, fileName string) {

	file, err := png.Decode(reader)
	handleError(err)

	out, err := os.Create("tmp/" + fileName + ".png")

	err = png.Encode(out, file)
	handleError(err)
}

func base2gif(reader io.Reader, fileName string) {

	file, err := gif.DecodeAll(reader)
	handleError(err)

	out, err := os.Create("tmp/" + fileName + ".gif")

	err = gif.EncodeAll(out, file)
	handleError(err)
}

//ExampleDecodeConfig ssss
func ExampleDecodeConfig() {
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader("data1"))
	config, format, err := image.DecodeConfig(reader)
	if err != nil {
		log.Fatal(err)
	}
	println("Width:", config.Width, "Height:", config.Height, "Format:", format)
}
