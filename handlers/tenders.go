package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/noctusha/tender/models"
)

func (h *Handler) ListTenders(w http.ResponseWriter, r *http.Request) {
	var (
		limit       int
		offset      int
		serviceType string
		err         error
	)
	for name, vals := range r.URL.Query() {
		switch name {
		case "service_type":
			serviceType = vals[0]
		case "limit":
			limit, err = strconv.Atoi(vals[0])
			if err != nil {
				respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid limit format: %v", err))
				return
			}
		case "offset":
			offset, err = strconv.Atoi(vals[0])
			if err != nil {
				respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid offset format: %v", err))
				return
			}
		default:
			respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("unknown parameter: %s", name))
			return
		}
	}

	switch serviceType {
	case "", serviceTypeConstruction, serviceTypeDelivery, serviceTypeManufacture:
		break
	default:
		respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("unknown service type: %s", serviceType))
		return
	}

	tenders, err := h.repo.TendersList(serviceType, limit, offset)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed to select tender from database: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, JSON{Tenders: &tenders})
}

func (h *Handler) NewTender(w http.ResponseWriter, r *http.Request) {
	var tender models.Tender

	err := json.NewDecoder(r.Body).Decode(&tender)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse JSON format: %v", err))
		return
	}

	if tender.Name == "" {
		respondJSONError(w, http.StatusBadRequest, "missing tender title")
		return
	}

	tender.ID = uuid.New().String()

	tender.Status = statusCreated

	if tender.CreatorUserName == "" {
		respondJSONError(w, http.StatusBadRequest, "missing tender creatorUsername")
		return
	}

	organizationId, ok, err := h.repo.GetOrganizationIDByUsername(tender.CreatorUserName)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get organization: %v", err))
		return
	}

	if !ok {
		respondJSONError(w, http.StatusNotFound, "organization not found")
		return
	}

	if tender.OrganizationID != organizationId {
		respondJSONError(w, http.StatusForbidden, fmt.Sprintf("user %s does not belong to organization %s", tender.CreatorUserName, tender.OrganizationID))
		return
	}

	err = h.repo.NewTender(tender)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save tender: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, tender)
}

func (h *Handler) MyTenders(w http.ResponseWriter, r *http.Request) {
	var (
		limit    int
		offset   int
		username string
		err      error
	)
	for name, vals := range r.URL.Query() {
		switch name {
		case "username":
			username = vals[0]
		case "limit":
			limit, err = strconv.Atoi(vals[0])
			if err != nil {
				respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid limit format: %v", err))
				return
			}
		case "offset":
			offset, err = strconv.Atoi(vals[0])
			if err != nil {
				respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid offset format: %v", err))
				return
			}
		default:
			respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("unknown parameter: %s", name))
			return
		}
	}

	tenders, err := h.repo.MyTendersList(username, limit, offset)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "failed to select tender from database")
		return
	}

	respondJSON(w, http.StatusOK, JSON{Tenders: &tenders})
}

func (h *Handler) GetTenderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	tenderID, err := uuid.Parse(vars["tenderId"])
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid tenderID format")
		return
	}

	tender, ok, err := h.repo.GetTenderByID(tenderID)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get tender: %v", err))
		return
	}

	if !ok {
		respondJSONError(w, http.StatusNotFound, fmt.Sprintf("tender not found"))
		return
	}

	respondJSON(w, http.StatusOK, tender.Status)
}

func (h *Handler) EditTender(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	tenderID, err := uuid.Parse(vars["tenderId"])
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid tenderID format")
		return
	}

	var username string
	for name, vals := range r.URL.Query() {
		switch name {
		case "username":
			username = vals[0]
		default:
			respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("unknown parameter: %s", name))
			return
		}
	}

	if username == "" {
		respondJSONError(w, http.StatusBadRequest, "missing username")
		return
	}

	tender, ok, err := h.repo.GetTenderByID(tenderID)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get tender: %v", err))
		return
	}

	if !ok {
		respondJSONError(w, http.StatusNotFound, fmt.Sprintf("tender not found"))
		return
	}

	organizationId, userFound, err := h.repo.GetOrganizationIDByUsername(username)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get organization by username: %v", err))
		return
	}

	if !userFound {
		respondJSONError(w, http.StatusUnauthorized, fmt.Sprintf("user not found: %s", username))
		return
	}

	if tender.OrganizationID != organizationId {
		respondJSONError(w, http.StatusForbidden, fmt.Sprintf("user %s does not have permissions to this tender", username))
		return
	}

	var updatedTender models.Tender
	err = json.NewDecoder(r.Body).Decode(&updatedTender)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid JSON format")
		return
	}

	err = h.repo.AddTenderVersion(&models.TenderVersion{
		TenderID:    tender.ID,
		Name:        tender.Name,
		Description: tender.Description,
	})
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to add tender version: %v", err))
		return
	}

	if updatedTender.Name != "" {
		tender.Name = updatedTender.Name
	}
	if updatedTender.Description != "" {
		tender.Description = updatedTender.Description
	}

	err = h.repo.UpdateTender(tender)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to update tender: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, tender)
}

