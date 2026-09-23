package ports

import "context"

type GeocodingInputParams struct {
	Latitude  float64
	Longitude float64
}

type GeocodingAPI interface {
	ReverseGeocoding(ctx context.Context, GeocodingInputParams *GeocodingInputParams) (string, error)
}
