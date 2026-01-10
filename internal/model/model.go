package model

import "time"

type Incident struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	RadiusM     int       `json:"radius_m"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LocationRequest struct {
	UserID    int64   `json:"user_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type LocationResponse struct {
	LocationRequest
	LocationsIDS []int64 `json:"locations_ids"`
}
