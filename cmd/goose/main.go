package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/config"
	_ "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/migrations"
)

var (
	flags = flag.NewFlagSet("goose", flag.ExitOnError)
	dir   = flags.String("dir", "./migrations", "directory with migration files")
)

func main() {
	if err := flags.Parse(os.Args[0:]); err != nil {
		log.Fatal("flags parse error")
	}
	args := flags.Args()

	if len(args) < 2 {
		flags.Usage()
		return
	}

	command := args[1]

	config, err := config.New()

	if err != nil {
		log.Fatal("cannot init config")
	}

	dbstring := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost(),
		config.DBPort(),
		config.DBUser(),
		config.DBPassword(),
		config.DBName(),
	)

	db, err := goose.OpenDBWithDriver("postgres", dbstring)
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v\n", err)
		}
	}()

	arguments := []string{}
	if len(args) > 2 {
		arguments = append(arguments, args[2:]...)
	}

	if err := goose.Run(command, db, *dir, arguments...); err != nil {
		log.Fatalf("goose %v: %v", command, err)
	}
}
