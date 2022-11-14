package main

//todo возможно стоиит разнести на доммены т.е. /currency/storage /currency/service и т.д.
import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	jaggerConfig "github.com/uber/jaeger-client-go/config"
	currencyClient "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/clients/currency"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/clients/tg"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/currency"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/hello"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/month_budget"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/report"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands/spend"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/common/infrastructure/cache"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/config"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages/middleware"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/repository/command"
	currencyStorage "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/repository/currency"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/repository/spending"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/repository/user"
	currencyService "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/services/currency"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/workers/http"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/workers/update_rates"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func main() {
	ctx, closeCtx := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM,
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
	currencyStorage := currencyStorage.NewStorageWithCache(db, cacheSrv)

	var updateRatesCh = make(chan update_rates.ChannelR)
	currSvc := currencyService.New(ratesClient, cfg, currencyStorage, updateRatesCh)
	ratesWorker := update_rates.New(currSvc, cfg, updateRatesCh)
	httpWorker := http.New(cfg)

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
	baseMsg.AddCommand("/report", report.New(tgClient, spendingStorage, cfg, userStorage, currSvc))

	var decoratedMsg tg.Message
	decoratedMsg = middleware.NewLogger(baseMsg)
	decoratedMsg = middleware.NewTracer(decoratedMsg)
	decoratedMsg = middleware.NewMetrics(decoratedMsg)

	ratesWorker.Run(ctx)
	httpWorker.Run(ctx)
	//крутится в основной горутине
	tgClient.ListenUpdates(ctx, decoratedMsg)

	go func() {
		<-ctx.Done()
		log.Println("app stopped")
	}()
}

func initConfig() *config.Service {
	cfg, err := config.New()
	if err != nil {
		log.Fatal("config init failed:", err)
	}

	return cfg
}

func iniTracing(cfg *config.Service) {
	cfgJagger := jaggerConfig.Configuration{
		Sampler: &jaggerConfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaggerConfig.ReporterConfig{
			LocalAgentHostPort: cfg.JaegerHostPort(),
		},
	}

	_, err := cfgJagger.InitGlobalTracer(cfg.ServiceName())
	if err != nil {
		logger.Error("Cannot init tracing", zap.Error(err))
	}
}

func initDb(cfg *config.Service) *sql.DB {
	dbconfig := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost(),
		cfg.DBPort(),
		cfg.DBUser(),
		cfg.DBPassword(),
		cfg.DBName(),
	)

	db, err := sql.Open("postgres", dbconfig)
	if err != nil {
		logger.Fatal(err.Error())
	}

	return db
}

func initTg(cfg *config.Service) *tg.Client {
	client, err := tg.New(cfg)
	if err != nil {
		logger.Fatal("tg client init failed:", zap.Error(err))
	}

	return client
}

func initRDB(ctx context.Context, cfg *config.Service) *redis.Client {
	rdb := redis.NewClient(
		&redis.Options{
			Addr:     cfg.RedisHostPort(),
			Password: cfg.RedisPassword(),
			DB:       cfg.RedisDB(),
		})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		logger.Fatal("cannot init redis", zap.Error(err))
	}

	return rdb
}
