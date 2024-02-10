package main

import (
	"database/sql"
	"os"

	"github.com/go-sql-driver/mysql"
)

func NewDBConnection() *sql.DB {
	cfg := mysql.Config{
		User:   os.Getenv("DB_USER"),
		Passwd: os.Getenv("DB_PASS"),
		DBName: os.Getenv("DB_NAME"),
		Net:    "tcp",
		Addr:   os.Getenv("DB_ENDPOINT"),
	}

	var err error
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		panic(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		db.Close()
		panic(pingErr)
	}
	// fmt.Println("New connection created successfully!")

	return db
}
