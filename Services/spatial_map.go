package services

import (
    "fmt"
    "math"
    "sync"

    "github.com/SangBejoo/service-parking/models"
)

type SpatialMap struct {
    cells    map[string]map[string]*models.Vehicle
    gridSize float64
    mutex    *sync.RWMutex
}

func NewSpatialMap() *SpatialMap {
    return &SpatialMap{
        cells:    make(map[string]map[string]*models.Vehicle),
        gridSize: 0.01,
        mutex:    &sync.RWMutex{},
    }
}

func (sm *SpatialMap) hashKey(lat, lon float64) string {
    return fmt.Sprintf("%d:%d",
        int(math.Floor(lat/sm.gridSize)),
        int(math.Floor(lon/sm.gridSize)))
}
