package main

import (
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/alihasan00/gator/internal/config"
	"github.com/alihasan00/gator/internal/database"
)

func main() {
	s := &state{
		config: &config.Config{},
	}
	err := s.config.Read()
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("postgres", s.config.DBUrl)

	if err != nil {
		panic(err)
	}

	s.db = database.New(db)
	CommandHandler(s)
}

// goose postgres "postgres://alihasan:@localhost:5432/gator?sslmode=disable" up
