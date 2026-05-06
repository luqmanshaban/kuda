package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/lib/pq"
)

func ConnectToDB() *sql.DB {
	cfg := pq.Config{
		Host: os.Getenv("DB_HOST"),
		Port: 5432,
		Database: os.Getenv("DB_NAME"),
		User: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASS"),
	}

	c, err := pq.NewConnectorConfig(cfg)
	if err != nil {
		panic(err)
	}

	db = sql.OpenDB(c)

	err = db.Ping(); 
	if err != nil {
		panic(err)
	}
	log.Println("Database connected")

	return db
}