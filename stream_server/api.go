package main

import (
	"errors"
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
	"github.com/rs/zerolog/log"
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

	api.RequestLogger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	r := mux.NewRouter()
	api.SubRouter = r.PathPrefix(fmt.Sprintf("/v%s/", api.Version)).Subrouter()
	api.AddRoutes()
	r.Use(api.LoggingMiddleware)

	var err error
	api.dm, err = gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
		"localhost", 5432, "postgres", "postgres", "webapp", "disable"))
	if err != nil {
		log.Fatal().Msg("error connecting to database: " + err.Error())
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
		panic(errors.New("error receiving RequestContext from http.Request"))
	}
	api.RequestLogger.Error().Str("request_id", rc.ID).Msgf(format, v...)
}

// warn

// info
func (api *API) reqLogInfo(r *http.Request, format string, v ...interface{}) {
	rc, ok := FromRequestContext(r.Context())
	if !ok {
		panic(errors.New("error receiving RequestContext from http.Request"))
	}
	api.RequestLogger.Info().Str("request_id", rc.ID).Msgf(format, v...)
}

// debug

// trace
func (api *API) reqLogTrace(r *http.Request, format string, v ...interface{}) {
	rc, ok := FromRequestContext(r.Context())
	if !ok {
		panic(errors.New("error receiving RequestContext from http.Request"))
	}
	api.RequestLogger.Trace().Str("request_id", rc.ID).Msgf(format, v...)
}
