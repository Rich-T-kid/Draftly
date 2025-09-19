package internal

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	DbInstance *sql.DB
)

func GetDocument(DocumentID string) (string, error) {
	if fileExists(DocumentID) {
		fmt.Println("File exists locally, reading from local storage.")
		content, err := os.ReadFile(dirName + "/" + DocumentID)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}
	bucket := cfg.Bucket
	item := DocumentID
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region),
		Credentials: credentials.NewStaticCredentials(
			cfg.AwsAccessKey,
			cfg.AwsSecretKey,
			"",
		),
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		return "", err
	}
	downloader := s3manager.NewDownloader(sess)

	file, err := os.Create(dirName + "/" + item)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return "", err
	}
	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	})
	if err != nil {
		return "", err
	}
	defer file.Close()
	return "", nil
}
func fileExists(filename string) bool {
	info, err := os.Stat(dirName + "/" + filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Postgress

func Connect() *sql.DB {
	if DbInstance != nil {
		return DbInstance
	}
	// Connection parameters
	host := cfg.Host
	port := cfg.Port
	user := cfg.User
	password := cfg.Password
	dbname := cfg.DbName

	// Build connection string
	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	// Open database
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	// Verify connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}

	DbInstance = db
	return db
}
