package catalogue

import "time"

type Film struct {
	ID         int       `json:"-"`
	ExternalID int       `json:"-"`
	UUID       string    `json:"uuid"`
	Title      string    `json:"title"`
	Year       int       `json:"year"`
	LastUpdate time.Time `json:"-"`
}
