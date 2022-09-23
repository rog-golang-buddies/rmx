package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"github.com/rog-golang-buddies/rapidmidiex/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	// ? want to move to viper ASAP
	port := getEnv("PORT", "8889")

	sCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	// ? should this defined within the instantiation of a new service
	c := cors.Options{
		AllowedOrigins:   []string{"*"}, // ? band-aid, needs to change to a flag
		AllowCredentials: true,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}

	srv := http.Server{
		Addr:    ":" + port,
		Handler: cors.New(c).Handler(service.New(chi.NewMux())),
		// max time to read request from the client
		ReadTimeout: 10 * time.Second,
		// max time to write response to the client
		WriteTimeout: 10 * time.Second,
		// max time for connections using TCP Keep-Alive
		IdleTimeout: 120 * time.Second,
		BaseContext: func(_ net.Listener) context.Context { return sCtx },
	}

	g, gCtx := errgroup.WithContext(sCtx)

	g.Go(func() error {
		// Run the server
		log.Printf("App server starting on %s", srv.Addr)
		return srv.ListenAndServe()
	})

	g.Go(func() error {
		<-gCtx.Done()
		return srv.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		log.Printf("exit reason: %s \n", err)
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func init() {
	// // name of config file (without extension)
	// viper.SetConfigName("config")
	// // REQUIRED if the config file does not have the extension in the name
	// viper.SetConfigType("env")
	// // optionally look for config in the working directory
	// viper.AddConfigPath(".")

	//// Set Default variables
	// viper.SetDefault("PORT", "8080")

	// viper.AutomaticEnv()

	// if err := viper.ReadInConfig(); err != nil {
	// 	panic(err)
	// }
}

// func LoadConfig(path string) (config Config, err error) {
// 	// Read file path
// 	viper.AddConfigPath(path)
// 	// set config file and path
// 	viper.SetConfigName("app")
// 	viper.SetConfigType("env")
// 	// watching changes in app.env
// 	viper.AutomaticEnv()
// 	// reading the config file
// 	err = viper.ReadInConfig()
// 	if err != nil {
// 		return
// 	}

// 	err = viper.Unmarshal(&config)
// 	return
// }

func loadConfig() error {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Join(filepath.Dir(b), "../")
	viper.SetConfigFile(basepath + ".env")
	// viper.AddConfigPath("../")
	viper.SetConfigType("dotenv")
	// viper.SetConfigFile(".env")

	return viper.ReadInConfig()
}

var (
	port = flag.Int("SERVER_PORT", 8888, "The port that the server will be running on")
)
