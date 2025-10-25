package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bandidos_do_byte/api/internal/config"
	"github.com/bandidos_do_byte/api/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/fx"
)

type Server struct {
	router *chi.Mux
	config *config.Config
}

func NewServer(config *config.Config, handler *handler.Handler) *Server {
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Registrar rotas
	handler.RegisterRoutes(r)

	return &Server{
		router: r,
		config: config,
	}
}

func (s *Server) Start(lifecycle fx.Lifecycle) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", s.config.Port),
		Handler: s.router,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				fmt.Printf("Server starting on port %s\n", s.config.Port)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					fmt.Printf("Server error: %v\n", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("Server stopping...")
			return srv.Shutdown(ctx)
		},
	})
}
