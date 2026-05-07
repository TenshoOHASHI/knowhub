package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
)

type config struct {
	DBHost       string
	DBPort       string
	DBName       string
	DBUser       string
	DBPassword   string
	MigrationDir string
}

func main() {
	cfg := loadConfig()

	if err := goose.SetDialect("mysql"); err != nil {
		log.Fatalf("set goose dialect: %v", err)
	}

	db, err := sql.Open("mysql", mysqlDSN(cfg))
	if err != nil {
		log.Fatalf("open mysql: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping mysql: %v", err)
	}

	command := "status"
	if len(os.Args) > 1 {
		command = strings.ToLower(os.Args[1])
	}

	switch command {
	case "status":
		err = goose.Status(db, cfg.MigrationDir)
	case "up":
		err = goose.Up(db, cfg.MigrationDir)
	case "up-by-one":
		err = goose.UpByOne(db, cfg.MigrationDir)
	case "down":
		err = goose.Down(db, cfg.MigrationDir)
	case "version":
		var version int64
		version, err = goose.GetDBVersion(db)
		if err == nil {
			fmt.Printf("goose db version: %d\n", version)
		}
	default:
		log.Fatalf("unknown command %q. use: status, up, up-by-one, down, version", command)
	}

	if err != nil {
		log.Fatalf("goose %s failed: %v", command, err)
	}
}

func loadConfig() config {
	cfg := config{
		DBHost:       env("MYSQL_HOST", "db"),
		DBPort:       env("MYSQL_PORT", "3306"),
		DBName:       os.Getenv("MYSQL_DATABASE"),
		DBUser:       os.Getenv("MYSQL_USER"),
		DBPassword:   os.Getenv("MYSQL_PASSWORD"),
		MigrationDir: env("MIGRATION_DIR", "/migrations"),
	}

	missing := make([]string, 0, 3)
	if cfg.DBName == "" {
		missing = append(missing, "MYSQL_DATABASE")
	}
	if cfg.DBUser == "" {
		missing = append(missing, "MYSQL_USER")
	}
	if cfg.DBPassword == "" {
		missing = append(missing, "MYSQL_PASSWORD")
	}
	if len(missing) > 0 {
		log.Fatalf("missing required env: %s", strings.Join(missing, ", "))
	}

	return cfg
}

func mysqlDSN(cfg config) string {
	mysqlCfg := mysql.Config{
		User:            cfg.DBUser,
		Passwd:          cfg.DBPassword,
		Net:             "tcp",
		Addr:            net.JoinHostPort(cfg.DBHost, cfg.DBPort),
		DBName:          cfg.DBName,
		ParseTime:       true,
		MultiStatements: true,
		Params: map[string]string{
			"charset": "utf8mb4",
		},
	}

	return mysqlCfg.FormatDSN()
}

func env(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
