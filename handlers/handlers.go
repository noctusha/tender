package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/noctusha/tender/connection"
	"github.com/noctusha/tender/models"
)

const (
	statusCreated   = "CREATED"
	statusPublished = "PUBLISHED"
	statusClosed    = "CLOSED"
	statusCancelled = "CANCELLED"

	authorTypeUser         = "User"
	authorTypeOrganization = "Organization"

	serviceTypeConstruction = "Construction"
	serviceTypeDelivery     = "Delivery"
	serviceTypeManufacture  = "Manufacture"
)

type Handler struct {
	repo *connection.Repository
}

type JSON struct {
	Err     string           `json:"error,omitempty"`
	Tenders *[]models.Tender `json:"tender,omitempty"`
	Bids    *[]models.Bid    `json:"bid,omitempty"`
}

func NewHandler(repo *connection.Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

func respondJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, writeErr := w.Write([]byte("Internal server error"))
		if writeErr != nil {
			log.Printf("error writing an error in respondJSON: %v", writeErr)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	_, writeErr := w.Write(response)
	if writeErr != nil {
		log.Printf("error writing response in respondJSON: %v", writeErr)
	}
}

func respondJSONError(w http.ResponseWriter, statusCode int, message string) {
	respondJSON(w, statusCode, JSON{Err: message})
}

func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("ok"))
	if err != nil {
		log.Printf("error writing response in PingHandler: %v", err)
	}
}
