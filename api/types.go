package api

import "time"

// Owner represents a HubSpot owner (user)
type Owner struct {
	ID        string      `json:"id"`
	Email     string      `json:"email"`
	FirstName string      `json:"firstName"`
	LastName  string      `json:"lastName"`
	UserID    int64       `json:"userId"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
	Archived  bool        `json:"archived"`
	Teams     []OwnerTeam `json:"teams,omitempty"`
}

// OwnerTeam represents a team assignment for an owner
type OwnerTeam struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Primary bool   `json:"primary"`
}

// FullName returns the owner's full name
func (o *Owner) FullName() string {
	if o.FirstName == "" && o.LastName == "" {
		return o.Email
	}
	if o.FirstName == "" {
		return o.LastName
	}
	if o.LastName == "" {
		return o.FirstName
	}
	return o.FirstName + " " + o.LastName
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Results []interface{} `json:"results"`
	Paging  *Paging       `json:"paging,omitempty"`
}

// Paging contains pagination information
type Paging struct {
	Next *PagingNext `json:"next,omitempty"`
}

// PagingNext contains the next page cursor
type PagingNext struct {
	After string `json:"after"`
	Link  string `json:"link,omitempty"`
}
