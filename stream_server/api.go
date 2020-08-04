package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

// API represents main program configuration
type API struct {
	// http server details, using http.Server to take advantage of built in
	// cancellation
	Server *http.Server
	// API version used for all endpoints, expects integer as string
	Version string
	// SubRouter contains the version prefix to be used by all routes
	SubRouter *mux.Router
	// CORS details
	AllowedHeaders []string
	AllowedMethods []string
	AllowedOrigins []string
	// Kafka configuration details
	Kafka
	// data ware
	dm *gorm.DB
	// RequestLogger
	RequestLogger zerolog.Logger
}

// Init should read a configuration to initialize the program
func (api *API) Init() error {
	api.Version = "0"
	// CORS options
	api.AllowedHeaders = []string{"X-Requested-With", "Content-Type", "Authorization"}
	api.AllowedMethods = []string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}
	api.AllowedOrigins = []string{"*"}

	reqLoggerFileName := "stream_server_requests.log"
	reqLoggerFile, err := os.OpenFile(reqLoggerFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Print("could not open request log file: " + err.Error())
		api.RequestLogger = logger
	} else {
		api.RequestLogger = zerolog.New(reqLoggerFile).With().Timestamp().Logger()
	}

	r := mux.NewRouter()
	api.SubRouter = r.PathPrefix(fmt.Sprintf("/v%s/", api.Version)).Subrouter()
	api.AddRoutes()
	r.Use(api.LoggingMiddleware)

	api.dm, err = gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
		"localhost", 5432, "postgres", "postgres", "webapp", "disable"))
	if err != nil {
		logger.Fatal().Msg("error connecting to database: " + err.Error())
		os.Exit(1)
	}

	api.Server = &http.Server{
		// TODO: get port from config file
		Addr:    "127.0.0.1:3000",
		Handler: handlers.CORS(handlers.AllowedHeaders(api.AllowedHeaders), handlers.AllowedMethods(api.AllowedMethods), handlers.AllowedOrigins(api.AllowedOrigins))(r),
		// TODO: will need to play with these timeouts because likely would want to allow
		// data streaming beyond 30 minutes
		WriteTimeout: 30 * time.Minute,
		ReadTimeout:  30 * time.Minute,
	}

	return nil
}

func (api *API) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		r = r.WithContext(NewRequestContext(r.Context(), &RequestContext{ID: xid.New().String()}))
		api.reqLogTrace(r, "request: %s", r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

// panic
// fatal

// error
func (api *API) reqLogError(r *http.Request, format string, v ...interface{}) {
	rc, ok := FromRequestContext(r.Context())
	if !ok {
		logger.Print("error receiving RequestContext from http.Request")
	} else {
		api.RequestLogger.Error().Str("request_id", rc.ID).Msgf(format, v...)
	}
}

// warn

// info
func (api *API) reqLogInfo(r *http.Request, format string, v ...interface{}) {
	rc, ok := FromRequestContext(r.Context())
	if !ok {
		logger.Print("error receiving RequestContext from http.Request")
	} else {
		api.RequestLogger.Info().Str("request_id", rc.ID).Msgf(format, v...)
	}
}

// debug

// trace
func (api *API) reqLogTrace(r *http.Request, format string, v ...interface{}) {
	rc, ok := FromRequestContext(r.Context())
	if !ok {
		logger.Print("error receiving RequestContext from http.Request")
	} else {
		api.RequestLogger.Trace().Str("request_id", rc.ID).Msgf(format, v...)
	}
}
