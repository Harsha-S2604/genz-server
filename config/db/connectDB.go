package db

import (
	"log"
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDB() (*sql.DB, error) {
	log.Println("Connecting to database...")
	userName, password, port, dbName := os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME")

	dbURL := userName+":"+password+"@tcp(127.0.0.1:"+port+")/"+dbName+"?parseTime=true"
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		return nil, err
	}

	// make sure the connection is available
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
