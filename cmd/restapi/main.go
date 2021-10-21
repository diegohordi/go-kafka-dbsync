package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/diegohordi/go-kafka/internal/catalogue"
	"github.com/diegohordi/go-kafka/internal/configs"
	"github.com/diegohordi/go-kafka/internal/database"
	"github.com/diegohordi/go-kafka/internal/kafka"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var configPath = flag.String("config", "", "Config file path")

// loadConfigurations loads system configurations based on the given config file.
func loadConfigurations() configs.Configurer {
	config, err := configs.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

// createDBConnection creates a new database connection based on the given configuration.
func createDBConnection(config configs.Configurer) database.Connection {
	dbConn, err := database.NewConnection(config.DB())
	if err != nil {
		log.Fatal(err)
	}
	return dbConn
}

func main() {

	flag.Parse()
	config := loadConfigurations()
	dbConn := createDBConnection(config)
	logger := log.New(os.Stdout, "", log.LstdFlags)

	router := chi.NewRouter()
	router.Use(middleware.Heartbeat("/health"))
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.SetHeader("Content-type", "application/json"))

	kafkaClient := kafka.NewClient(config.Kafka(), "films")

	catalogueService := catalogue.NewService(dbConn, kafkaClient)
	catalogue.Setup(router, catalogueService)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.App().Port()),
		Handler:      router,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	log.Println(fmt.Sprint("server started listening at ", config.App().Port()))

	<-exit
	log.Println(logger, "server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		dbConn.Close()
		kafkaClient.Close()
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal(fmt.Errorf("an error occurred while server is shutting down: %w", err))
	}

	log.Println("server shutdown successfully")
}
