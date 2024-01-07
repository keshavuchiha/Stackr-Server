package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"server/problems"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func DB_START() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	connStr = fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", dbHost, dbPort, dbUser, dbName, dbPassword)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxIdleTime(time.Minute * 5)
	db.SetMaxOpenConns(10)
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()
	var problemModal problems.ProblemModel
	if !problemModal.CheckForProblemOfTheDay(tx) {
		fmt.Println("Adding Problem of the day")
		problemModal.AddProblemOfTheDay(tx)
	}
	tx.Commit()
	fmt.Println("Successfully connected to database!")
}
