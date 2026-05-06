package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/joho/godotenv"
)

var db *sql.DB

func main()  {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	db = ConnectToDB()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", CreateJH)
	mux.HandleFunc("GET /jobs/{job_id}", GetJH)

	fmt.Println("SERVER RUNNING ON localhost:8000")

	http.ListenAndServe(":8000", mux)
}