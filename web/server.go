package web

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"

	r "statistics/pkg/repository"
)

// Server is ...
type Server struct {
	// Port is the local machine TCP Port to bind the HTTP Server to
	Port    string
	Timeout Timeout
}

// Timeout is ...
type Timeout struct {
	// Server is the general server timeout to use
	// for graceful shutdowns
	Server time.Duration

	// Write is the amount of time to wait until an HTTP server
	// write opperation is cancelled
	Write time.Duration

	// Read is the amount of time to wait until an HTTP server
	// read operation is cancelled
	Read time.Duration

	// Read is the amount of time to wait
	// until an IDLE HTTP session is closed
	Idle time.Duration
}

// Database is ...
type Database struct {
	User     string
	Password string
	Dbname   string
	Host     string
	Port     string
}

// Config struct for webapp config
type Config struct {
	Server   Server
	Database Database
}

// NewConfig returns a new decoded Config struct
func NewConfig() (*Config, error) {
	// Create config structure
	server, _ := strconv.Atoi(os.Getenv("SERVER"))
	write, _ := strconv.Atoi(os.Getenv("WRITE"))
	read, _ := strconv.Atoi(os.Getenv("READ"))
	idle, _ := strconv.Atoi(os.Getenv("IDLE"))

	config := &Config{
		Server: Server{
			Port: os.Getenv("PORT"),
			Timeout: Timeout{
				Server: time.Duration(server),
				Write:  time.Duration(write),
				Read:   time.Duration(read),
				Idle:   time.Duration(idle),
			},
		},
		Database: Database{
			User:     os.Getenv("MYSQL_USER"),
			Password: os.Getenv("MYSQL_PASSWORD"),
			Dbname:   os.Getenv("MYSQL_DATABASE"),
			Host:     os.Getenv("DATABASE_HOST"),
			Port:     os.Getenv("MYSQL_PORT"),
		},
	}
	return config, nil
}

// NewRouter generates the router used in the HTTP Server
func NewRouter(w WebserviceHandler) *mux.Router {
	// Create router and define routes and return that router
	r := mux.NewRouter()
	r.HandleFunc("/stats", w.PostStats).Methods("POST")
	r.HandleFunc("/stats", w.GetStats).Methods("GET")
	r.HandleFunc("/stats", w.ClearStats).Methods("DELETE")
	r.Use(w.ValidationMiddleware)

	return r
}

// Run will run the HTTP Server
func (config Config) Run(w WebserviceHandler) {
	// Set up a channel to listen to for interrupt signals
	var runChan = make(chan os.Signal, 1)

	// Set up a context to allow for graceful server shutdowns in the event
	// of an OS interrupt (defers the cancel just in case)
	ctx, cancel := context.WithTimeout(
		context.Background(),
		config.Server.Timeout.Server,
	)
	defer cancel()

	// Define server options
	server := &http.Server{
		Addr:         ":" + config.Server.Port,
		Handler:      NewRouter(w),
		ReadTimeout:  config.Server.Timeout.Read * time.Second,
		WriteTimeout: config.Server.Timeout.Write * time.Second,
		IdleTimeout:  config.Server.Timeout.Idle * time.Second,
	}

	// Handle ctrl+c/ctrl+x interrupt
	signal.Notify(runChan, os.Interrupt, syscall.SIGTSTP)

	// Alert the user that the server is starting
	log.Printf("Server is starting on /%s\n", server.Addr)

	// Run the server on a new goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				// Normal interrupt operation, ignore
			} else {
				log.Fatalf("Server failed to start due to err: %v", err)
			}
		}
	}()

	// Block on this channel listeninf for those previously defined syscalls assign
	// to variable so we can let the user know why the server is shutting down
	interrupt := <-runChan

	// If we get one of the pre-prescribed syscalls, gracefully terminate the server
	// while alerting the user
	log.Printf("Server is shutting down due to %+v\n", interrupt)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server was unable to gracefully shutdown due to err: %+v", err)
	}
}

// InitDB is connect to database and return handle
func (config Config) InitDB() WebserviceHandler {
	// "%s:%s@%s(%s:%s)/%s
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Dbname,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(10)
	err = db.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		panic(err)
	}
	log.Println("Connected to: ", dsn)
	sdb := &r.StatsDB{DB: db}
	return WebserviceHandler{Rep: sdb}
}
