package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"

	"github.com/mgajewskik/payment-platform/internal/domain/service"
	"github.com/mgajewskik/payment-platform/internal/domain/simulator"
	"github.com/mgajewskik/payment-platform/internal/env"
	"github.com/mgajewskik/payment-platform/internal/setup"
	"github.com/mgajewskik/payment-platform/internal/storage"
	"github.com/mgajewskik/payment-platform/internal/version"

	"github.com/lmittmann/tint"
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))

	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

type config struct {
	baseURL          string
	httpPort         int
	merchantID       string
	awsRegion        string
	awsDynamoDBTable string
	jwt              struct {
		secretKey string
	}
	setup bool
}

type application struct {
	config  config
	service *service.Service
	logger  *slog.Logger
	wg      sync.WaitGroup
}

func run(logger *slog.Logger) error {
	var cfg config

	cfg.baseURL = env.GetString("BASE_URL", "http://localhost:4444")
	cfg.httpPort = env.GetInt("HTTP_PORT", 4444)
	cfg.merchantID = env.GetString(
		"MERCHANT_ID",
		"test@merchant",
	) // NOTE: this is only used for token generation
	cfg.awsRegion = env.GetString("AWS_REGION", "us-east-1")
	cfg.awsDynamoDBTable = env.GetString("AWS_DYNAMODB_TABLE", "payment-platform-table")
	cfg.jwt.secretKey = env.GetString("JWT_SECRET_KEY", "dqohby7dgnt6dus6rnch26n3p6kwhsbn")
	cfg.setup = env.GetBool("SETUP", false)

	showVersion := flag.Bool("version", false, "display version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(cfg.awsRegion))
	if err != nil {
		return err
	}

	if cfg.setup {
		// NOTE: setting up the table and adding test data
		dbSetup := setup.NewDBSetup(awsCfg, cfg.awsDynamoDBTable)
		err = dbSetup.Setup()
		if err != nil {
			return err
		}
		defer dbSetup.Teardown()
	}

	storage := storage.NewDynamoDBRepository(cfg.awsDynamoDBTable, awsCfg, logger)
	bank := simulator.NewBankSimulator(logger)
	svc := service.NewService(storage, bank, logger)

	app := &application{
		config:  cfg,
		service: svc,
		logger:  logger,
	}

	return app.serveHTTP()
}
