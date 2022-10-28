package main

//todo возможно стоиит разнести на доммены т.е. /currency/storage /currency/service и т.д.
import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	currencyClient "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/clients/currency"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/clients/tg"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/currency"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/hello"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/month_budget"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/report"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/spend"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/config"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/repository/command"
	currencyStorage "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/repository/currency"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/repository/spending"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/repository/user"
	currencyService "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/services/currency"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/workers/update_rates"
	"golang.org/x/net/context"
)

func main() {
	ctx, closeCtx := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM,
	)
	defer closeCtx()

	config, err := config.New()
	if err != nil {
		log.Fatal("config init failed:", err)
	}

	dbconfig := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost(),
		config.DBPort(),
		config.DBUser(),
		config.DBPassword(),
		config.DBName(),
	)

	db, err := sql.Open("postgres", dbconfig)
	if err != nil {
		log.Fatal(err)
	}

	tgClient, err := tg.New(config)
	if err != nil {
		log.Fatal("tg client init failed:", err)
	}
	ratesClient := currencyClient.New()

	// командам не нужно персистентное хранилище
	commandStorage := command.NewStorage()
	spendingStorage := spending.NewStorage(db)
	userStorage := user.NewStorage(db)
	currencyStorage := currencyStorage.NewStorage(db)

	var updateRatesCh = make(chan update_rates.ChannelR)
	currSvc := currencyService.New(ratesClient, config, currencyStorage, updateRatesCh)
	ratesWorker := update_rates.New(currSvc, config, updateRatesCh)

	msgModel := messages.New(commandStorage)

	msgModel.SetDefaultCommand(hello.NotFoundCommand(tgClient))
	msgModel.SetStopCommand(hello.StopCommand(tgClient))
	msgModel.AddCommand("/start", hello.Hello(tgClient, userStorage, config))
	msgModel.AddCommand("/spend", spend.New(tgClient, spendingStorage, config, userStorage, currSvc))
	msgModel.AddCommand("/help", hello.Help(tgClient))
	msgModel.AddCommand(
		"/currency",
		currency.Menu(tgClient, config, userStorage).SetNext(currency.Input(tgClient, config, userStorage)),
	)
	msgModel.AddCommand(
		"/budget",
		month_budget.Menu(tgClient, userStorage, config, currSvc).
			SetNext(month_budget.Input(tgClient, userStorage, config, currSvc)),
	)
	msgModel.AddCommand("/report", report.New(tgClient, spendingStorage, config, userStorage, currSvc))

	ratesWorker.Run(ctx)
	tgClient.ListenUpdates(ctx, msgModel)

	go func() {
		<-ctx.Done()
		log.Println("app stopped")
	}()
}
