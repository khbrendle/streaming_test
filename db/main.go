package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/rs/zerolog"
	yaml "gopkg.in/yaml.v1"
)

type PostgresConf struct {
	Host     *string `yaml:"host"`
	Port     *int    `yaml:"port"`
	Database *string `yaml:"database"`
	User     *string `yaml:"user"`
	Password *string `yaml:"password"`
	SSLMode  *string `yaml:"sslmode"`
}

var logger zerolog.Logger

func main() {
	prog := os.Args[0]
	logFileName := prog + ".log"
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("could not open log file: " + err.Error())
		os.Exit(1)
	}

	logger = zerolog.New(logFile).With().Timestamp().Str("service", "db_init").Logger()
	pidFile := prog + ".pid"

	pid := os.Getpid()
	logger.Printf("running on pid %d", pid)

	err = ioutil.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0600)
	if err != nil {
		logger.Print("error writing pid file: " + err.Error())
		os.Exit(1)
	}

	conf := "db_conf.yaml"
	doneChan := make(chan bool, 1)
	errChan := make(chan error, 1)
	go run(conf, doneChan, errChan)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	// this will block until context is cancelled or program cancelled
	exitStatus := 0
	select {
	case <-sigterm:
		logger.Print("got signal to cancel")
	case err = <-errChan:
		logger.Print("execution error: " + err.Error())
		exitStatus = 1
	case <-doneChan:
		logger.Print("process complete")
	}

	// clean up pid file
	err = os.Remove(pidFile)
	if err != nil {
		logger.Print("error removing pid file: " + err.Error())
		os.Exit(1)
	}

	os.Exit(exitStatus)
}

func String(s string) *string {
	return &s
}

func run(confFile string, done chan<- bool, errChan chan<- error) {
	logger.Print("reading config file")
	b, err := ioutil.ReadFile(confFile)
	if err != nil {
		errChan <- errors.New("error reading config: " + err.Error())
	}

	var conf PostgresConf
	logger.Print("parsing config file")
	if err = yaml.Unmarshal(b, &conf); err != nil {
		errChan <- errors.New("error parsing config YAML: " + err.Error())
	}
	if conf.Host == nil {
		errChan <- errors.New("missing db config param Host")
	}
	if conf.Port == nil {
		errChan <- errors.New("missing db config param Port")
	}
	if conf.Database == nil {
		errChan <- errors.New("missing db config param Database")
	}
	if conf.User == nil {
		errChan <- errors.New("missing db config param User")
	}
	if conf.Password == nil {
		errChan <- errors.New("missing db config param Password")
	}
	if conf.SSLMode == nil {
		errChan <- errors.New("missing db config param SSLMode")
	}

	logger.Print("connecting to database")
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
		*conf.Host, *conf.Port, *conf.User, *conf.Database, *conf.Password, *conf.SSLMode))
	if err != nil {
		errChan <- errors.New("error connecting to database: " + err.Error())
	}
	defer db.Close()

	f, err := ioutil.ReadFile("create_tables.sql")
	if err != nil {
		errChan <- errors.New("error reading sql file: " + err.Error())
	}

	err = db.Exec(string(f)).Error
	if err != nil {
		errChan <- errors.New("error creating schema: " + err.Error())
	}

	logger.Print("creating date dimension records")
	ts := time.Now()
	var dd *DateDimension
	for i := 0; i < 365; i++ {
		// create dimension information
		dd = NewDateDimension(&ts)
		// write to database
		err = db.Create(&dd).Error
		if err != nil {
			logger.Print("error writing date dimension record: " + err.Error())
		}
		// increment day
		ts = ts.AddDate(0, 0, 1)
	}

	logger.Print("creating time dimension records")
	ts = time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
	chunkSize := 800
	var td = make([]interface{}, chunkSize)
	var q string
	nRecords := 86_400 // seconds in a day
	for i := 0; i < nRecords; i += chunkSize {
		// create dimension information
		for ti := 0; ti < chunkSize; ti++ {
			if i+ti == nRecords {
				break
			}
			td[ti] = *NewTimeDimension(&ts)
			// increment day
			ts = ts.Add(time.Second)
		}
		// create insert query
		q, err = CreateMultiInsertQuery(&CreateMultiInsertQueryInput{
			Schema: "mart",
			Table:  "time_dimension",
			Vals:   &td,
		})
		if err != nil {
			errChan <- errors.New("error creating insert query: " + err.Error())
		}
		// write to database
		err = db.Exec(q).Error
		if err != nil {
			errChan <- errors.New("error writing date dimension record: " + err.Error())
		}
	}

	done <- true
}
