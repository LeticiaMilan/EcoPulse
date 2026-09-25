package LatLngWorkGeocoding

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"report-service/internal/ports"
)

type LatLngWorkGeocodingInputParams struct {
	lat    string `json:"lat"`
	lng    string `json:"lon"`
	Apikey string `json:"key"`
}

type LatLngWorkGeocodingAPI struct {
	BaseUrl    string
	Params     LatLngWorkGeocodingInputParams
	httpClient *http.Client
}

func NewLatLngWorkGeocodingApi(httpClient *http.Client) *LatLngWorkGeocodingAPI {
	return &LatLngWorkGeocodingAPI{
		BaseUrl: os.Getenv("LATLNG_WORK_REVERSE_GEOCODING_API_BASE_URL"),
		Params: LatLngWorkGeocodingInputParams{
			lat:    "",
			lng:    "",
			Apikey: os.Getenv("LATLNG_WORK_API_KEY"),
		},
		httpClient: httpClient,
	}
}

func (g *LatLngWorkGeocodingAPI) IsBrazil(ctx context.Context, GeocodingInputParams *ports.GeocodingInputParams) (bool, error) {
	g.Params.lat = FormatLatLng(GeocodingInputParams.Latitude)
	g.Params.lng = FormatLatLng(GeocodingInputParams.Longitude)
	finalUrl := fmt.Sprintf("%s/reverse?lat=%s")

	req, _ := http.NewRequestWithContext(ctx, "GET", finalUrl, nil)
	req.Header.Add("X-Api-Key", g.Params.Apikey)
	response, err := g.httpClient.Do(req)
	return false, err
}
func FormatLatLng(latitude float64) string {
	return fmt.Sprintf("%f")

}
