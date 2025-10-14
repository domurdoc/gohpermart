package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/domurdoc/gophermart/internal/utils"
)

type Options struct {
	RunAddress           string        `env:"RUN_ADDRESS"`
	DatabaseURI          string        `env:"DATABASE_URI"`
	AccrualSystemAddress string        `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LogLevel             string        `env:"LOG_LEVEL"`
	JWTSecret            string        `env:"JWT_SECRET"`
	JWTDuration          time.Duration `env:"JWT_DURATION"`
	CookieName           string        `env:"COOKIE_NAME"`
	CookieMaxAge         time.Duration `env:"COOKIE_MAX_AGE"`

	// BonusService Params
	CheckedOrderBatchMaxSize  int
	CheckedOrderBatchInterval time.Duration
	UserBatchMaxSize          int
	UserBatchInterval         time.Duration
	CheckWorkers              int
	SaveWorkers               int

	// BonusClient Params
	PoolSize         int
	PoolTimeout      time.Duration
	MaxReries        int
	RetryWaitTime    time.Duration
	RetryMaxWaitTime time.Duration

	DebugClient bool
}

func New() *Options {
	options := Options{
		RunAddress:           "localhost:8000",
		DatabaseURI:          "postgresql://domurdoc@localhost/praktikum?sslmode=disable",
		AccrualSystemAddress: "localhost:8001",
		LogLevel:             "debug",
		JWTSecret:            utils.GenerateRandomString(utils.ALPHA, 32),
		JWTDuration:          600 * time.Second,
		CookieName:           "ilovesber",
		CookieMaxAge:         30 * time.Minute,

		CheckedOrderBatchMaxSize:  0, // for tests
		CheckedOrderBatchInterval: 10 * time.Second,
		UserBatchMaxSize:          0, // for tests
		UserBatchInterval:         10 * time.Second,
		CheckWorkers:              10,
		SaveWorkers:               10,

		PoolSize:         10,
		PoolTimeout:      60 * time.Second,
		MaxReries:        100,
		RetryWaitTime:    10 * time.Second,
		RetryMaxWaitTime: 60 * time.Second,

		DebugClient: false,
	}
	parseEnv(&options)
	parseArgs(&options)
	return &options
}

func parseEnv(options *Options) error {
	return env.Parse(options)
}

func parseArgs(options *Options) {
	flag.StringVar(&options.RunAddress, "a", options.RunAddress, "run address")
	flag.StringVar(&options.DatabaseURI, "d", options.DatabaseURI, "database uri")
	flag.StringVar(&options.AccrualSystemAddress, "r", options.AccrualSystemAddress, "accrual system address")
	flag.StringVar(&options.LogLevel, "l", options.LogLevel, "log level")
	flag.BoolVar(&options.DebugClient, "c", options.DebugClient, "use debug client")
	flag.Parse()
}
