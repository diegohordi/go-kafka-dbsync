package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/diegohordi/go-kafka/internal/configs"
	"github.com/diegohordi/go-kafka/internal/database"
	"github.com/diegohordi/go-kafka/internal/kafka"
	"github.com/google/uuid"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const getFilmByUUIDSQL = "select id from films where uuid = ? or external_id = ?"
const insertFilmSQL = "insert into films (external_id, uuid, title, year, last_update) select ?, ?, ?, ?, ? where (select count(id) from films where uuid = ?) = 0"
const updateFilmSQL = "update films set title = ?, year = ?, external_id = ? where uuid = ?"

type Year int

func (y *Year) UnmarshalJSON(i []byte) error {
	value := strings.Trim(string(i), `"`)
	if value == "" {
		return nil
	}
	time, err := time.Parse("2006-01-02", value)
	if err != nil {
		return err
	}
	*y = Year(time.Year())
	return nil
}

type Timestamp struct {
	time.Time
}

func (t *Timestamp) String() string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func (t *Timestamp) UnmarshalJSON(i []byte) error {
	value := strings.Trim(string(i), `"`)
	if value == "" {
		return nil
	}
	millis, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	time := time.UnixMilli(int64(millis))
	if err != nil {
		return err
	}
	t.Time = time
	return nil
}

type Film struct {
	FilmID      int       `json:"film_id"`
	Title       string    `json:"title"`
	ReleaseYear Year      `json:"release_year"`
	LastUpdate  Timestamp `json:"last_update"`
	UUID        string    `json:"uuid"`
}

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
	payload := &struct {
		Payload *Film `json:"payload"`
	}{}
	if err := json.NewDecoder(r).Decode(payload); err != nil {
		return err
	}
	return insertOrUpdate(payload.Payload)
}

func insertOrUpdate(film *Film) error {
	ctx, cancel := dbConn.CreateContext(context.TODO())
	defer cancel()
	var id int
	err := dbConn.DB().QueryRowContext(ctx, getFilmByUUIDSQL, film.UUID, film.FilmID).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return insert(film)
	case err != nil:
		return fmt.Errorf("an error occured while searching: %w", err)
	default:
		return update(film)
	}
}

func insert(film *Film) error {
	ctx, cancel := dbConn.CreateContext(context.TODO())
	defer cancel()
	filmUUID := uuid.New().String()
	res, err := dbConn.DB().ExecContext(ctx, insertFilmSQL, film.FilmID, filmUUID, film.Title, film.ReleaseYear, film.LastUpdate.Time, film.UUID)
	if err != nil {
		return fmt.Errorf("an error occured while inserting: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("an error occured while inserting: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("the film with external ID %d was not inserted", film.FilmID)
	}
	return nil
}

func update(film *Film) error {
	ctx, cancel := dbConn.CreateContext(context.TODO())
	defer cancel()
	res, err := dbConn.DB().ExecContext(ctx, updateFilmSQL, film.Title, film.ReleaseYear, film.FilmID, film.UUID)
	if err != nil {
		return fmt.Errorf("an error occured while updating: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("an error occured while updating: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("the film with external ID %d was not updated", film.FilmID)
	}
	return nil
}

func main() {

	flag.Parse()
	config := loadConfigurations()
	dbConn = createDBConnection(config.DB())

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	kafkaClient := kafka.NewClient(config.Kafka(), "films")

	ctx := context.Background()

	go func() {

		for {
			err := kafkaClient.Read(ctx, readFilm)
			if err != nil {
				log.Println(err)
			}
		}

	}()

	log.Println("catalogue consumer started")

	<-exit
	log.Println("server stopped")

	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		kafkaClient.Close()
		cancel()
	}()

	log.Println("consumer shutdown successfully")
}
