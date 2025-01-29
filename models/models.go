package models

type Tender struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	ServiceType     string `json:"serviceType"`
	Status          string `json:"status"`
	OrganizationID  string `json:"organizationId"`
	CreatorUserName string `json:"creatorUsername"`
}

type TenderVersion struct {
	ID          string `json:"id"`
	TenderID    string `json:"tender_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Bid struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Status          string `json:"status"`
	TenderID        string `json:"tenderId"`
	CreatorUserName string `json:"creatorUsername,omitempty"`
	AuthorType      string `json:"authorType"`
	AuthorId        string `json:"authorId"`
}

type BidVersion struct {
	ID          string `json:"id"`
	BidID       string `json:"bid_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Employee struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type Organization struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type OrganizationResponsible struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	UserID         string `json:"userId"`
}
