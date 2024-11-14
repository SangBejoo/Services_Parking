package services

import (
    "time"

    gron "github.com/roylee0704/gron"
    api "github.com/SangBejoo/service-parking/Api"
)

type Scheduler struct {
    spatialMap *SpatialMap
    graph      *gron.Gron
}

func (s *Scheduler) updateLocations(vehicles []api.Vehicle) {
    // Add logic to update locations of vehicles
}

func (s *Scheduler) Start() {
    s.graph.Add(gron.Every(5*time.Minute), func() {
        vehicles := api.GetDummyVehicles()
        s.updateLocations(vehicles)
    })
    s.graph.Start()
}