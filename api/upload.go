package api

import (
	"github.com/cijianapp/server/oss"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/ksuid"

	"os"
)

// upload one file
func newUpload(c *gin.Context) {

	form, err := c.MultipartForm()
	handleError(err)

	files := form.File["filepond"]

	fileID := ksuid.New().String()

	c.SaveUploadedFile(files[0], "./tmp/"+fileID+".mp4")

	oss.PutObjectFromFile(fileID, "video", files[0].Filename)

	os.Remove("./tmp/" + fileID + ".mp4")

	c.String(200, "video"+"/"+fileID+"/"+files[0].Filename)
}

type postImage struct {
	Image string `form:"image" json:"image" binding:"required"`
}

func uploadImage(c *gin.Context) {

	var postImageVals postImage

	if err := c.ShouldBind(&postImageVals); err != nil {
		c.JSON(401, gin.H{"error": "cannot upload the image"})
		return
	}

	_, imageURL := oss.PutObject(postImageVals.Image)

	c.JSON(200, gin.H{"code": 200, "imageURL": imageURL})

}
