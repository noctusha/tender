package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/noctusha/tender/models"
)

func (h *Handler) validateNewBid(bid models.Bid) error {
	if bid.Name == "" {
		return fmt.Errorf("name is mandatory")
	}

	if bid.TenderID == "" {
		return fmt.Errorf("tenderId is mandatory")
	}

	if bid.AuthorType == "" {
		return fmt.Errorf("authorType is mandatory")
	}

	if bid.AuthorId == "" {
		return fmt.Errorf("authorId is mandatory")
	}

	switch bid.AuthorType {
	case authorTypeUser, authorTypeOrganization:
		break
	default:
		return fmt.Errorf("unknown author type: %s", bid.AuthorType)
	}

	return nil
}

func (h *Handler) NewBid(w http.ResponseWriter, r *http.Request) {
	var bid models.Bid

	err := json.NewDecoder(r.Body).Decode(&bid)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse JSON format: %v", err))
		return
	}

	if err := h.validateNewBid(bid); err != nil {
		respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("validation error: %v", err))
		return
	}

	bid.ID = uuid.New().String()
	bid.Status = statusCreated

	tenderID, err := uuid.Parse(bid.TenderID)
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

	switch bid.AuthorType {
	case authorTypeUser:
		orginazationId, ok, err := h.repo.GetOrganizationIDByUserID(bid.AuthorId)
		if err != nil {
			respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get organization: %v", err))
			return
		}

		if !ok {
			respondJSONError(w, http.StatusUnauthorized, fmt.Sprintf("user not found"))
			return
		}

		if tender.OrganizationID != orginazationId {
			respondJSONError(w, http.StatusForbidden, fmt.Sprintf("user does not have permissions to this tender"))
			return
		}
	case authorTypeOrganization:
		if tender.OrganizationID != bid.AuthorId {
			respondJSONError(w, http.StatusForbidden, fmt.Sprintf("tender does not belong to organization %s", bid.AuthorId))
			return
		}
	}

	err = h.repo.NewBid(bid)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save bid: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, bid)
}

func (h *Handler) MyBids(w http.ResponseWriter, r *http.Request) {
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

	userId, ok, err := h.repo.GetUserIDByUsername(username)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get user: %v", err))
		return
	}

	if !ok {
		respondJSONError(w, http.StatusUnauthorized, "user not found")
		return
	}

	organizationId, ok, err := h.repo.GetOrganizationIDByUsername(username)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get organization: %v", err))
		return
	}

	if !ok {
		respondJSONError(w, http.StatusUnauthorized, "user not found")
		return
	}

	bids, err := h.repo.MyBidsList(userId, organizationId, limit, offset)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed to select bid from database: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, JSON{Bids: &bids})
}

func (h *Handler) ListBidsByTenderId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	tenderID, err := uuid.Parse(vars["tenderId"])
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid tenderID format")
		return
	}

	var (
		limit    int
		offset   int
		username string
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

	if username == "" {
		respondJSONError(w, http.StatusBadRequest, "username is mandatory")
		return
	}

	organizationId, ok, err := h.repo.GetOrganizationIDByUsername(username)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get organization: %v", err))
		return
	}

	if !ok {
		respondJSONError(w, http.StatusUnauthorized, "user not found")
		return
	}

	tender, ok, err := h.repo.GetTenderByID(tenderID)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get tender: %v", err))
		return
	}

	if !ok {
		respondJSONError(w, http.StatusNotFound, "tender not found")
		return
	}

	if tender.OrganizationID != organizationId {
		respondJSONError(w, http.StatusForbidden, fmt.Sprintf("user %s does not have permissions to this tender", username))
		return
	}

	bids, err := h.repo.BidsByTenderId(tenderID.String(), limit, offset)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed to select bid from database: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, JSON{Bids: &bids})
}

func (h *Handler) EditBid(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	bidID, err := uuid.Parse(vars["bidId"])
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid bidID format")
		return
	}

	bid, err := h.repo.GetBidByID(bidID)
	if err != nil {
		respondJSONError(w, http.StatusNotFound, fmt.Sprintf("failed to get bid: %v", err))
		return
	}

	var updatedBid models.Bid
	err = json.NewDecoder(r.Body).Decode(&updatedBid)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid JSON format")
		return
	}

	err = h.repo.AddBidVersion(&models.BidVersion{
		BidID:       bid.ID,
		Name:        bid.Name,
		Description: bid.Description,
	})
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to add bid version: %v", err))
		return
	}

	if updatedBid.Name != "" {
		bid.Name = updatedBid.Name
	}
	if updatedBid.Description != "" {
		bid.Description = updatedBid.Description
	}

	err = h.repo.UpdateBid(bid)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to update bid: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, bid)
}

func (h *Handler) RollbackBid(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	bidID, err := uuid.Parse(vars["bidId"])
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid bidID format")
		return
	}

	bid, err := h.repo.GetBidByID(bidID)
	if err != nil {
		respondJSONError(w, http.StatusNotFound, fmt.Sprintf("failed to get bid: %v", err))
		return
	}

	version, err := uuid.Parse(vars["version"])
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid version format")
		return
	}

	bidVer, err := h.repo.GetBidVersionByID(version)
	if err != nil {
		respondJSONError(w, http.StatusNotFound, fmt.Sprintf("failed to get bid version: %v", err))
		return
	}

	if bidVer.BidID != bid.ID {
		respondJSONError(w, http.StatusForbidden, "bid version does not belong to bid")
		return
	}

	bid.Name = bidVer.Name
	bid.Description = bidVer.Description

	err = h.repo.UpdateBid(bid)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to update bid: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, bid)
}
