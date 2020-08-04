package main

import (
	"fmt"
	"testing"
	"time"
)

func TestNewDateDimension(t *testing.T) {
	dt := time.Date(2009, 11, 10, 0, 0, 0, 0, time.UTC)

	dim := NewDateDimension(&dt)

	if dim.DateKey != dt {
		t.Error("incorrect DateKey")
	}
	if dim.TheDate != dt {
		t.Error("incorrect TheDate")
	}
	if dim.TheYear != 2009 {
		t.Error("incorrect TheYear")
	}
	if dim.QuarterNumber != 4 {
		t.Error("incorrect QuarterNumber")
	}
	if dim.QuarterName != "Q4" {
		t.Error("incorrect QuarterName")
	}
	if dim.MonthNumber != 11 {
		t.Error("incorrect MonthNumber")
	}
	if dim.MonthName != "November" {
		t.Error("incorrect MonthName")
	}
	if dim.WeekOfYear != 44 {
		t.Error("incorrect WeekOfYear")
	}
	if dim.WeekdayNumber != 2 {
		t.Error("incorrect WeekdayNumber")
	}
	if dim.WeekdayName != "Tuesday" {
		t.Error("incorrect WeekdayName")
	}
	if dim.DayOfMonth != 10 {
		t.Error("incorrect DayOfMonth")
	}
	if dim.DayOfYear != 314 {
		t.Error("incorrect DayOfYear")
	}
	if dim.Weekend != false {
		t.Error("incorrect Weekend")
	}
}

func TestNewTimeDimension(t *testing.T) {
	dt := time.Date(2009, 11, 10, 13, 15, 30, 0, time.UTC)

	dim := NewTimeDimension(&dt)

	if dim.TimeKey != dt {
		t.Error("incorrect TimeKey")
	}
	if dim.Hour12 != 1 {
		t.Error("incorrect Hour12")
	}
	if dim.Hour24 != 13 {
		t.Error("incorrect Hour24")
	}
	if dim.TheMinute != 15 {
		t.Error("incorrect ")
	}
	if dim.TheSecond != 30 {
		t.Error("incorrect TheSecond")
	}
}

func TestCreateMultiInsertQuery(t *testing.T) {
	inp := CreateMultiInsertQueryInput{
		Schema: "mart",
		Table:  "date_dim",
		Cols:   []string{"time_key", "month"},
		Vals: &[]interface{}{
			TimeDimension{
				TimeKey:   131530,
				Hour12:    1,
				Hour24:    13,
				TheMinute: 15,
				TheSecond: 30,
			},
			TimeDimension{
				TimeKey:   131531,
				Hour12:    1,
				Hour24:    13,
				TheMinute: 15,
				TheSecond: 31,
			},
		},
	}
	q, err := CreateMultiInsertQuery(&inp)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("final query: " + q)
}

func TestTemplateSQLValues(t *testing.T) {
	td := TimeDimension{
		TimeKey:   131530,
		Hour12:    1,
		Hour24:    13,
		TheMinute: 15,
		TheSecond: 30,
	}

	s, err := templateSQLValues(td)
	if err != nil {
		t.Error(err)
	}

	if s != "131530, 1, 13, 15, 30" {
		t.Error("incorrect sql values")
	}
}
