package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
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

func main() {
	log.Printf("running on pid %d\n", os.Getpid())

	var err error
	log.Println("reading config file")
	b, err := ioutil.ReadFile("db_conf.yaml")
	if err != nil {
		log.Fatal("error reading config: " + err.Error())
	}

	var conf PostgresConf
	log.Println("parsing config file")
	if err = yaml.Unmarshal(b, &conf); err != nil {
		log.Fatal("error parsing config YAML: " + err.Error())
	}
	if conf.Host == nil {
		log.Fatal("missing db config param Host")
	}
	if conf.Port == nil {
		log.Fatal("missing db config param Port")
	}
	if conf.Database == nil {
		log.Fatal("missing db config param Database")
	}
	if conf.User == nil {
		log.Fatal("missing db config param User")
	}
	if conf.Password == nil {
		log.Fatal("missing db config param Password")
	}
	if conf.SSLMode == nil {
		log.Fatal("missing db config param SSLMode")
	}

	log.Println("connecting to database")
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
		*conf.Host, *conf.Port, *conf.User, *conf.Database, *conf.Password, *conf.SSLMode))
	if err != nil {
		log.Fatal("error connecting to database: " + err.Error())
	}
	defer db.Close()

	f, err := ioutil.ReadFile("create_tables.sql")
	if err != nil {
		log.Fatal("error reading sql file: " + err.Error())
	}

	err = db.Exec(string(f)).Error
	if err != nil {
		log.Fatal("error creating schema: " + err.Error())
	}

	log.Println("creating date dimension records")
	ts := time.Now()
	var dd *DateDimension
	for i := 0; i < 365; i++ {
		// create dimension information
		dd = NewDateDimension(&ts)
		// write to database
		err = db.Create(&dd).Error
		if err != nil {
			log.Fatal("error writing date dimension record: " + err.Error())
		}
		// increment day
		ts = ts.AddDate(0, 0, 1)
	}

	log.Println("creating time dimension records")
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
			log.Fatal("error creating insert query: " + err.Error())
		}
		// write to database
		err = db.Exec(q).Error
		if err != nil {
			log.Fatal("error writing date dimension record: " + err.Error())
		}
	}

}

func String(s string) *string {
	return &s
}
