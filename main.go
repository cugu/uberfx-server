package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/alecthomas/kong"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type CLI struct {
	WASIDir string `help:"Directory to store WASI modules" type:"path" default:"./wasi"`
	Address string `help:"Address to listen on" default:":8080"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	var cli CLI
	kong.Parse(&cli)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	ctx := context.Background()

	registry, err := NewRegistry(cli.WASIDir)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, server := range registry.Servers() {
		server := server

		wg.Add(1)

		go func() {
			if err := server.Serve(ctx, &wg); err != nil {
				log.Fatal(err)
			}
		}()
	}

	wg.Wait()

	r.HandleFunc("/*", endpoint(registry))

	return http.ListenAndServe(cli.Address, r)
}

func endpoint(registry *Registry) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		server, err := registry.Server(r.Host)
		if err != nil {
			log.Printf(err.Error())
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}

		if server.Address == "" {
			log.Printf("server address not set for %s", server.Name)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}

		u := &url.URL{
			Scheme: "http",
			Host:   server.Address,
		}

		httputil.NewSingleHostReverseProxy(u).ServeHTTP(w, r)
	}
}
