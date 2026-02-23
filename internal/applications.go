package internal

import (
	"context"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Application struct {
	ID      uuid.UUID `gorm:"primarykey"`
	Ratings []Rating
}

type Rating struct {
	ApplicationID uuid.UUID `gorm:"primarykey"`
	Member        string    `gorm:"primarykey"`
	SubmittedTime time.Time
	Score         int
}

/* ==============
ACTUAL FUNCTIONS GO HERE
			============= */

func uploadApplication(file multipart.File) Application {
	app := Application{ID: uuid.New()}
	db.Create(&app)
	_, err := s3client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(env.BucketName),
		Key:    aws.String(app.ID.String()),
		Body:   file,
	})
	if err != nil {
		log.Println("Failed while uploading application", err)
	}
	return app
}

func (app Application) get() {

}

/* ============
WEB FUNCTIONS GO HERE
		  ============ */

// GET Request
func HandleApplicationUploadPage(c *gin.Context) {
	c.HTML(http.StatusOK, "uploadApplication.tmpl", templateHeaders(c))

}

// POST Request
func HandleApplicationFileUpload(c *gin.Context) {
	fileH, err := c.FormFile("applicationFile")
	if err != nil {
		log.Println("Something went wrong with the application upload.\n\t", err)
	}
	file, err := fileH.Open()
	if err != nil {
		log.Println("Failed to get file from application upload", err)
	}
	uploadApplication(file)

	//TODO: display status of upload?
	HandleApplicationUploadPage(c)
}
