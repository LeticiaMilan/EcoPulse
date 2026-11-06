package domain

import (
	"encoding/json"

	"github.com/google/uuid"
)

type AreaCenter struct {
	Latitude  float64
	Longitude float64
}
type Area struct {
	Id      uuid.UUID
	Type    string
	Status  string
	Area_M2 float64
	Center  AreaCenter
	Polygon json.RawMessage
	Reports []uuid.UUID
	Users   int
}
