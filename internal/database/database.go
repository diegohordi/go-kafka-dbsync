package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/diegohordi/go-kafka/internal/configs"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type defaultConnection struct {
	db *sql.DB
}

type Connection interface {
	DB() *sql.DB
	CreateContext(ctx context.Context) (context.Context, context.CancelFunc)
	Close()
}

func NewConnection(dbConfig configs.DBConfigurer) (Connection, error) {
	db, err := sql.Open("mysql", dbConfig.DSN())
	if err != nil {
		return nil, fmt.Errorf("could not create a connection: %w", err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("database is not reachable: %w", err)
	}
	return &defaultConnection{db: db}, nil
}

func (d *defaultConnection) DB() *sql.DB {
	return d.db
}

func (d *defaultConnection) CreateContext(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout := 5 * time.Second
	return context.WithTimeout(ctx, timeout)
}

func (d *defaultConnection) Close() {
	if d.DB() == nil {
		return
	}
	if err := d.DB().Close(); err != nil {
		log.Printf("could not close the database connection %v\n", err)
		return
	}
	log.Printf("database connection released succesfully")
}