func (h *Handler) validateSetTenderStatus(status string, username string) error {
	if status == "" {
		return fmt.Errorf("status is mandatory")
	}

	if username == "" {
		return fmt.Errorf("username is mandatory")
	}

	switch status {
	case statusCreated, statusPublished, statusCancelled, statusClosed:
		break
	default:
		return fmt.Errorf("invalid status: %s", status)
	}

	return nil
}

func (h *Handler) SetTenderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var (
		status   string
		username string
	)
	for name, vals := range r.URL.Query() {
		switch name {
		case "status":
			status = strings.ToUpper(vals[0])
		case "username":
			username = vals[0]
		default:
			respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("unknown parameter: %s", name))
			return
		}
	}

	if err := h.validateSetTenderStatus(status, username); err != nil {
		respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("validation error: %v", err))
		return
	}

	tenderID, err := uuid.Parse(vars["tenderId"])
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid tenderID format")
		return
	}

	tender, ok, err := h.repo.GetTenderByID(tenderID)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get tender: %v", err))
		return
	}

	if !ok {
		respondJSONError(w, http.StatusNotFound, fmt.Sprintf("tender not found"))
		return
	}

	tender.Status = status

	organizationId, userFound, err := h.repo.GetOrganizationIDByUsername(username)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get organization by username: %v", err))
		return
	}

	if !userFound {
		respondJSONError(w, http.StatusUnauthorized, fmt.Sprintf("user not found: %s", username))
		return
	}

	if tender.OrganizationID != organizationId {
		respondJSONError(w, http.StatusForbidden, fmt.Sprintf("user %s does not have permissions to this tender", username))
		return
	}

	err = h.repo.UpdateTenderStatus(tender.ID, status)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to update tender status: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, tender)
}

func (h *Handler) RollbackTender(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	tenderID, err := uuid.Parse(vars["tenderId"])
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid tenderID format")
		return
	}

	var username string
	for name, vals := range r.URL.Query() {
		switch name {
		case "username":
			username = vals[0]
		default:
			respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("unknown parameter: %s", name))
			return
		}
	}

	if username == "" {
		respondJSONError(w, http.StatusBadRequest, "missing username")
		return
	}

	tender, ok, err := h.repo.GetTenderByID(tenderID)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get tender: %v", err))
		return
	}

	if !ok {
		respondJSONError(w, http.StatusNotFound, fmt.Sprintf("tender not found"))
		return
	}

	organizationId, userFound, err := h.repo.GetOrganizationIDByUsername(username)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get organization by username: %v", err))
		return
	}

	if !userFound {
		respondJSONError(w, http.StatusUnauthorized, fmt.Sprintf("user not found: %s", username))
		return
	}

	if tender.OrganizationID != organizationId {
		respondJSONError(w, http.StatusForbidden, fmt.Sprintf("user %s does not have permissions to this tender", username))
		return
	}

	version, err := uuid.Parse(vars["version"])
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid version format")
		return
	}

	tenderVer, err := h.repo.GetTenderVersionByID(version)
	if err != nil {
		respondJSONError(w, http.StatusNotFound, fmt.Sprintf("failed to get tender version: %v", err))
		return
	}

	if tenderVer.TenderID != tender.ID {
		respondJSONError(w, http.StatusForbidden, "tender version does not belong to tender")
		return
	}

	tender.Name = tenderVer.Name
	tender.Description = tenderVer.Description

	if err := h.repo.UpdateTender(tender); err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to update tender: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, tender)
}
