package models

import "time"

type Vehicle struct {
    ID        string    `json:"id"`
    Latitude  float64   `json:"latitude"`
    Longitude float64   `json:"longitude"`
    Timestamp time.Time `json:"timestamp"`
}