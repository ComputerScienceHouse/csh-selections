package internal

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Application struct {
	Id uuid.UUID
}

/* ==============
ACTUAL FUNCTIONS GO HERE
			============= */

func uploadApplication() Application {
	//TODO: S3 upload here

	return Application{}
}

func (app Application) get() {

}

func HandleApplicationUpload(c *gin.Context) {
	file, err := c.FormFile("applicationFile")
	if err != nil {
		fmt.Println("sad")
	}
	_, _ = file.Open()

}
