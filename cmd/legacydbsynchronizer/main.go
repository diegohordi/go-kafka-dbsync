package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/diegohordi/go-kafka/internal/catalogue"
	"github.com/diegohordi/go-kafka/internal/configs"
	"github.com/diegohordi/go-kafka/internal/database"
	"github.com/diegohordi/go-kafka/internal/kafka"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const getFilmByUUIDSQL = "select film_id from film where uuid = ?"
const insertFilmSQL = "insert into film (uuid, language_id, title, release_year, last_update) select ?, 1, ?, ?, ? where (select count(film_id) from film where uuid = ?) = 0"
const updateFilmSQL = "update film set title = ?, release_year = ? where uuid = ?"

var configPath = flag.String("config", "", "Config file path")
var dbConn database.Connection

func loadConfigurations() configs.Configurer {
	config, err := configs.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

func createDBConnection(config configs.DBConfigurer) database.Connection {
	conn, err := database.NewConnection(config)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func readFilm(key, value []byte) error {
	r := bytes.NewReader(value)
	film := &catalogue.Film{}
	if err := json.NewDecoder(r).Decode(film); err != nil {
		return err
	}
	return insertOrUpdate(film)
}

func insertOrUpdate(film *catalogue.Film) error {
	ctx, cancel := dbConn.CreateContext(context.TODO())
	defer cancel()
	var id int
	err := dbConn.DB().QueryRowContext(ctx, getFilmByUUIDSQL, film.UUID).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return insert(film)
	case err != nil:
		return fmt.Errorf("an error occured while searching: %w", err)
	default:
		return update(film)
	}
}

func insert(film *catalogue.Film) error {
	ctx, cancel := dbConn.CreateContext(context.TODO())
	defer cancel()
	film.LastUpdate = time.Now()
	res, err := dbConn.DB().ExecContext(ctx, insertFilmSQL, film.UUID, film.Title, film.Year, film.LastUpdate, film.UUID)
	if err != nil {
		return fmt.Errorf("an error occured while inserting: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("an error occured while inserting: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("the film with UUID %s was not inserted", film.UUID)
	}
	return nil
}

func update(film *catalogue.Film) error {
	ctx, cancel := dbConn.CreateContext(context.TODO())
	defer cancel()
	res, err := dbConn.DB().ExecContext(ctx, updateFilmSQL, film.Title, film.Year, film.UUID)
	if err != nil {
		return fmt.Errorf("an error occured while updating: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("an error occured while updating: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("the film with UUID %s was not updated", film.UUID)
	}
	return nil
}

func main() {

	flag.Parse()
	config := loadConfigurations()
	dbConn = createDBConnection(config.DB())

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	kafkaClient := kafka.NewClient(config.Kafka(), "catalogue")

	ctx := context.Background()

	go func() {

		for {
			err := kafkaClient.Read(ctx, readFilm)
			if err != nil {
				log.Println(err)
			}
		}

	}()

	log.Println("legacydb consumer started")

	<-exit
	log.Println("server stopped")

	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		kafkaClient.Close()
		cancel()
	}()

	log.Println("consumer shutdown successfully")
}
