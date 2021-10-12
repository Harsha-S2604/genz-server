package main

import (
	"os"
	"log"
	"database/sql"


	"github.com/Harsha-S2604/genz-server/config/db"
	"github.com/Harsha-S2604/genz-server/routes"
)

var genzDB *sql.DB
var dbErr error

func init() {
	// create and load log file
	logFile, fileLoaderr := os.OpenFile("genz_logs.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if fileLoaderr != nil {
		log.Fatal(fileLoaderr.Error())
	}
	log.SetOutput(logFile)

	// connect to database
	genzDB, dbErr = db.ConnectDB()
	if dbErr != nil {
		panic(dbErr.Error())
	}
	log.Println("Database running on port no.3306.")
}

func main() {
	
	// start server
	r := routes.SetupRouter(genzDB)
	r.Run(":8080")
	defer genzDB.Close()

}
