package main

import (
	"context"
	"log"

	_ "github.com/go-sql-driver/mysql"
	rmx "github.com/rog-golang-buddies"
	"github.com/rog-golang-buddies/ent"
	"github.com/spf13/viper"
)

func main() {
	log.Println("loading config")
	err := rmx.LoadConfig()
	if err != nil {
		log.Fatalf("fatal error config file: %v", err.Error())
	}

	dsn := viper.GetString("DB_USER") + ":" +
		viper.GetString("DB_PASSWORD") + "@tcp(" +
		viper.GetString("DB_HOST") + ":" +
		viper.GetString("DB_PORT") + ")/" +
		viper.GetString("DB_NAME") + "?parseTime=True"

	log.Println("opening db connection")
	client, err := ent.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed opening connection to mysql: %v", err)
	}
	defer client.Close()
	// Run the auto migration tool.
	log.Println("running auto migration")
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	s := rmx.Server{
		Port: ":8080",
	}

	log.Println("starting server")
	s.ServeHTTP()
}
