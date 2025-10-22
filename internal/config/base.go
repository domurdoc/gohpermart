package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/domurdoc/gophermart/internal/utils"
)

type Options struct {
	Server       ServerOptions
	Repositories RepositoriesOptions
	Services     ServicesOptions
}

type ServerOptions struct {
	RunAddress string `env:"RUN_ADDRESS"`
}

type RepositoriesOptions struct {
	DatabaseURI string `env:"DATABASE_URI"`
}

type ServicesOptions struct {
	LogLevel             string `env:"LOG_LEVEL"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`

	// Auth
	JWTSecret    string        `env:"JWT_SECRET"`
	JWTDuration  time.Duration `env:"JWT_DURATION"`
	CookieName   string        `env:"COOKIE_NAME"`
	CookieMaxAge time.Duration `env:"COOKIE_MAX_AGE"`

	// BonusService Params
	SaverBatchMaxSize      int
	SaverBatchInterval     time.Duration
	RefresherBatchMaxSize  int
	RefresherBatchInterval time.Duration
	CheckerPoolSize        int
	SaverPoolSize          int

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
		Server: ServerOptions{
			RunAddress: "localhost:8000",
		},
		Repositories: RepositoriesOptions{
			DatabaseURI: "postgresql://domurdoc@localhost/praktikum?sslmode=disable",
		},
		Services: ServicesOptions{
			AccrualSystemAddress: "localhost:8001",
			LogLevel:             "debug",

			JWTSecret:    utils.GenerateRandomString(utils.ALPHA, 32),
			JWTDuration:  600 * time.Second,
			CookieName:   "ilovesber",
			CookieMaxAge: 30 * time.Minute,

			SaverBatchMaxSize:      0, // for tests
			SaverBatchInterval:     10 * time.Second,
			RefresherBatchMaxSize:  0, // for tests
			RefresherBatchInterval: 10 * time.Second,
			CheckerPoolSize:        10,
			SaverPoolSize:          10,

			PoolSize:         10,
			PoolTimeout:      60 * time.Second,
			MaxReries:        100,
			RetryWaitTime:    10 * time.Second,
			RetryMaxWaitTime: 60 * time.Second,

			DebugClient: false,
		},
	}

	parseEnv(&options)
	parseArgs(&options)
	return &options
}

func parseEnv(options *Options) error {
	return env.Parse(options)
}

func parseArgs(options *Options) {
	flag.StringVar(&options.Server.RunAddress, "a", options.Server.RunAddress, "run address")
	flag.StringVar(&options.Repositories.DatabaseURI, "d", options.Repositories.DatabaseURI, "database uri")
	flag.StringVar(&options.Services.AccrualSystemAddress, "r", options.Services.AccrualSystemAddress, "accrual system address")
	flag.StringVar(&options.Services.LogLevel, "l", options.Services.LogLevel, "log level")
	flag.BoolVar(&options.Services.DebugClient, "c", options.Services.DebugClient, "use debug client")
	flag.Parse()
}
