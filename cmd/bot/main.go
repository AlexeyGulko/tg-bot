package main

//todo возможно стоиит разнести на доммены т.е. /currency/storage /currency/service и т.д.
import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	_ "github.com/lib/pq"
	currencyClient "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/clients/currency"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/clients/tg"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/currency"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/hello"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/month_budget"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/report"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/spend"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/common/infrastructure/cache"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages/middleware"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/repository/command"
	currencyStorage "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/repository/currency"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/repository/spending"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/repository/user"
	currencyService "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/services/currency"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/workers/Queue"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/workers/http"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/workers/update_rates"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func main() {
	ctx, closeCtx := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer closeCtx()

	cfg := initConfig()
	logger.Init(cfg)
	iniTracing(cfg)
	db := initDb(cfg)
	rdb := initRDB(ctx, cfg)
	tgClient := initTg(cfg)
	ratesClient := currencyClient.New()

	// командам не нужно персистентное хранилище
	commandStorage := command.NewStorage()

	cacheSrv := cache.New(rdb)
	spendingStorage := spending.NewStorageWithCache(db, cacheSrv)
	userStorage := user.NewStorage(db)
	currencyStr := currencyStorage.NewStorageWithCache(db, cacheSrv)

	var updateRatesCh = make(chan update_rates.ChannelR)
	currSvc := currencyService.New(ratesClient, cfg, currencyStr, updateRatesCh)
	ratesWorker := update_rates.New(currSvc, cfg, updateRatesCh)
	httpWorker := http.New(cfg, userStorage, tgClient)
	BrokersList := []string{cfg.KafkaHost() + ":" + strconv.Itoa(int(cfg.KafkaPort()))}
	queueWorker, err := Queue.New(BrokersList)
	if err != nil {
		logger.Error("queue init failed", zap.Error(err))
	}
	queueCh := queueWorker.MessageChannel()

	baseMsg := messages.New(commandStorage)
	baseMsg.SetDefaultCommand(hello.NotFoundCommand(tgClient))
	baseMsg.SetStopCommand(hello.StopCommand(tgClient))
	baseMsg.AddCommand("/start", hello.Hello(tgClient, userStorage, cfg))
	baseMsg.AddCommand("/spend", spend.New(tgClient, spendingStorage, cfg, userStorage, currSvc))
	baseMsg.AddCommand("/help", hello.Help(tgClient))
	baseMsg.AddCommand(
		"/currency",
		currency.Menu(tgClient, cfg, userStorage).SetNext(currency.Input(tgClient, cfg, userStorage)),
	)
	baseMsg.AddCommand(
		"/budget",
		month_budget.Menu(tgClient, userStorage, cfg, currSvc).
			SetNext(month_budget.Input(tgClient, userStorage, cfg, currSvc)),
	)
	baseMsg.AddCommand("/report", report.New(tgClient, spendingStorage, cfg, userStorage, currSvc, queueCh))

	var decoratedMsg tg.Message
	decoratedMsg = middleware.NewLogger(baseMsg)
	decoratedMsg = middleware.NewTracer(decoratedMsg)
	decoratedMsg = middleware.NewMetrics(decoratedMsg)

	ratesWorker.Run(ctx)
	httpWorker.Run(ctx)
	queueWorker.Run(ctx)
	go tgClient.ListenUpdates(ctx, decoratedMsg)

	<-ctx.Done()
	logger.Info("app stopped")
}
