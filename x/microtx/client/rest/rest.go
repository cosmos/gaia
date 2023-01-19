package rest

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/gorilla/mux"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx client.Context, r *mux.Router, storeName string) {

	// register the Data query handler
	r.HandleFunc(fmt.Sprintf("/%s/data/{%s}", storeName, "key"), DataHandler(cliCtx, storeName)).Methods("GET")
}
