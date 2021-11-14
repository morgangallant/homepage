package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	godotenv.Load()
}

func must(k string) string {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	panic("missing variable " + k)
}

func fallback(k, fv string) string {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	return fv
}

type config struct {
	databaseUrl string
	port        string
	whkey       string
}

var cfg *config

func getConfig() *config {
	if cfg == nil {
		cfg = &config{
			databaseUrl: must("DATABASE_URL"),
			port:        fallback("PORT", "8080"),
			whkey:       must("WEBHOOK_KEY"),
		}
	}
	return cfg
}

type LogEntry struct {
	ID        uint64 `gorm:"primaryKey"`
	CreatedAt time.Time
	Contents  string
}

var db *gorm.DB

func getDB() *gorm.DB {
	if db == nil {
		var err error
		db, err = gorm.Open(postgres.Open(getConfig().databaseUrl), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			panic(err)
		}
		if err := db.AutoMigrate(&LogEntry{}).Error; err != nil {
			panic(err)
		}
	}
	return db
}

//go:embed site
var efs embed.FS

func run() error {
	mux := http.NewServeMux()
	sub, err := fs.Sub(efs, "site")
	if err != nil {
		return err
	}
	mux.Handle("/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("/api/logs", logsHandler())
	mux.HandleFunc("/_wh/postmark", postmarkHandler())
	return (&http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%s", getConfig().port),
		Handler:      mux,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second * 10,
	}).ListenAndServe()
}

func logsHandler() http.HandlerFunc {
	type logEntry struct {
		Timestamp time.Time `json:"ts"`
		Contents  string    `json:"contents"`
	}
	type response struct {
		Logs []logEntry `json:"logs"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var logs []LogEntry
		if err := getDB().WithContext(r.Context()).Order("created_at desc").Find(&logs).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp := response{Logs: []logEntry{}}
		for _, l := range logs {
			resp.Logs = append(resp.Logs, logEntry{
				Timestamp: l.CreatedAt,
				Contents:  l.Contents,
			})
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func postmarkHandler() http.HandlerFunc {
	type request struct {
		From     string `json:"From"`
		TextBody string `json:"TextBody"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if key := r.URL.Query().Get("key"); key != getConfig().whkey {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if !(req.From == "morgan@morgangallant.com" || req.From == "morgan@operand.ai") {
			http.Error(w, "invalid sender", http.StatusBadRequest)
			return
		}
		if err := getDB().WithContext(r.Context()).Create(&LogEntry{
			Contents: req.TextBody,
		}).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error in run(): %+v", err)
	}
}
