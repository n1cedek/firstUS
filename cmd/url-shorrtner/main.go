package main

import (
	"awesomeProject1/iternal/config"
	"awesomeProject1/iternal/http-server/handlers/delete_url"
	"awesomeProject1/iternal/http-server/handlers/redirect"
	"awesomeProject1/iternal/http-server/handlers/url/save"
	mwLogger "awesomeProject1/iternal/http-server/middleware/logger"
	"awesomeProject1/iternal/lib/logger/handlers/slogpretty"
	"awesomeProject1/iternal/lib/logger/sl"
	"awesomeProject1/iternal/storage/sqlite"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
)

const (
	envLocal = "local"
	envProd  = "prod"
	envDev   = "dev"
)

func main() {

	//TODO: init config: cleanenv
	cfg := config.MustLoad()

	// TODO: init logger: slog
	log := setupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env)) //постоянный параметр(будет во всех строках)
	log.Info("starting url-shortener")
	log.Debug("debug logging are enabled")

	// TODO: init storage: sqlite
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("error to initializing sqlite storage", sl.Err(err))
		//падает если не инициализироваласб бд
		os.Exit(1)
	}
	//
	//id, err := storage.SaveURL("https://google.com", "google")
	//if err != nil {
	//	log.Error("error to save url", sl.Err(err))
	//	os.Exit(1)
	//}
	//log.Info("saved url", slog.Int64("id", id))
	//
	//id, err = storage.SaveURL("https://google.com", "google")
	//if err != nil {
	//	log.Error("error to save url", sl.Err(err))
	//	os.Exit(1)
	//}
	_ = storage

	// TODO: init router: chi, "chi render"
	router := chi.NewRouter()
	//middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		r.Post("/", save.New(log, storage))

		r.Delete("/{alias}", deleteurl.New(log, storage))

	})

	router.Get("/{alias}", redirect.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	//не дает коду идти дальше
	if err := srv.ListenAndServe(); err != nil {
		log.Error("error to start server")
	}
	log.Error("server stopped")
	// TODO: run server
}
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
