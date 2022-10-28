package seeds

import (
	"log"
	"math/rand"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type Seed struct {
	db     *sqlx.DB
	config Config
}

type Config interface {
	Currencies() []string
}

func NewSeed(db *sqlx.DB, cfg Config) Seed {
	return Seed{
		db:     db,
		config: cfg,
	}
}

/*
в пакете сидирования используется рефлексия, в ней команды сортируются по алфвавиту,
поэтому добавляем в название методов префикы
*/

func (s Seed) AUsersSeed() {
	currencies := s.config.Currencies()

	for i := 0; i < 50; i++ {
		currency := currencies[rand.Intn(len(currencies)-1)]
		var budget int64
		if rand.Float32() < 0.5 {
			budget = int64(rand.Intn(1000000))
		}

		builder := getBuilder().Insert("users").Columns(
			"created_at",
			"currency",
			"tg_id",
			"month_budget",
			"id",
		).Values(
			faker.Date(),
			currency,
			rand.Intn(1000000),
			budget,
			uuid.New(),
		)

		query, args, err := builder.ToSql()
		if err != nil {
			log.Fatalf("error seeding roles: %v", err)
		}

		_, err = s.db.Exec(query, args...)
		if err != nil {
			log.Fatalf("error seeding roles: %v", err)
		}
	}
}

type User struct {
	Id     int64
	Budget decimal.Decimal
}

func (s Seed) BSpendingSeed() {
	userQB := getBuilder().Select("id", "month_budget").From("users")

	query, args, err := userQB.ToSql()
	if err != nil {
		log.Fatalf("error seeding roles: %v", err)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Fatalf("error seeding roles: %v", err)
	}

	users := make([]User, 0, 50)

	var id int64
	var budget decimal.NullDecimal
	for rows.Next() {
		err := rows.Scan(&id, &budget)
		if err != nil {
			log.Fatalf("error seeding roles: %v", err)
		}

		users = append(users, User{Id: id, Budget: budget.Decimal})
	}

	for _, user := range users {
		count := rand.Intn(10) + 1
		for i := 0; i < count; i++ {

			date, err := time.Parse(faker.BaseDateFormat, faker.Date())
			if err != nil {
				log.Fatalf("error seeding roles: %v", err)
			}

			year, month, day := date.Date()
			date = time.Date(year, month, day, 0, 0, 0, 0, date.Location())

			builder := getBuilder().Insert("spendings").Columns(
				"user_id",
				"category",
				"amount",
				"date",
				"created_at",
				"id",
			).Values(
				user.Id,
				faker.Word(),
				rand.Intn(100000),
				date,
				faker.Date(),
				uuid.New(),
			)

			query, args, err := builder.ToSql()
			if err != nil {
				log.Fatalf("error seeding roles: %v", err)
			}

			_, err = s.db.Exec(query, args...)
			if err != nil {
				log.Fatalf("error seeding roles: %v", err)
			}
		}
	}
}

func getBuilder() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}
