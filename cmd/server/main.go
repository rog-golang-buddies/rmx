package main

import (
	"log"

	// _ "github.com/go-sql-driver/mysql"
	rmx "github.com/rog-golang-buddies/rapidmidiex"
)

func main() {
	log.Println("loading config")
	err := rmx.LoadConfig()
	if err != nil {
		log.Fatalf("fatal error config file: %v", err.Error())
	}

	// database connection shouldn't be implemented here
	// 	dsn := viper.GetString("DB_USER") + ":" +
	// 		viper.GetString("DB_PASSWORD") + "@tcp(" +
	// 		viper.GetString("DB_HOST") + ":" +
	// 		viper.GetString("DB_PORT") + ")/" +
	// 		viper.GetString("DB_NAME") + "?parseTime=True"
	//
	// 	log.Println("opening db connection")
	// 	db, err := sql.Open("mysql", dsn)
	// 	if err != nil {
	// 		log.Fatalf("database connectoin failed: %v", err.Error())
	// 	}

	s := rmx.Server{
		Port: ":8080",
	}

	log.Println("starting server")
	s.ServeHTTP()
}
