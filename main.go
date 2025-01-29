package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/noctusha/tender/connection"
	"github.com/noctusha/tender/handlers"
)

func main() {

	repo, err := connection.NewRepository()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer repo.Close()

	err = repo.InitSchema()
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	handler := handlers.NewHandler(repo)

	router := mux.NewRouter()

	router.Methods(http.MethodGet).Path("/api/ping").HandlerFunc(handler.PingHandler)

	router.Methods(http.MethodGet).Path("/api/tenders").HandlerFunc(handler.ListTenders)
	router.Methods(http.MethodPost).Path("/api/tenders/new").HandlerFunc(handler.NewTender)
	router.Methods(http.MethodGet).Path("/api/tenders/my").HandlerFunc(handler.MyTenders)
	router.Methods(http.MethodGet).Path("/api/tenders/{tenderId}/status").HandlerFunc(handler.GetTenderStatus)
	router.Methods(http.MethodPut).Path("/api/tenders/{tenderId}/status").HandlerFunc(handler.SetTenderStatus)
	router.Methods(http.MethodPatch).Path("/api/tenders/{tenderId}/edit").HandlerFunc(handler.EditTender)
	router.Methods(http.MethodPut).Path("/api/tenders/{tenderId}/rollback/{version}").HandlerFunc(handler.RollbackTender)

	router.Methods(http.MethodPost).Path("/api/bids/new").HandlerFunc(handler.NewBid)
	router.Methods(http.MethodGet).Path("/api/bids/my").HandlerFunc(handler.MyBids)
	router.Methods(http.MethodGet).Path("/api/bids/{tenderId}/list").HandlerFunc(handler.ListBidsByTenderId)
	router.Methods(http.MethodPatch).Path("/api/bids/{bidId}/edit").HandlerFunc(handler.EditBid)
	router.Methods(http.MethodPut).Path("/api/bids/{bidId}/rollback/{version}").HandlerFunc(handler.RollbackBid)

	fmt.Println("server is running")

	err = http.ListenAndServe(os.Getenv("SERVER_ADDRESS"), router)
	if err != nil {
		log.Fatalf("error handling Listen and Serve: %v", err)
	}

}
