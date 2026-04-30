package internal

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
)

type Application struct {
	ID           int `gorm:"primarykey;autoIncrement"`
	Assigned     bool
	Ratings      *[]Rating `gorm:"-"`
	PresignedURL string    `gorm:"-"`
}

type Rating struct {
	ApplicationID int    `gorm:"primarykey"`
	Member        string `gorm:"primarykey"`
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

func uploadApplication(appID int, file multipart.File) (Application, error) {
	//tx := db.Where("1 = 1").Find(&[]Application{})
	app := Application{ID: appID} //int(tx.RowsAffected + 1)
	_, err := s3client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(env.BucketName),
		Key:    aws.String(strconv.Itoa(app.ID) + ".pdf"),
		Body:   file,
	})
	if err != nil {
		log.Println("Failed while uploading application", err)
		return Application{}, err
	}
	tx := db.Create(&app)
	if tx.Error != nil {
		log.Println("Failed while add application to DB", tx.Error)
		return Application{}, tx.Error
	}
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
	db.Order("id").Find(&res)
	for _, app := range res {
		app.GetPresignedURL()
		app.GetRatings()
	}
	return res
}

func getApplication(id int) *Application {
	res := Application{ID: id}
	res.GetPresignedURL()
	res.GetRatings()
	return &res
}

func GetApplicationScore(id int) int {
	application := getApplication(id)
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

func getApplicationScores(id int) []int {
	ret := make([]int, 0)
	application := getApplication(id)
	if len(*application.Ratings) == 0 {
		return ret
	}
	for _, rating := range *application.Ratings {
		ret = append(ret, rating.Score)
	}
	return ret
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

func didUserRateApplication(applicationId int, username string) bool {
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
		Key:    aws.String(strconv.Itoa(app.ID) + ".pdf")})
	if err != nil {
		return err
	}
	db.Delete(&app)
	return nil
}

func (app *Application) GetPresignedURL() {
	res, err := s3presign.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(env.BucketName),
		Key:    aws.String(strconv.Itoa(app.ID) + ".pdf"),
	}, func(options *s3.PresignOptions) {
		options.Expires = time.Minute
	})
	if err != nil {
		log.Println("Failed while presigning application url", err)
		app.PresignedURL = ""
	}
	app.PresignedURL = res.URL
}

func (app *Application) GetApplicationData() []byte {
	ret := make([]byte, 0)
	objectRes, err := s3client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(env.BucketName),
		Key:    aws.String(strconv.Itoa(app.ID) + ".pdf"),
	})
	if err != nil {
		log.Println("Failed while getting application data", err)
		return ret
	}
	all, err := io.ReadAll(objectRes.Body)
	if err != nil {
		log.Println("Failed while reading application data", err)
		return ret
	}
	//ret = base64.StdEncoding.EncodeToString(all)
	return all
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
	teams := getAllTeams()
	biggestTeam := 0
	for _, team := range teams {
		clen := len(team.Members) // one of these days I'll stop naming the temporary variables something stupid
		if clen > biggestTeam {
			biggestTeam = clen
		}
	}
	allScores := make(map[int][]int)
	totalScores := make([]int, 0)
	apps := getApplications()
	for _, app := range apps {
		allScores[app.ID] = getApplicationScores(app.ID)
		totalScores = append(totalScores, GetApplicationScore(app.ID))
	}

	c.HTML(http.StatusOK, "applicationManagement.tmpl", templateHeaders(c, map[string]any{
		"Applications":      apps,
		"Teams":             teams,
		"BiggestTeam":       biggestTeam,
		"ApplicationScores": allScores,
		"TotalScores":       totalScores,
	}))
}

func HandleApplicationGet(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("Failed while parsing application id", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := getUserData(c)
	if !(getTeamForMember(user.Username).ApplicationID == id || isUserAdmin(c)) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	app := Application{ID: id}
	//c.JSON(http.StatusOK, map[string]any{"data": app.GetApplicationData()})
	c.Writer.Header().Set("Content-Type", "application/pdf")
	c.Writer.Write(app.GetApplicationData())
}

// POST Requests

func HandleApplicationFileUpload(c *gin.Context) {
	if !isUserAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	appID, ok := c.GetPostForm("applicationID")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing applicationID"})
		return
	}
	atoi, err := strconv.Atoi(appID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": appID + " is not an integer."})
		return
	}
	fileH, err := c.FormFile("applicationFile")
	if err != nil {
		log.Println("Something went wrong with the application upload.\n\t", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Something went wrong with the application upload. Please try again."})
		return
	}
	if fileH.Header.Get("Content-Type") != "application/pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You did not upload a PDF. Please upload a PDF."})
		return
	}
	file, err := fileH.Open()
	if err != nil {
		log.Println("Failed to get file from application upload", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Something was wrong with the application upload. Please try again."})
	}

	_, err = uploadApplication(atoi, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	id, err := strconv.Atoi(c.Param("id"))
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("Failed while parsing application id", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}
	user := getUserData(c)
	if !(getTeamForMember(user.Username).ApplicationID == id || isUserAdmin(c)) {
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
	csvOut.Write([]string{"ApplicationID", "Total", "Score By Member"}) // Headers
	for _, application := range getApplications() {
		out := []string{strconv.Itoa(application.ID), strconv.Itoa(GetApplicationScore(application.ID))}
		// append the scores one by one to the output
		for _, val := range getApplicationScores(application.ID) {
			out = append(out, strconv.Itoa(val))
		}
		err := csvOut.Write(out)
		if err != nil {
			log.Println("Failed to write to csv", err)
			return
		}
	}
	csvOut.Flush()
}
