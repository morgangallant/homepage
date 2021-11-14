package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
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

//go:embed site
var efs embed.FS

func run() error {
	mux := http.NewServeMux()
	sub, err := fs.Sub(efs, "site")
	if err != nil {
		return err
	}
	mux.Handle("/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("/_wh/postmark", postmarkHandler())
	return (&http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%s", getConfig().port),
		Handler:      mux,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second * 10,
	}).ListenAndServe()
}

func postmarkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if key := r.URL.Query().Get("key"); key != getConfig().whkey {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("got body from postmark: %s\n", string(body))
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error in run(): %+v", err)
	}
}
