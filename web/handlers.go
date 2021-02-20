package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"net/http"
	r "statistics/pkg/repository"
	uc "statistics/pkg/usecases"
	"statistics/pkg/validation"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/schema"
)

func toData(data validation.InputStat) r.Data {
	views, _ := strconv.Atoi(data.Views)
	clicks, _ := strconv.Atoi(data.Clicks)
	decimal, _ := strconv.ParseFloat(data.Cost, 64)
	cost := int(decimal * 100)
	return r.Data{
		Date:   data.Date,
		Views:  views,
		Clicks: clicks,
		Cost:   cost,
	}
}

// WebserviceHandler is ...
type WebserviceHandler struct {
	Rep r.StatsRepository
}

// ValidationMiddleware прослойка валидации входных параметров
// Исполняется до основного обрабочика
func (h *WebserviceHandler) ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := schema.NewDecoder()
		var params interface{}
		var err error
		switch r.Method {
		case http.MethodPost:
			params = &validation.InputStat{}
			r.ParseForm()
			err = decoder.Decode(params, r.PostForm)
		case http.MethodGet:
			params = &validation.Range{}
			err = decoder.Decode(params, r.URL.Query())
		case http.MethodDelete:
			params := r.URL.Query()
			if len(params) != 0 {
				err = errors.New("ValidationMiddleware: query isn't empty")
			}
		}
		if err != nil {
			http.Error(w, "ValidationMiddleware: bad keys in request", http.StatusBadRequest)
			return
		}
		_, err = govalidator.ValidateStruct(params)
		if err != nil {
			http.Error(w, "ValidationMiddleware: bad values in request", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// PostStats обработчик POST запроса. Запускает сценарий AddStat
func (h *WebserviceHandler) PostStats(w http.ResponseWriter, r *http.Request) {
	log.Println("POST request")
	r.ParseForm()
	msg := &validation.InputStat{}
	decoder := schema.NewDecoder()
	decoder.Decode(msg, r.PostForm)
	data := toData(*msg)
	if err := uc.AddStat(data, h.Rep); err != nil {
		log.Println("PostStats: ", err, data)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

// GetStats обработчик GET запроса. Запускает сценарий GetStatWithinFromAndTo
// Возвращает полученные данные в формате JSON
func (h *WebserviceHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	log.Println("GET request")
	msg := &validation.Range{}
	decoder := schema.NewDecoder()
	err := decoder.Decode(msg, r.URL.Query())

	data, err := uc.GetStatWithinFromAndTo(msg.From, msg.To, msg.OrderBy, h.Rep)
	if err != nil {
		log.Println("GetStats: ", err, data)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	result, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	log.Println("GetStats returned: ", string(result))
	fmt.Fprintln(w, string(result))
}

// ClearStats обработчик DELETE запроса. Запускает сценарий ClearRepository
func (h *WebserviceHandler) ClearStats(w http.ResponseWriter, r *http.Request) {
	log.Println("DELTE request")
	err := uc.ClearRepository(h.Rep)
	if err != nil {
		log.Println("ClearStats: ", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}
