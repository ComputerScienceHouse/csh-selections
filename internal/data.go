package internal

import (
	"context"
	"log"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var s3client *s3.Client
var oidcClient OIDCClient

func InitData() {
	selectionsDb, err := gorm.Open(postgres.Open(env.DatabaseUri), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalln("Couldn't open connection to database:", err)
		return
	}
	db = selectionsDb
	err = db.AutoMigrate(&Application{})
	if err != nil {
		log.Println("Couldn't migrate Application table:", err)
		return
	}
	err = db.AutoMigrate(&Team{})
	if err != nil {
		log.Println("Couldn't migrate Team table:", err)
		return
	}
	err = db.AutoMigrate(&Membership{})
	if err != nil {
		log.Println("Couldn't migrate Membership table:", err)
		return
	}
	err = db.AutoMigrate(&SessionState{})
	if err != nil {
		log.Println("Couldn't migrate SessionState table:", err)
		return
	}
	err = db.AutoMigrate(&Rating{})
	if err != nil {
		log.Println("Couldn't migrate Rating table:", err)
		return
	}
	err = db.AutoMigrate(&SessionAttendance{})
	if err != nil {
		log.Println("Couldn't migrate Rating table:", err)
		return
	}

	s3config, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Println("Error loading s3 config:", err)
	}
	s3client = s3.NewFromConfig(s3config)

	oidcClient.setupOidcClient(env.OIDCClientID, env.OIDCClientSecret)
}
