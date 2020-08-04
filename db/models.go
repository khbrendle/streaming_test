package main

import (
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"time"
)

type DateDimension struct {
	DateKey       time.Time `json:"date_key" gorm:"column:date_key"`
	TheDate       time.Time `json:"the_date" gorm:"column:the_date"`
	TheYear       uint16    `json:"the_year" gorm:"column:the_year"`
	QuarterNumber uint8     `json:"quarter_number" gorm:"column:quarter_number"`
	QuarterName   string    `json:"quarter_name" gorm:"column:quarter_name"`
	MonthNumber   uint8     `json:"month_number" gorm:"column:month_number"`
	MonthName     string    `json:"month_name" gorm:"column:month_name"`
	WeekOfYear    uint8     `json:"week_of_year" gorm:"column:week_of_year"`
	WeekdayNumber uint8     `json:"weekday_number" gorm:"column:weekday_number"`
	WeekdayName   string    `json:"weekday_name" gorm:"column:weekday_name"`
	DayOfMonth    uint8     `json:"day_of_month" gorm:"column:day_of_month"`
	DayOfYear     uint16    `json:"day_of_year" gorm:"column:day_of_year"`
	Weekend       bool      `json:"weekend" gorm:"column:weekend"`
}

func (dd *DateDimension) TableName() string {
	return `mart.date_dimension`
}

func NewDateDimension(t *time.Time) *DateDimension {
	var d DateDimension
	// var err error
	// d.DateKey, err = strconv.Atoi(t.Format("20060102"))
	// if err != nil {
	// 	panic(err)
	// }
	d.DateKey = *t

	d.TheDate = *t
	d.TheYear = uint16(t.Year())

	d.MonthNumber = uint8(t.Month())
	d.MonthName = t.Month().String()

	switch d.MonthNumber {
	case 1, 2, 3:
		d.QuarterNumber = 1
		d.QuarterName = "Q1"
	case 4, 5, 6:
		d.QuarterNumber = 2
		d.QuarterName = "Q2"
	case 7, 8, 9:
		d.QuarterNumber = 3
		d.QuarterName = "Q3"
	case 10, 11, 12:
		d.QuarterNumber = 4
		d.QuarterName = "Q4"
	}

	d.DayOfMonth = uint8(t.Day())
	d.DayOfYear = uint16(t.YearDay())

	d.WeekOfYear = uint8(d.DayOfYear / 7)
	d.WeekdayNumber = uint8(t.Weekday())
	d.WeekdayName = t.Weekday().String()

	d.Weekend = d.WeekdayNumber == 0 || d.WeekdayNumber == 6

	return &d
}

type TimeDimension struct {
	TimeKey   string `json:"time_key" gorm:"column:time_key"`
	TheTime24 string `json:"the_time_24" gorm:"column:the_time_24"`
	TheTime12 string `json:"the_time_12" gorm:"column:the_time_12"`
	Hour12    uint8  `json:"hour_12" gorm:"column:hour_12"`
	Hour24    uint8  `json:"hour_24" gorm:"column:hour_24"`
	TheMinute uint8  `json:"the_minute" gorm:"column:the_minute"`
	TheSecond uint8  `json:"the_second" gorm:"column:the_second"`
}

func (td *TimeDimension) TableName() string {
	return `mart.time_dimension`
}

func NewTimeDimension(t *time.Time) *TimeDimension {
	var td TimeDimension

	td.Hour24 = uint8(t.Hour())

	if td.Hour24 > 12 {
		td.Hour12 = td.Hour24 - 12
	} else {
		td.Hour12 = td.Hour24
	}

	td.TheMinute = uint8(t.Minute())
	td.TheSecond = uint8(t.Second())

	// td.TimeKey = uint32(td.Hour24)*10000 + uint32(td.TheMinute)*100 + uint32(td.TheSecond)
	td.TimeKey = fmt.Sprintf("%d:%d:%d", td.Hour24, td.TheMinute, td.TheSecond)
	td.TheTime24 = fmt.Sprintf("%d:%d:%d", td.Hour24, td.TheMinute, td.TheSecond)
	td.TheTime12 = fmt.Sprintf("%d:%d:%d", td.Hour12, td.TheMinute, td.TheSecond)

	return &td
}

type CreateMultiInsertQueryInput struct {
	Schema, Table string
	Cols          []string
	Vals          *[]interface{}
}

func templateSQLValues(td interface{}) (string, error) {
	tv := reflect.Indirect(reflect.ValueOf(td))
	nFields := tv.NumField()
	vals := make([]string, nFields)
	for i := 0; i < nFields; i++ {
		value := tv.Field(i)
		vals[i] = valueToSQL(value)
	}

	tmpl := `{{ range $i, $e := . }}{{ if gt $i 0}}, {{ end -}} {{ $e }} {{- end }}`

	t := template.Must(template.New("timeDimensionSQLValues").Funcs(template.FuncMap{
		"valueToSQL": valueToSQL,
	}).Parse(tmpl))

	var s strings.Builder

	if err := t.Execute(&s, vals); err != nil {
		return "", err
	}

	return s.String(), nil
}

func CreateMultiInsertQuery(inp *CreateMultiInsertQueryInput) (string, error) {
	// get column names
	tda := (*inp.Vals)[0]
	th := reflect.TypeOf(tda)
	nFields := th.NumField()
	inp.Cols = make([]string, nFields)
	var field reflect.StructField
	for i := 0; i < nFields; i++ {
		field = th.Field(i)

		inp.Cols[i] = strings.Split(field.Tag.Get("gorm"), ":")[1]
	}

	// log.Printf("cols: %v", inp.Cols)
	// log.Println("n records: ", len(*inp.Vals))

	tmpl := `insert into "{{ .Schema }}"."{{ .Table }}" ({{ range $i, $e := .Cols }}{{ if gt $i 0 }}, {{ end }}"{{ $e }}"{{ end }})
values {{ range $i, $e := .Vals}}
  {{ if gt $i 0}},{{ end }}({{ templateSQLValues $e }}){{ end }}`

	t := template.Must(template.New("mulitInsertQuery").Funcs(template.FuncMap{
		"templateSQLValues": templateSQLValues,
	}).Parse(tmpl))

	var s strings.Builder

	if err := t.Execute(&s, inp); err != nil {
		return "", err
	}

	return s.String(), nil
}

func valueToSQL(v reflect.Value) (val string) {
	vk := v.Kind()
	switch vk {
	case reflect.String:
		// log.Println("is string")
		val = fmt.Sprintf(`'%s'`, v.String())
	case reflect.Bool:
		// log.Println("is bool")
		if v.Bool() {
			val = "true"
		} else {
			val = "false"
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// log.Println("is int")
		val = fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// log.Println("is uint")
		val = fmt.Sprintf("%d", v.Uint())
	case reflect.Ptr:
		// log.Println("is Pointer")
		v = v.Elem()
		val = valueToSQL(v)
	case reflect.Invalid:
		// log.Println("is nil")
		val = "NULL"
	case reflect.Struct:
		switch v.Type().Name() {
		case "Time":
			val = v.Interface().(time.Time).String()
		default:
			logger.Print("unhandled struct type: " + v.Type().Name())
		}
	default:
		logger.Printf("unhandled type %+v", v.Type())
	}
	return
}
