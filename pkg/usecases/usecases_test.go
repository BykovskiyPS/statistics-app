package usecases

import (
	"errors"
	"math"
	r "statistics/pkg/repository"
	"testing"
)

type MockStatsDB struct {
	db *map[string]r.Data
}

// заглушка БД для тестирования Usecase
type MockDB map[string]r.Data

func (m *MockDB) FindByDate(date string) (r.Data, error) {
	result := r.Data{"2020-01-01", 10, 11, 12}
	if date == result.Date {
		return result, nil
	}
	return r.Data{}, errors.New("Date not found")
}

func (m *MockDB) Storage(data r.Data) error {
	(*m)[data.Date] = data
	return nil
}

func (m *MockDB) Update(data r.Data) error {
	(*m)[data.Date] = data
	return nil
}

func (m *MockDB) FindByPeriodDate(from, to string) ([]r.Data, error) {
	return []r.Data{
			{"2021-11-25", 112, 123, 166},
			{"2021-08-23", 51, 11, 440},
			{"2021-06-17", 18, 12, 120},
			{"2021-05-12", 12, 15, 16}},
		nil
}

func (m *MockDB) DeleteFromRepository() (int, error) {
	(*m) = make(map[string]r.Data)
	return 0, nil
}

func TestAddUsecase(t *testing.T) {
	// Инициализируем заглушку.
	// "2020-05-05": Data{"2020-05-05", 110, 111, 115}
	// m := &MockStatsDB{db: map[string]Data{"2020-01-01": Data{"2020-01-01", 10, 11, 12}}}
	m := &MockDB{"2020-01-01": r.Data{"2020-01-01", 10, 11, 12}}
	// Передает закглушку в usecase Add()
	AddStat(r.Data{"2020-01-01", 50, 120, 150}, m)

	// Проверяем значение по этой дате в бд
	var exp = r.Data{"2020-01-01", 60, 131, 150}

	// Проверяем, соответствует ли возвращаемое значение ожиданиям на основе
	// фальшивых входных данных.
	if (*m)["2020-01-01"] != exp {
		t.Fatalf("got %v; expected %v", (*m)["2020-01-01"], exp)
	}

	// Передаю новое значение
	AddStat(r.Data{"2020-05-05", 100, 101, 102}, m)
	var exp1 = r.Data{"2020-05-05", 100, 101, 102}
	if (*m)["2020-05-05"] != exp1 {
		t.Fatalf("got %v; expected %v", (*m)["2020-05-05"], exp1)
	}
}

func TestGetUsecase(t *testing.T) {
	m := &MockDB{}

	result, _ := GetStatWithinFromAndTo("2020-06-06", "2020-11-30", "date", m)

	cpc := func(cost, clicks int) float64 {
		result := float64(cost) / float64(clicks)
		return math.Round(result*100) / 100
		// return float64(cost) / float64(clicks)
	}
	cpm := func(cost, views int) float64 {
		result := (float64(cost) / float64(views)) * 1000
		return math.Round(result*100) / 100
		// return (float64(cost) / float64(views)) * 1000
	}

	expect := []OutputData{
		{"2021-11-25", 112, 123, 166, cpc(166, 123), cpm(166, 112)},
		{"2021-08-23", 51, 11, 440, cpc(440, 11), cpm(440, 51)},
		{"2021-06-17", 18, 12, 120, cpc(120, 12), cpm(120, 18)},
		{"2021-05-12", 12, 15, 16, cpc(16, 15), cpm(16, 12)},
	}

	for i, value := range result {
		if value != expect[i] {
			t.Fatalf("got %v; expected %v", value, expect[i])
		}
	}
}

func TestClearUsecase(t *testing.T) {
	m := &MockDB{"2020-01-01": r.Data{"2020-01-01", 10, 11, 12}}
	ClearRepository(m)
	if len(*m) != 0 {
		t.Fatalf("got %v; expected %v", len(*m), 0)
	}
}

func TestSortByFieldFunction(t *testing.T) {
	input := []OutputData{
		{Date: "2020-01-01",
			Views:  10,
			Clicks: 11,
			Cost:   153,
			Cpc:    48.32,
			Cpm:    65.33},
		{Date: "2020-01-12",
			Views:  21,
			Clicks: 8,
			Cost:   100,
			Cpc:    55.32,
			Cpm:    35.33,
		},
	}
	By(Prop("date", false)).Sort(input)
	if input[0].Date != "2020-01-12" {
		t.Fatalf("Sort by date: got %v; expected %s", input[0].Date, "2020-01-12")
	}
	By(Prop("clicks", false)).Sort(input)
	if input[0].Clicks != 11 {
		t.Fatalf("Sort by clicks: got %v; expected %d", input[0].Clicks, 11)
	}
	By(Prop("cpc", false)).Sort(input)
	if input[0].Cpc != 55.32 {
		t.Fatalf("Sort by cpc: got %v; expected %f", input[0].Cpc, 55.32)
	}
}
