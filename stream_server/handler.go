package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// GetHealth just returns 200 if is accessible
func (api *API) GetHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (api *API) StreamMessages(w http.ResponseWriter, r *http.Request) {
	api.reqLogTrace(r, "handling stream request")
	vars := mux.Vars(r)
	topic := vars["topic"]

	f, ok := w.(http.Flusher)
	if !ok {
		msg := "Streaming unsupported!"
		api.reqLogError(r, msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	rc, ok := FromRequestContext(r.Context())

	// Listen to the closing of the http connection via the CloseNotifier
	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		api.reqLogTrace(r, "client closed request")
		// handle closing the connection
		err := api.Kafka.Unsubscribe(&rc.ID, &topic)
		if err != nil {
			api.reqLogError(r, err.Error())
			// client does not need to know this error
			// http.Error(w, "error detaching data source", http.StatusInternalServerError)
		}
		return
	}()

	// check if that Kafka consumer has been started for the topic
	api.reqLogTrace(r, "subscribing client to topic")
	api.Kafka.Subscribe(&rc.ID, &topic)

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	var res []struct {
		X time.Time `json:"x" gorm:"column:x"`
		Y int       `json:"y" gorm:"column:y"`
	}
	gMin, ok := r.URL.Query()["groupMinute"]
	if !ok {
		api.reqLogInfo(r, "groupMinute query parameter not found, defaulting to 1")
		gMin = []string{"1"}
	}
	groupMinute, err := strconv.Atoi(gMin[0])
	if err != nil {
		api.reqLogError(r, "error converting groupMinute "+vars["groupMinute"]+" to integer: "+err.Error())
	}
	err = api.dm.Raw(fmt.Sprintf(`
select ts x, sum(n) y
from (
	select
		date_key + make_time(
			extract(hour from date_key + time_key)::int,
			cast(floor(extract(minute from date_key + time_key) / %d) * %d as int),
			0
		) ts
		,n
	from mart.customer_fact cf2
) t
group by ts
order by ts`, groupMinute, groupMinute)).Scan(&res).Error
	if err != nil {
		api.reqLogError(r, err.Error())
	}

	b, err := json.Marshal(&res)
	if err != nil {
		api.reqLogError(r, err.Error())
	}
	log.Trace().Msg("initial data: " + string(b))
	fmt.Fprintf(w, "data: %s\n\n", string(b))
	f.Flush()

	// Don't close the connection, instead loop endlessly.
	for {

		// Read from our messageChan.
		msg, open := <-*api.Kafka.GetMessage(&rc.ID, &topic)

		if !open {
			// If our messageChan was closed, this means that the client has
			// disconnected.
			// this should not be closed during active request
			api.reqLogTrace(r, "Kafka message channel closed")
			break
		}

		// Write to the ResponseWriter, `w`.
		fmt.Fprintf(w, "data: %s\n\n", string(msg.Value))

		// Flush the response.  This is only possible if
		// the repsonse supports streaming.
		f.Flush()
	}

	// Done.
	api.reqLogTrace(r, "Finished HTTP request at %s", r.URL.Path)
}
