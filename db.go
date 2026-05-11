package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/lib/pq"
)

func ConnectToDB() *sql.DB {
	cfg := pq.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     5432,
		Database: os.Getenv("DB_NAME"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASS"),
		SSLMode: "disable",
	}

	c, err := pq.NewConnectorConfig(cfg)
	if err != nil {
		panic(err)
	}

	// var err error

	for i := range 10 {
		db = sql.OpenDB(c)

		err = db.Ping()
		if err == nil {
			log.Println("Database connected")
			return db
		}
		log.Printf("Attempt %d/10 failed: %v\n", i+1, err)
		time.Sleep(2 * time.Second)
	}

	panic("Could not connect to database after 10 attempts")
}
