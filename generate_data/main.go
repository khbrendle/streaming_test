package main

import (
	"bytes"
	"encoding/csv"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/rs/zerolog"
)

type Customer struct {
	CustomerID    int       `json:"customer_id" gorm:"column:customer_id"`
	NameFirst     string    `json:"name_first" gorm:"column:name_first"`
	NameLast      string    `json:"name_last" gorm:"column:name_last"`
	StreetAddress string    `json:"street_address" gorm:"column:street_address"`
	City          string    `json:"city" gorm:"column:city"`
	State         string    `json:"state" gorm:"column:state"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt     time.Time `json:"deleted_at" gorm:"column:deleted_at"`
}

type Order struct {
	OrderID    int       `json:"order_id" gorm:"column:order_id"`
	CustomerID int       `json:"customer_id" gorm:"column:customer_id"`
	Product    string    `json:"product" gorm:"column:product"`
	Quantity   int       `json:"quantity" gorm:"column:quantity"`
	UnitPrice  string    `json:"unit_price" gorm:"column:unit_price"`
	SalesPrice float32   `json:"sales_price" gorm:"column:sales_price"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt  time.Time `json:"deleted_at" gorm:"column:deleted_at"`
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

	logger = zerolog.New(logFile).With().Timestamp().Str("service", "generate_data").Logger()
	pidFile := prog + ".pid"
	pid := os.Getpid()
	logger.Printf("running on pid %d", pid)

	err = ioutil.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0600)
	if err != nil {
		logger.Print("error writing pid file: " + err.Error())
		os.Exit(1)
	}

	// 2 main processes so going to keep track of how many completed
	doneCount := 0
	doneChan := make(chan int, 2)

	logger.Print("establishing db connection")
	db, err := gorm.Open("postgres", "host=localhost port=5432 dbname=postgres user=postgres password=webapp sslmode=disable")
	if err != nil {
		logger.Print(err)
		os.Exit(1)
	}
	defer db.Close()

	var timeMultiplier time.Duration = 1
	// var timeMultiplier time.Duration = 6

	createdCustomers := make(chan int, 100)

	// one thread will add customers every so often
	customersErr := make(chan error, 1)
	go func(db *gorm.DB, errChan chan<- error, sendCust chan<- int, doneChan chan<- int) {

		// read in the fake data
		rawCustomers, err := ioutil.ReadFile("./customers.csv")
		if err != nil {
			errChan <- err
		}
		var customers []*Customer
		r := csv.NewReader(bytes.NewReader(rawCustomers))

		// skip header
		_, err = r.Read()
		if err != nil {
			errChan <- err
		}

		var record []string
		for {
			record, err = r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				errChan <- err
			}

			customers = append(customers, &Customer{
				NameLast:      record[1],
				NameFirst:     record[0],
				StreetAddress: record[2],
				City:          record[3],
				State:         record[4],
			})
		}

		var waitTime = time.Second * 5 * timeMultiplier
		var threshold float32 = .01
		var tmpRes Customer
		for _, c := range customers {
			time.Sleep(waitTime)
			// every so often
			logger.Print("creating customer")
			err = db.Raw(`insert into customer (name_first, name_last, street_address, city, state)
      values (?, ?, ?, ?, ?) returning customer_id`, c.NameFirst, c.NameLast, c.StreetAddress, c.City, c.State).Scan(&tmpRes).Error
			if err != nil {
				errChan <- err
			}
			// send the customer ID to the orders thread
			sendCust <- tmpRes.CustomerID
			// every so often, increase liklihood of reducing wait time
			if rand.Float32() < threshold && waitTime > time.Second*5 {
				waitTime -= time.Second
				logger.Printf("decreasing customer creation wait time to %.2f seconds", waitTime.Seconds())
				// threshold += 0.01
			}
		}

		doneChan <- 1
	}(db, customersErr, createdCustomers, doneChan)

	// another thread will creat orders every so often
	// this will require a customer fk so will want the create customer
	// to return the ID so that it can be used in this thread
	ordersErr := make(chan error, 1)
	go func(db *gorm.DB, errChan chan<- error, getCustomers <-chan int, doneChan chan<- int) {

		// read in orders data
		rawOrders, err := ioutil.ReadFile("./orders.csv")
		if err != nil {
			errChan <- err
		}

		// deserialize
		var orders []*Order
		r := csv.NewReader(bytes.NewReader(rawOrders))

		// skip header
		_, err = r.Read()
		if err != nil {
			errChan <- err
		}

		// read in the order records
		var record []string
		var q int
		for {
			record, err = r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				errChan <- err
			}

			q, err = strconv.Atoi(record[1])
			if err != nil {
				errChan <- err
			}

			orders = append(orders, &Order{
				Product:   record[0],
				Quantity:  q,
				UnitPrice: record[2],
			})
		}

		var nCustomers int
		var knownCustomers []*int
		var order *Order
		var threshold float32 = .3
		var orderTime = time.Second * 8 * timeMultiplier
		var submitOrder = make(chan *Order, 1)
		var finalOrder = make(chan bool, 1)
		// timer thread
		go func(submitChan chan<- *Order, finalOrder chan<- bool) {
			// run the order timer
			// start with an order every 45 seconds, then decreasing time delta
			for _, order := range orders {
				// wait to submit order
				time.Sleep(orderTime)
				// select random customer
				order.CustomerID = *knownCustomers[rand.Intn(nCustomers)]
				// send order to be submitted
				submitChan <- order
				// optinally reduce wait time
				if rand.Float32() < threshold && orderTime > time.Second*2 {
					orderTime -= time.Second
					logger.Printf("decreasing order creation wait time to %.2f seconds", orderTime.Seconds())
					threshold += 0.05
				}
			}
			finalOrder <- true
		}(submitOrder, finalOrder)

	PlaceOrder:
		for {
			select {
			case customer := <-getCustomers:
				// add known customer ID
				nCustomers++
				knownCustomers = append(knownCustomers, &customer)
			case order = <-submitOrder:
				// submit order
				logger.Print("submitting order")
				err = db.Exec(`insert into "order" (customer_id, product, quantity, unit_price) values (?, ?, ?, ?)`,
					order.CustomerID, order.Product, order.Quantity, order.UnitPrice).Error
				if err != nil {
					errChan <- err
				}
			}
			// using second select to check if that was the final order
			select {
			case <-finalOrder:
				doneChan <- 1
				break PlaceOrder
			default:
				continue
			}
		}
	}(db, ordersErr, createdCustomers, doneChan)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	// this will block until context is cancelled or program cancelled
	exitStatus := 0

	// since we have 2 threads waiting to finish, will want to loop across the select.
	// if we receive an error or termination signal will break the loop & exit
	// otherwise if a process completes, increment the counter and wait for both to
	// be completed before exiting successfully
WaitLoop:
	for {
		select {
		case inc := <-doneChan:
			doneCount += inc
			if doneCount == 2 {
				exitStatus = 0
				break WaitLoop
			}
		case <-sigterm:
			logger.Print("got signal to cancel")
			break WaitLoop
		case err = <-customersErr:
			logger.Print("customers generation process error: " + err.Error())
			exitStatus = 1
			break WaitLoop
		case err = <-ordersErr:
			logger.Print("orders generation process error: " + err.Error())
			exitStatus = 1
			break WaitLoop
		}
	}

	// clean up pid file
	err = os.Remove(pidFile)
	if err != nil {
		logger.Print("error removing pid file: " + err.Error())
		os.Exit(1)
	}

	os.Exit(exitStatus)
}
