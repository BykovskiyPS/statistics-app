package main

import (
	"log"
	"statistics/web"
)

// POST   http://localhost:8080/stats?date=2021-02-02&views=10&clicks=10&cost=5.05
// GET    http://localhost:8080/stats?from=2021-02-02&to=2021-02-05
// DELETE http://localhost:8080/stats

// curl -X POST -d "date=2020-01-01&clicks=100" http://localhost:8080/stats
// curl -G -d "from=2020-01-01&to=2020-01-10&orderby=date" http://localhost:8080/stats
// curl -X DELETE http://localhost:8080/stats

func main() {
	cfg, err := web.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	// Init database
	hdl := cfg.InitDB()

	// Run the server
	cfg.Run(hdl)
}
