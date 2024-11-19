package services

import (
    "time"

    gron "github.com/roylee0704/gron"
    "github.com/SangBejoo/service-parking/models"
    "github.com/SangBejoo/service-parking/api"
)

type Scheduler struct {
    graph      *gron.Cron
}

func (s *Scheduler) updateLocations(vehicles []models.Vehicle) {
    // Add logic to update locations of vehicles
}

func (s *Scheduler) Start() {
    s.graph.Add(gron.Every(5*time.Minute), gron.JobFunc(func() {
        vehicles := api.GetDummyVehicles()
        var vehicleList []models.Vehicle
        for _, v := range vehicles {
            vehicleList = append(vehicleList, *v)
        }
        s.updateLocations(vehicleList)
    }))
    s.graph.Start()
}