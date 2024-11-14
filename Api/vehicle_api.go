package api

import (
    "time"

    "github.com/SangBejoo/service-parking/models"
)

func GetDummyVehicles() []*models.Vehicle {
    return []*models.Vehicle{
        {
            ID:        "v1",
            Latitude:  1.2345,
            Longitude: 103.8765,
            Timestamp: time.Now(),
        },
        {
            ID:        "v2",
            Latitude:  1.2355,
            Longitude: 103.8775,
            Timestamp: time.Now(),
        },
    }
}