package main

import (
	"ecom_apiv1/db"
	"ecom_apiv1/internal/handler"
	"ecom_apiv1/internal/server"
	"ecom_apiv1/internal/storer"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

const minSecretKeySize = 32

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	secretKey := os.Getenv("SECRET_KEY")
	if len(secretKey) < minSecretKeySize {
		log.Fatalf("SECRET_KEY must be at least %d characters", minSecretKeySize)
	}
	sqlx, err := db.GetConnection()
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	defer sqlx.Close()
	log.Println("Succesfully connecting database")

	str := storer.NewMySQLStorage(sqlx)
	srv := server.NewServer(str)

	hdl := handler.NewHandler(srv, secretKey)
	handler.RegisterRoutes(hdl)
	handler.Start(":8000")
}
