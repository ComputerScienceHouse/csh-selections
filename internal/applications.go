package internal

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Application struct {
	ID           uuid.UUID `gorm:"primarykey"`
	Assigned     bool
	Ratings      *[]Rating `gorm:"-"`
	PresignedURL string    `gorm:"-"`
}

type Rating struct {
	ApplicationID uuid.UUID `gorm:"primarykey"`
	Member        string    `gorm:"primarykey"`
	SubmittedTime time.Time
	Score         int
}

type Criterion struct {
	Name     string `gorm:"primarykey"`
	MinScore int
	MaxScore int
	Weight   int
}

func (Criterion) TableName() string {
	return "criteria"
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
		app.GetPresignedURL()
		app.GetRatings()
	}
	return res
}

func getApplication(uuid uuid.UUID) *Application {
	res := Application{ID: uuid}
	res.GetPresignedURL()
	res.GetRatings()
	return &res
}

func GetApplicationScore(uuid uuid.UUID) int {
	application := getApplication(uuid)
	rateLen := len(*application.Ratings)
	if rateLen == 0 {
		return 0
	}
	score := 0
	for _, rating := range *application.Ratings {
		score += rating.Score
	}
	return score / rateLen
}

func getCriteria() []Criterion {
	if ret, b := goCache.Get("criteria"); b {
		return ret.([]Criterion)
	}
	var res []Criterion
	db.Find(&res)
	goCache.SetDefault("criteria", res)
	return res
}

func didUserRateApplication(applicationId uuid.UUID, username string) bool {
	tx := db.Find(&Rating{ApplicationID: applicationId, Member: username})
	return tx.RowsAffected > 0
}

func dropAllApplications() {
	clearBucket()
	db.Where("1 = 1").Delete(&Application{})
}

func (app *Application) Delete() error {
	_, err := s3client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(env.BucketName),
		Key:    aws.String(app.ID.String() + ".pdf")})
	if err != nil {
		return err
	}
	db.Delete(&app)
	return nil
}

func (app *Application) GetPresignedURL() {
	res, err := s3presign.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(env.BucketName),
		Key:    aws.String(app.ID.String() + ".pdf"),
	}, func(options *s3.PresignOptions) {
		options.Expires = time.Minute
	})
	if err != nil {
		log.Println("Failed while presigning application url", err)
		app.PresignedURL = ""
	}
	app.PresignedURL = res.URL
}

func (app *Application) GetRatings() {
	db.Find(&app.Ratings, Rating{ApplicationID: app.ID})
	if app.Ratings == nil {
		app.Ratings = &[]Rating{}
	}
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
	fmt.Println("hello")
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
	c.JSON(http.StatusOK, map[string]any{"url": app.PresignedURL})
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
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Println("Failed while parsing application id", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	app := &Application{ID: id}
	err = app.Delete()
	if err != nil {
		log.Println("Error deleting application", id, err)
		return
	}
}

func HandleApplicationRating(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Println("Failed while parsing application id", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}
	user := getUserData(c)
	if getTeamForMember(user.Username).ApplicationID != id || !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if didUserRateApplication(id, user.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You already rated this application"})
		return
	}
	//score the application
	totalScore := 0
	for _, crit := range getCriteria() {
		score := c.PostForm(crit.Name)
		scoreInt, err := strconv.Atoi(score)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		totalScore += scoreInt * crit.Weight
	}
	//present score
	application := getApplication(id)
	rating := Rating{
		ApplicationID: id,
		Member:        user.Username,
		SubmittedTime: time.Now(),
		Score:         totalScore,
	}
	db.Save(&rating)
	*application.Ratings = append(*application.Ratings, rating)
}

func HandleApplicationExport(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	filename := "CSH_SELECTIONS_" + time.Now().Format(time.DateOnly) + ".csv"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv")
	csvOut := csv.NewWriter(c.Writer)
	csvOut.Write([]string{"ApplicationID", "Score"}) // Headers
	for _, application := range getApplications() {
		err := csvOut.Write([]string{application.ID.String(), strconv.Itoa(GetApplicationScore(application.ID))})
		if err != nil {
			log.Println("Failed to write to csv", err)
			return
		}
	}
	csvOut.Flush()
}
