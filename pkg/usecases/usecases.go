package usecases

import (
	"log"
	"math"
	"reflect"
	"sort"
	r "statistics/pkg/repository"
	"strings"
)

// OutputData структура, возврщаемая на "верхний" уровень (handlers).
// Формирутеся в usecase получения данных
type OutputData struct {
	Date   string
	Views  int
	Clicks int
	Cost   float64
	Cpc    float64
	Cpm    float64
}

// AddStat usecase сценарий добавления новой статистики или
// обновления уже существующей дате
// Параметры clicks, views прибавляются к уже существующим,
// а cost заменяется на новый
func AddStat(data r.Data, rep r.StatsRepository) error {
	st, err := rep.FindByDate(data.Date)
	if err != nil {
		// there is not date in db
		if err := rep.Storage(data); err != nil {
			log.Println("Usecase AddStat. Storage: ", err, data)
			return err
		}
	} else {
		st.Cost = data.Cost
		st.Clicks += data.Clicks
		st.Views += data.Views
		if err := rep.Update(st); err != nil {
			log.Println("Usecase AddStat. Update: ", err, data)
			return err
		}
	}
	return nil
}

// GetStatWithinFromAndTo сценарий, в котором возвращется статистика за даты между
// двумя заданными (from, to) и отсортированными по полю by
// Параметр by по умолчанию равен "date"
// Считаются поля cpc, cpm до 2х знаков после запятой
func GetStatWithinFromAndTo(from, to, by string, rep r.StatsRepository) ([]OutputData, error) {
	// сортировка по умолчанию
	if by == "" {
		by = "date"
	}
	by = strings.Title(by)
	var data []r.Data
	data, err := rep.FindByPeriodDate(from, to)
	if err != nil {
		log.Println("Usecase GetStatWithinFromAndTo. FindByPeriodDate: ", err)
		return nil, err
	}
	var result []OutputData
	tofloat := func(cost int) float64 {
		result := float64(cost) / 100
		return math.Round(result*100) / 100
	}
	cpc := func(cost float64, clicks int) float64 {
		if clicks == 0 {
			return 0.0
		}
		result := cost / float64(clicks)
		return math.Round(result*100) / 100
		// return float64(cost) / float64(clicks)
	}
	cpm := func(cost float64, views int) float64 {
		if views == 0 {
			return 0.0
		}
		result := (cost / float64(views)) * 1000
		return math.Round(result*100) / 100
		// return (float64(cost) / float64(views)) * 1000
	}
	for _, value := range data {
		newcost := tofloat(value.Cost)
		result = append(result, OutputData{
			Date:   value.Date,
			Views:  value.Views,
			Clicks: value.Clicks,
			Cost:   newcost,
			Cpc:    cpc(newcost, value.Clicks),
			Cpm:    cpm(newcost, value.Views),
		})
	}
	By(Prop(by, false)).Sort(result)
	return result, nil
}

// ClearRepository сценарий очистки таблицы
func ClearRepository(rep r.StatsRepository) error {
	return rep.TruncateRepository()
}

// Далее реализованы вспомогательные функции для сортировки по
// любому полю выходной структуры Output

// Prop is propeties function that auto choose field in struct
func Prop(field string, asc bool) func(p1, p2 *OutputData) bool {
	return func(p1, p2 *OutputData) bool {

		v1 := reflect.Indirect(reflect.ValueOf(p1)).FieldByName(field)
		v2 := reflect.Indirect(reflect.ValueOf(p2)).FieldByName(field)

		ret := false

		switch v1.Kind() {
		case reflect.Int64:
			ret = int64(v1.Int()) < int64(v2.Int())
		case reflect.Float64:
			ret = float64(v1.Float()) < float64(v2.Float())
		case reflect.String:
			ret = string(v1.String()) < string(v2.String())
		}

		if asc {
			return ret
		}
		return !ret
	}
}

// By is the type of a "less" function that defines the ordering of its Output arguments.
type By func(p1, p2 *OutputData) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(outputs []OutputData) {
	ps := &outputSorter{
		outputs: outputs,
		by:      by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

type outputSorter struct {
	outputs []OutputData
	by      func(p1, p2 *OutputData) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *outputSorter) Len() int { return len(s.outputs) }

// Swap is part of sort.Interface.
func (s *outputSorter) Swap(i, j int) {
	s.outputs[i], s.outputs[j] = s.outputs[j], s.outputs[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *outputSorter) Less(i, j int) bool {
	return s.by(&s.outputs[i], &s.outputs[j])
}
