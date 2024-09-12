package models

import "time"

type Tender struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	ServiceType string    `json:"serviceType"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}

type NewTenderRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	ServiceType     string `json:"serviceType"`
	OrganizationID  string `json:"organizationId"`
	CreatorUsername string `json:"creatorUsername"`
}

type EditTenderRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ServiceType string `json:"serviceType,omitempty"`
}

type Bid struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	AuthorType string    `json:"authorType"`
	AuthorID   string    `json:"authorId"`
	Version    string    `json:"version"`
	CreatedAt  time.Time `json:"createdAt"`
}

type BidRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	TenderID    string `json:"tenderId"`
	AuthorType  string `json:"authorType"`
	AuthorID    string `json:"authorId"`
}

type EditBidRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type Feedback struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}
