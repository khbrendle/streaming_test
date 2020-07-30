package main

// AddRoutes attaches the routes to the server
// use to move the routes into their own file for easier maintenance
func (api *API) AddRoutes() {
	api.SubRouter.HandleFunc("/health", api.GetHealth).Methods("Get")

	// listen to data stream
	api.SubRouter.HandleFunc("/stream/subscribe/{topic}", api.StreamMessages).Methods("Get")
}
