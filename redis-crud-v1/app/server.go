package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

// Server ...
type Server struct {
	Router      *mux.Router
	RedisClient *redis.Client
	TTL         int
}

// Initialize ...
func (s *Server) Initialize(opt *redis.Options) error {
	s.RedisClient = redis.NewClient(opt)
	_, err := s.RedisClient.Get(context.TODO(), "key").Result()
	if err != nil && err != redis.Nil {
		return err
	}

	ttlString := strings.TrimSpace(os.Getenv("TTL_SECS"))
	ttl, err := strconv.Atoi(ttlString)
	if err != nil {
		return fmt.Errorf("ttl should be a valid number")
	}

	s.TTL = ttl
	s.Router = mux.NewRouter()
	return nil
}

// InitializeRoutes ...
func (s *Server) InitializeRoutes() {

	s.Router.HandleFunc("/healthcheck", func(rw http.ResponseWriter, _ *http.Request) {
		rw.Write([]byte("Redis Crud v1"))
	}).Methods(http.MethodGet).Name("HealthCheck")

	s.Router.HandleFunc("/get/{key}", func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		val, err := s.RedisClient.Get(ctx, vars["key"]).Result()
		if err != nil {
			rw.WriteHeader(404)
			rw.Write([]byte(err.Error()))
			return
		}

		rw.Write([]byte(val))
		return
	}).Methods(http.MethodGet).Name("GetKey")

	s.Router.HandleFunc("/set/{key}/{value}", func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		val, err := s.RedisClient.Set(ctx, vars["key"], vars["value"], time.Duration(s.TTL)*time.Second).Result()
		if err != nil && err != redis.Nil {
			rw.WriteHeader(500)
			rw.Write([]byte(err.Error()))
			return
		}

		rw.Write([]byte(val))
		return
	}).Methods(http.MethodOptions, http.MethodPost).Name("SetKey")
}

// Run ...
func (s *Server) Run(addr string) {
	srv := &http.Server{
		Addr:    addr,
		Handler: s.Router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Print("server started on port ", addr)

	<-done
	log.Print("server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed:%+v", err)
	}
	log.Print("server exited properly")
}
