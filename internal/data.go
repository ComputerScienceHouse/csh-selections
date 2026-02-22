package internal

import (
	"context"
	"log"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var s3client *s3.Client

func InitData() {
	selections_db, err := gorm.Open(postgres.Open(env.DatabaseUri), &gorm.Config{})
	if err != nil {
		log.Println("Couldn't open connection to database:", err)
	}
	db = selections_db
	db.AutoMigrate(&Application{})
	db.AutoMigrate(&Team{})

	s3config, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Println("Error loading s3 config:", err)
	}
	s3client = s3.NewFromConfig(s3config)
}
