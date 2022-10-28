package main

import (
	"fmt"
	"log"

	"github.com/danvergara/seeder"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/config"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/seeds"
)

func main() {
	cfg, err := config.New()

	if err != nil {
		log.Fatal("cannot init config")
	}

	dbstring := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost(),
		cfg.DBPort(),
		cfg.DBUser(),
		cfg.DBPassword(),
		cfg.DBName(),
	)

	db, err := sqlx.Open("postgres", dbstring)
	if err != nil {
		log.Fatalf("error opening a connection with the database %s\n", err)
	}

	s := seeds.NewSeed(db, cfg)

	if err := seeder.Execute(s); err != nil {
		log.Fatalf("error seeding the db %s\n", err)
	}
}
