package store

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/luqmanshaban/kuda/internal/config"
)

func Connect(cfg *config.Config) *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", 
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		slog.Error("Failed to connect to postgres", "error", err)
		panic(err)
	}

	for i := range 10 {
		if err = db.Ping(); err == nil {
			slog.Info("Database connected")
			return db
		}
		slog.Warn("Attempt to connect to the database failed", "attempts", i + 1, "error", err)
		time.Sleep(2 * time.Second)
	}

	panic("failed to connect to the database after 10 attempts")
}