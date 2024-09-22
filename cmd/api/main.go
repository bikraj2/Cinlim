package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"cinlim.bikraj.net/internal/data"
	"cinlim.bikraj.net/internal/jsonlog"
	"cinlim.bikraj.net/internal/mailer"
	_ "github.com/lib/pq"
)

var (
	buildTime string
	version   string
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		MaxOpenCons  int
		MaxIdleConns int
		MaxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config
	// Read values of port and environments from flag

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "dev", "enviorment", "Development | Stagling | Production")
	// Flags for Configuring Postgres Connecitons
	flag.IntVar(&cfg.db.MaxOpenCons, "maxOpenCons", 25, "Maximum Number of Open Connections")
	flag.IntVar(&cfg.db.MaxIdleConns, "maxIdleCons", 23, "Maximum Number of Open Idle Connections")
	flag.StringVar(&cfg.db.MaxIdleTime, "maxIdleTime", "15m", "Maximum Number of Open Idle Connections")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "Postgresql DSN")
	// Flag for Rate Limiters
	flag.BoolVar(&cfg.limiter.enabled, "rate-enabled", true, "Enable Rate Limitter")
	flag.IntVar(&cfg.limiter.burst, "burst", 4, "Rate Limiter maximum burst")
	flag.Float64Var(&cfg.limiter.rps, "limiter rps", 2, "Rate Limiter maximum burst")

	// Flag for Smtp Details
	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 587, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "1338148466653c", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "8af03244315cf7", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@bikraj.net>", "SMTP sender")
	flag.Func("cors-trusted-origin", "Trusted CORS orirgins (space seperated)", func(s string) error {
		cfg.cors.trustedOrigins = strings.Fields(s)
		return nil
	})

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()
	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build time:\t%s\n", buildTime)
		os.Exit(0)
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()

	logger.PrintInfo("Database connection established successfully", nil)

	models := data.NewModels(db)

	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))
	// Publish the database connection pool statistics.
	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))
	// Publish the current Unix timestamp.
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))
	expvar.NewString("Version").Set(version)
	app := &application{
		config: cfg,
		logger: logger,
		models: models,
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender)}
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

// Open DataBase Helper function
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.MaxOpenCons)
	db.SetMaxIdleConns(cfg.db.MaxIdleConns) // This should be lesser than  the MaxOpenCons
	duration, err := time.ParseDuration(cfg.db.MaxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
