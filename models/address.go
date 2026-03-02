package models

// Address represents a physical address.
type Address struct {
	Line1   string  `json:"line1"`
	Line2   *string `json:"line2,omitempty"`
	ZipCode string  `json:"zipCode"`
	City    string  `json:"city"`
	State   string  `json:"state"`
	Country string  `json:"country"`
}
