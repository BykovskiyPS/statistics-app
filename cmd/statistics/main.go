package main

import (
	"database/sql"
	"log"
	"net/http"
	r "statistics/pkg/repository"
	"statistics/web"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// POST   http://localhost:8080/stats?date=2021-02-02&views=10&clicks=10&cost=5.05
// GET    http://localhost:8080/stats?from=2021-02-02&to=2021-02-05
// DELETE http://localhost:8080/stats

// curl -X POST -d "date=2020-01-01&clicks=100" http://localhost:8080/stats
// curl -G -d "from=2020-01-01&to=2020-01-10&orderby=date" http://localhost:8080/stats
// curl -X DELETE http://localhost:8080/stats

func main() {
	// основные настройки к базе
	dsn := "pavel:pavel@tcp(localhost:3306)/stats?"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(10)
	err = db.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		panic(err)
	}

	sdb := &r.StatsDB{DB: db}
	w := web.WebserviceHandler{Rep: sdb}

	r := mux.NewRouter()
	r.HandleFunc("/stats", w.PostStats).Methods("POST")
	r.HandleFunc("/stats", w.GetStats).Methods("GET")
	r.HandleFunc("/stats", w.ClearStats).Methods("DELETE")
	r.Use(w.ValidationMiddleware)
	log.Println("starting server at :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
