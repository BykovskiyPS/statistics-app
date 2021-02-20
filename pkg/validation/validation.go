package validation

import (
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
)

// InputStat структура для валидации входного POST запроса
type InputStat struct {
	Date   string `schema:"date" valid:"date"`
	Views  string `schema:"views" valid:"int, optional"`
	Clicks string `schema:"clicks" valid:"int, optional"`
	Cost   string `schema:"cost" valid:"cost, optional"`
}

// Range струкртура для валидации входного GET запроса
type Range struct {
	From    string `schema:"from" valid:"date"`
	To      string `schema:"to" valid:"date, isGreaterFrom"`
	OrderBy string `schema:"orderby" valid:"in(date|cost|views|clicks|cpm|cpc), optional"`
}

// Валидация DELETE запроса - это проверка,
// что параметры URL пустые

func init() {
	govalidator.SetFieldsRequiredByDefault(true)

	// Проверка, что поле date подходит под шаблон YYYY-MM-DD
	govalidator.TagMap["date"] = govalidator.Validator(func(str string) bool {
		layout := "2006-01-02"
		_, err := time.Parse(layout, str)
		if err != nil {
			return false
		}
		return true
	})

	// Проверка, что поле cost равно одному из значений:
	// cost=100; 11.10; 11.07
	// Целая часть рубли, а дробная - копейки
	govalidator.TagMap["cost"] = govalidator.Validator(func(cost string) bool {
		idx := strings.Index(cost, ".")
		if idx == 0 || idx == len(cost)-1 {
			return false
		}
		if idx == -1 {
			_, err := strconv.Atoi(cost)
			if err != nil {
				return false
			}
			return true
		}
		f := func(c rune) bool {
			return c == '.'
		}
		fields := strings.FieldsFunc(cost, f)
		rub, err1 := strconv.Atoi(fields[0])
		if len(fields[1]) >= 3 {
			return false
		}
		cop, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil {
			return false
		}
		if rub < 0 || cop < 0 {
			return false
		}
		return true
	})

	// Проверка, что поле from <= поля to
	govalidator.CustomTypeTagMap.Set("isGreaterFrom", func(i interface{}, context interface{}) bool {
		toDate := func(str string) time.Time {
			layout := "2006-01-02"
			res, _ := time.Parse(layout, str)
			return res
		}
		switch v := context.(type) {
		case Range:
			return toDate(v.From).Before(toDate(v.To)) ||
				toDate(v.From).Equal(toDate(v.To))
		}
		return false
	})
}
