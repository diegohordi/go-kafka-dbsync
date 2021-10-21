package catalogue

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/diegohordi/go-kafka/internal/database"
	"github.com/diegohordi/go-kafka/internal/kafka"
	"github.com/google/uuid"
	"time"
)

var ErrNoFilmFound = errors.New("no film found")

const getFilmByUUIDSQL = "select id, uuid, title, year from films where uuid = ?;"
const insertFilmSQL = "insert into films (uuid, title, year, last_update) values (?, ?, ?, ?)"
const updateFilmSQL = "update films set title = ?, year = ?, last_update = ? where id = ?;"

type Service struct {
	dbConn      database.Connection
	kafkaClient kafka.Client
}

func NewService(dbConn database.Connection, kafkaClient kafka.Client) *Service {
	return &Service{dbConn: dbConn, kafkaClient: kafkaClient}
}

func (s *Service) publishToSync(ctx context.Context, film Film) error {
	ctx, cancel := context.WithTimeout(ctx, 5 * time.Second)
	defer cancel()
	return s.kafkaClient.Write(ctx, film)
}

func (s *Service) InsertFilm(ctx context.Context, film Film) (Film, error) {
	dbCtx, cancel := s.dbConn.CreateContext(ctx)
	defer cancel()
	film.UUID = uuid.New().String()
	film.LastUpdate = time.Now()
	res, err := s.dbConn.DB().ExecContext(dbCtx, insertFilmSQL, film.UUID, film.Title, film.Year, film.LastUpdate)
	if err != nil {
		return Film{}, fmt.Errorf("an error occured while inserting: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return Film{}, fmt.Errorf("an error occured while inserting: %w", err)
	}
	if rows != 1 {
		return Film{}, fmt.Errorf("an unexpected error occured and the given film was not inserted")
	}
	err = s.publishToSync(ctx, film)
	if err != nil {
		return Film{}, err
	}
	return s.GetFilm(ctx, film.UUID)
}

func (s *Service) GetFilm(ctx context.Context, filmUUID string) (Film, error) {
	ctx, cancel := s.dbConn.CreateContext(context.TODO())
	defer cancel()
	var id int
	var rowUUID, title string
	var year int
	err := s.dbConn.DB().QueryRowContext(ctx, getFilmByUUIDSQL, filmUUID).Scan(&id, &rowUUID, &title, &year)
	if err == sql.ErrNoRows {
		return Film{}, ErrNoFilmFound
	}
	if err != nil {
		return Film{}, err
	}
	return Film{
		ID: id,
		UUID: rowUUID,
		Title: title,
		Year: year,
	}, nil
}

func (s *Service) UpdateFilm(ctx context.Context, filmUUID string, film Film) (Film, error) {
	existingFilm, err := s.GetFilm(ctx, filmUUID)
	if err != nil {
		return Film{}, err
	}
	film.UUID = filmUUID
	dbCtx, cancel := s.dbConn.CreateContext(ctx)
	defer cancel()
	res, err := s.dbConn.DB().ExecContext(dbCtx, updateFilmSQL, film.Title, film.Year, time.Now(), existingFilm.ID)
	if err != nil {
		return Film{}, fmt.Errorf("an error occured while inserting: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return Film{}, fmt.Errorf("an error occured while updating: %w", err)
	}
	if rows != 1 {
		return Film{}, fmt.Errorf("an unexpected error occured and the given film was not updated")
	}
	err = s.publishToSync(ctx, film)
	if err != nil {
		return Film{}, err
	}
	return s.GetFilm(ctx, filmUUID)
}
