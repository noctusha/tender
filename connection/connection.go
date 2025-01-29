package connection

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/noctusha/tender/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository() (*Repository, error) {
	connStr := os.Getenv("POSTGRES_CONN")
	if connStr == "" {
		return nil, fmt.Errorf("POSTGRES_CONN not set")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping a database: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Close() {
	r.db.Close()
}

func (r *Repository) TendersList(serviceType string, limit int, offset int) ([]models.Tender, error) {
	tenders := []models.Tender{}
	var rows *sql.Rows
	var err error

	if limit == 0 {
		limit = 5
	}

	if serviceType == "" {
		rows, err = r.db.Query("SELECT id, name, description, service_type, status, organization_id, creator_username FROM tender WHERE status='PUBLISHED' ORDER BY name LIMIT $1 OFFSET $2", limit, offset)
	} else {
		rows, err = r.db.Query("SELECT id, name, description, service_type, status, organization_id, creator_username FROM tender WHERE service_type = $1 AND status='PUBLISHED' ORDER BY name LIMIT $2 OFFSET $3", serviceType, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to select data from tender: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		tender := models.Tender{}
		err := rows.Scan(&tender.ID, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status, &tender.OrganizationID, &tender.CreatorUserName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}
		tenders = append(tenders, tender)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return tenders, nil
}

func (r *Repository) NewTender(tender models.Tender) error {
	_, err := r.db.Exec(
		`INSERT INTO tender (id, name, description, service_type, status, organization_id, creator_username)
					VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		tender.ID, tender.Name, tender.Description, tender.ServiceType, tender.Status, tender.OrganizationID, tender.CreatorUserName)
	if err != nil {
		return fmt.Errorf("failed to insert data into tender: %w", err)
	}
	return nil
}

func (r *Repository) MyTendersList(username string, limit int, offset int) ([]models.Tender, error) {
	tenders := []models.Tender{}
	var rows *sql.Rows
	var err error

	if limit == 0 {
		limit = 5
	}

	if username == "" {
		return nil, fmt.Errorf("failed to find tenders by user: %w", err)
	} else {
		rows, err = r.db.Query("SELECT id, name, description, service_type, status, organization_id, creator_username FROM tender WHERE creator_username LIKE $1 LIMIT $2 OFFSET $3", username, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to select data from tender: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		tender := models.Tender{}
		err := rows.Scan(&tender.ID, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status, &tender.OrganizationID, &tender.CreatorUserName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}
		tenders = append(tenders, tender)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return tenders, nil
}

func (r *Repository) UpdateTender(tender *models.Tender) error {
	_, err := r.db.Exec(`UPDATE tender SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`,
		tender.Name, tender.Description, tender.ID)
	if err != nil {
		return fmt.Errorf("failed to update tender: %w", err)
	}
	return nil
}

func (r *Repository) UpdateTenderStatus(tenderId string, status string) error {
	_, err := r.db.Exec(`UPDATE tender SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		status, tenderId)
	if err != nil {
		return fmt.Errorf("failed to update tender status: %w", err)
	}
	return nil
}

func (r *Repository) AddTenderVersion(tenderVer *models.TenderVersion) error {
	_, err := r.db.Exec(`INSERT INTO tender_version (tender_id, name, description) VALUES ($1, $2, $3)`,
		tenderVer.TenderID, tenderVer.Name, tenderVer.Description)
	if err != nil {
		return fmt.Errorf("failed to insert data into tender_version: %w", err)
	}

	return nil
}

func (r *Repository) GetTenderByID(tenderID uuid.UUID) (*models.Tender, bool, error) {
	var tender models.Tender
	err := r.db.QueryRow(`SELECT id, name, description, service_type, status, organization_id, creator_username FROM tender WHERE id = $1`,
		tenderID.String()).Scan(&tender.ID, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status, &tender.OrganizationID, &tender.CreatorUserName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &tender, true, nil
}

func (r *Repository) GetTenderVersionByID(tenderVerID uuid.UUID) (*models.TenderVersion, error) {
	var tenderVer models.TenderVersion
	err := r.db.QueryRow(`SELECT id, tender_id, name, description FROM tender_version WHERE id = $1`,
		tenderVerID.String()).Scan(&tenderVer.ID, &tenderVer.TenderID, &tenderVer.Name, &tenderVer.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tender version no found")
		}
		return nil, fmt.Errorf("failed to select data from tender_version: %w", err)
	}
	return &tenderVer, nil
}

func (r *Repository) NewBid(bid models.Bid) error {
	_, err := r.db.Exec(
		`INSERT INTO bid (id, name, description, status, tender_id, author_type, author_id)
					VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		bid.ID, bid.Name, bid.Description, bid.Status, bid.TenderID,
		bid.AuthorType, bid.AuthorId)
	if err != nil {
		return fmt.Errorf("failed to insert data into bid: %w", err)
	}
	return nil
}

func (r *Repository) MyBidsList(userId string, organizationId string, limit int, offset int) ([]models.Bid, error) {
	if limit == 0 {
		limit = 5
	}

	bids := []models.Bid{}
	var rows *sql.Rows
	var err error

	if userId == "" || organizationId == "" {
		return nil, fmt.Errorf("userId and userId are mandatory")
	} else {
		rows, err = r.db.Query("SELECT id, name, description, status, tender_id, author_type, author_id FROM bid WHERE (author_type = 'User' AND author_id = $1) OR (author_type = 'Organization' AND author_id = $2) LIMIT $3 OFFSET $4",
			userId, organizationId, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to select data from bid: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		bid := models.Bid{}
		err := rows.Scan(&bid.ID, &bid.Name, &bid.Description, &bid.Status, &bid.TenderID, &bid.AuthorType, &bid.AuthorId)
		if err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}
		bids = append(bids, bid)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return bids, nil
}

func (r *Repository) BidsByTenderId(tenderID string, limit int, offset int) ([]models.Bid, error) {
	bids := []models.Bid{}
	var rows *sql.Rows
	var err error

	if limit == 0 {
		limit = 5
	}

	if tenderID == "" {
		return nil, fmt.Errorf("tenderID must not be empty")
	} else {
		rows, err = r.db.Query("SELECT id, name, description, status, tender_id, author_type, author_id FROM bid WHERE tender_id = $1 LIMIT $2 OFFSET $3", tenderID, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to select data from bid: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		bid := models.Bid{}
		err := rows.Scan(&bid.ID, &bid.Name, &bid.Description, &bid.Status, &bid.TenderID, &bid.AuthorType, &bid.AuthorId)
		if err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}
		bids = append(bids, bid)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return bids, nil
}

func (r *Repository) AddBidVersion(bidVer *models.BidVersion) error {
	_, err := r.db.Exec(`INSERT INTO bid_version (bid_id, name, description) VALUES ($1, $2, $3)`,
		bidVer.BidID, bidVer.Name, bidVer.Description)
	if err != nil {
		return fmt.Errorf("failed to insert data into bid_version: %w", err)
	}

	return nil
}

func (r *Repository) GetBidByID(bidID uuid.UUID) (*models.Bid, error) {
	var bid models.Bid
	err := r.db.QueryRow(`SELECT id, name, description, status, tender_id, author_type, author_id FROM bid WHERE id = $1`,
		bidID.String()).Scan(&bid.ID, &bid.Name, &bid.Description, &bid.Status, &bid.TenderID, &bid.AuthorType, &bid.AuthorId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bid not found")
		}
		return nil, fmt.Errorf("failed to select data from bid: %w", err)
	}
	return &bid, nil
}

func (r *Repository) GetBidVersionByID(bidVerID uuid.UUID) (*models.BidVersion, error) {
	var bidVer models.BidVersion
	err := r.db.QueryRow(`SELECT id, bid_id, name, description FROM bid_version WHERE id = $1`,
		bidVerID.String()).Scan(&bidVer.ID, &bidVer.BidID, &bidVer.Name, &bidVer.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bid version no found")
		}
		return nil, fmt.Errorf("failed to select data from bid_version: %w", err)
	}
	return &bidVer, nil
}

func (r *Repository) UpdateBid(bid *models.Bid) error {
	_, err := r.db.Exec(`UPDATE bid SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`,
		bid.Name, bid.Description, bid.ID)
	if err != nil {
		return fmt.Errorf("failed to update bid: %w", err)
	}
	return nil
}

func (r *Repository) GetUserIDByUsername(username string) (string, bool, error) {
	var userId string

	err := r.db.QueryRow(`SELECT employee.id FROM employee
    WHERE employee.username=$1`,
		username).Scan(&userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}

		return "", false, fmt.Errorf("failed to find user by username: %w", err)
	}

	return userId, true, nil
}

func (r *Repository) GetOrganizationIDByUsername(username string) (string, bool, error) {
	var organizationId string

	err := r.db.QueryRow(`SELECT organization.id FROM employee
    JOIN organization_responsible ON employee.id = organization_responsible.user_id
    JOIN organization ON organization_responsible.organization_id = organization.id
    WHERE employee.username=$1`,
		username).Scan(&organizationId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}

		return "", false, fmt.Errorf("failed to find organization by username: %w", err)
	}

	return organizationId, true, nil
}

func (r *Repository) GetOrganizationIDByUserID(userId string) (string, bool, error) {
	var organizationId string

	err := r.db.QueryRow(`SELECT organization.id FROM organization_responsible
    JOIN organization ON organization_responsible.organization_id = organization.id
    WHERE organization_responsible.user_id=$1`,
		userId).Scan(&organizationId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}

		return "", false, fmt.Errorf("failed to find organization by username: %w", err)
	}

	return organizationId, true, nil
}
