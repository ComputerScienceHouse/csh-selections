package internal

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Application struct {
	ID           uuid.UUID `gorm:"primarykey"`
	Ratings      []Rating  `gorm:"-"`
	PresignedURL string    `gorm:"-"`
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

func uploadApplication(file multipart.File) (Application, error) {
	app := Application{ID: uuid.New()}
	_, err := s3client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(env.BucketName),
		Key:    aws.String(app.ID.String() + ".pdf"),
		Body:   file,
	})
	if err != nil {
		log.Println("Failed while uploading application", err)
		return Application{ID: uuid.UUID{}}, err
	}
	db.Create(&app)
	return app, nil
}

func listBucketApplications() []types.Object {
	res := s3.NewListObjectsV2Paginator(s3client, &s3.ListObjectsV2Input{Bucket: aws.String(env.BucketName)})
	var ret []types.Object
	for res.HasMorePages() {
		output, err := res.NextPage(context.Background())
		if err != nil {
			log.Println("Failed while listing bucket applications", err)
			return nil
		}
		ret = append(ret, output.Contents...)
		for _, item := range output.Contents {
			fmt.Println("\t", *item.Key, item.LastModified, *item.Size)
		}
	}
	return ret
}

func clearBucket() {
	objects := listBucketApplications()
	for _, object := range objects {
		_, err := s3client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
			Bucket: aws.String(env.BucketName),
			Key:    aws.String(*object.Key)})
		if err != nil {
			log.Println("Issue while clearing bucket applications", err)
		}
	}
}

func getApplications() []*Application {
	res := make([]*Application, 0)
	db.Find(&res)
	for _, app := range res {
		app.PresignedURL = app.GetPresignedURL()
	}
	return res
}

func getApplication(uuid uuid.UUID) *Application {
	res := Application{ID: uuid}
	res.PresignedURL = res.GetPresignedURL()
	return &res
}

func (app Application) Delete() error {
	_, err := s3client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(env.BucketName),
		Key:    aws.String(app.ID.String() + ".pdf")})
	if err != nil {
		return err
	}
	db.Delete(&app)
	return nil
}

func (app Application) GetPresignedURL() string {
	res, err := s3presign.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(env.BucketName),
		Key:    aws.String(app.ID.String() + ".pdf"),
	}, func(options *s3.PresignOptions) {
		options.Expires = time.Minute
	})
	if err != nil {
		log.Println("Failed while presigning application url", err)
		return ""
	}
	return res.URL
}

/* ============
WEB FUNCTIONS GO HERE
		  ============ */

// GET Requests

func HandleApplicationManagementPage(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, nil)
		return
	}
	c.HTML(http.StatusOK, "applicationManagement.tmpl", templateHeaders(c, map[string]any{"Applications": getApplications(), "Teams": getAllTeams()}))
}

func HandleApplicationGet(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Println("Failed while parsing application id", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := getUserData(c)
	if getTeamForMember(user.Username).ApplicationID != id || !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	app := Application{ID: id}
	c.JSON(http.StatusOK, map[string]any{"url": app.GetPresignedURL()})
}

// POST Requests

func HandleApplicationFileUpload(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	fileH, err := c.FormFile("applicationFile")
	if err != nil {
		log.Println("Something went wrong with the application upload.\n\t", err)
		c.JSON(http.StatusBadRequest, "Something went wrong with the application upload. Please try again.")
		return
	}
	if fileH.Header.Get("Content-Type") != "application/pdf" {
		c.JSON(http.StatusBadRequest, "You did not upload a PDF. Please upload a PDF.")
		return
	}
	file, err := fileH.Open()
	if err != nil {
		log.Println("Failed to get file from application upload", err)
		c.JSON(http.StatusBadRequest, "Something was wrong with the application upload. Please try again.")
	}

	_, err = uploadApplication(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	//TODO: display status of upload?
	c.JSON(http.StatusOK, gin.H{})
}

func HandleApplicationDelete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Println("Failed while parsing application id", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	err = Application{ID: id}.Delete()
	if err != nil {
		log.Println("Error deleting application", id, err)
		return
	}
}
