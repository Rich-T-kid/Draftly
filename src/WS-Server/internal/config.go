package internal

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Host         string
	Port         string
	User         string
	Password     string
	DbName       string
	CrudPort     string
	WSPort       string
	Bucket       string // TODO:
	Region       string // TODO:
	AwsAccessKey string // TODO:
	AwsSecretKey string // TODO:
}

var (
	cfg     *Config
	dirName = "temp-storage"
)

func init() {
	cfg = NewConfig()
	err := os.Mkdir(dirName, 0755)
	if err != nil {
		if !os.IsExist(err) { // only fatal if it's NOT a "directory exists" error
			log.Fatal("Error creating directory:", err)
			return
		}

	}
}

func NewConfig() *Config {
	if cfg != nil {
		return cfg
	}
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	cfg := &Config{
		Host:         must("POSTGRESS_HOST"),
		Port:         must("POSTGRESS_PORT"),
		User:         must("POSTGRESS_USER"),
		Password:     must("POSTGRESS_PASSWORD"),
		DbName:       must("POSTGRESS_DB_NAME"),
		CrudPort:     must("CRUD_PORT"),
		WSPort:       must("WS_PORT"),
		Bucket:       must("BUCKET_NAME"),
		Region:       must("REGION"),
		AwsAccessKey: must("AWS_ACCESS_KEY"),
		AwsSecretKey: must("AWS_SECRET_KEY"),
	}
	return cfg
}

func must(name string) string {

	val := os.Getenv(name)
	if val == "" {
		log.Fatalf("Environment variable %s not set", name)
	}
	return val
}
