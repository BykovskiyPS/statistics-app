package repository

import (
	"database/sql"
	"log"
)

// StatsRepository интерфейс, описывающий возможные
// действия с базой данных статистики
type StatsRepository interface {
	FindByDate(date string) (Data, error)
	Storage(data Data) error
	Update(data Data) error
	FindByPeriodDate(from, to string) ([]Data, error)
	DeleteFromRepository() (int, error)
}

// Data структура, приходящая с "верхнего" уровня (usecase).
// записывается в базу данных
type Data struct {
	Date   string
	Views  int
	Clicks int
	Cost   int
}

// StatsDB структура содержащая хэндлер базы данных и
// реализующая интерфейс StatsRepository
type StatsDB struct {
	DB *sql.DB
}

// FindByDate находит запись по заданной дате
func (h *StatsDB) FindByDate(date string) (Data, error) {
	data := Data{}

	err := h.DB.QueryRow(
		"SELECT dat, clicks, views, cost "+
			"FROM stat WHERE dat = ?;",
		date).
		Scan(&data.Date, &data.Clicks, &data.Views, &data.Cost)
	if err != nil {
		return data, err
	}
	return data, nil
}

func checkError(method string, err error) error {
	if err != nil {
		log.Printf("Rep. %s: %v", method, err)
		return err
	}
	return nil
}

// Storage записывает в таблицу входные данные
func (h *StatsDB) Storage(data Data) error {
	_, err := h.DB.Exec(
		"INSERT INTO stat (dat, clicks, cost, views) VALUES (?, ?, ?, ?);",
		data.Date,
		data.Clicks,
		data.Cost,
		data.Views,
	)
	return checkError("Storage", err)
}

// Update обновляет запись в таблице с уже существующей датой
func (h *StatsDB) Update(data Data) error {
	_, err := h.DB.Exec(
		"UPDATE stat SET clicks = ?, cost = ?, views = ? WHERE dat = ?;",
		data.Clicks,
		data.Cost,
		data.Views,
		data.Date,
	)
	return checkError("Update", err)
}

// FindByPeriodDate находит записи, которые >= from и <= to
// Возвращает все поля
func (h *StatsDB) FindByPeriodDate(from, to string) ([]Data, error) {
	result := []Data{}
	rows, err := h.DB.Query(
		"SELECT dat, clicks, cost, views FROM stat WHERE dat >= ? AND dat <= ?;",
		from,
		to,
	)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		row := &Data{}
		err = rows.Scan(&row.Date, &row.Clicks, &row.Cost, &row.Views)
		if err != nil {
			log.Println("Rep. FindByPeriodDate: ", err)
			return nil, err
		}
		result = append(result, *row)
	}
	rows.Close()
	return result, nil
}

// DeleteFromRepository очищает таблицу
// и возвращаем количество удаленных строк
func (h *StatsDB) DeleteFromRepository() (int, error) {
	result, err := h.DB.Exec("DELETE FROM stat;")
	rows, err := result.RowsAffected()
	if err != nil {
		log.Println("Rep. DeleteFromRepository: ", err)
		return 0, err
	}
	return int(rows), nil
}
