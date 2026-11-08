package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"report-service/internal/application"
	"strconv"
	"time"
)

type InputParams struct {
	Since time.Time
	Limit int
}

func (in *InputParams) Validate() error {
	if in.Limit <= 0 || in.Limit > 100 {
		return errors.New("limit deve estar entre 1 e 100")
	}
	return nil
}

func ParseInputParams(r *http.Request) (InputParams, error) {
	q := r.URL.Query()

	limit := 20
	if rawLimit := q.Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			return InputParams{}, fmt.Errorf("parâmetro 'limit' inválido: %w", err)
		}
		limit = parsed
	}

	var since time.Time
	if rawSince := q.Get("since"); rawSince != "" {
		parsed, err := time.Parse(time.RFC3339, rawSince)
		if err != nil {
			return InputParams{}, fmt.Errorf("parâmetro 'since' deve seguir o formato RFC3339 (ex: 2026-11-07T15:04:05Z): %w", err)
		}
		since = parsed
	}

	input := InputParams{
		Since: since,
		Limit: limit,
	}

	if err := input.Validate(); err != nil {
		return InputParams{}, err
	}

	return input, nil
}

type HTTPReportHandler struct {
	ReportUseCase *application.ReportUseCase
}

func NewHttpReportHandler(reportUseCase *application.ReportUseCase) *HTTPReportHandler {
	return &HTTPReportHandler{
		ReportUseCase: reportUseCase,
	}
}

func (hrh HTTPReportHandler) HandlerReportSync(w http.ResponseWriter, r *http.Request) {

	input, err := ParseInputParams(r)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(err.Error()); err != nil {
		}
		return
	}
	ctx := r.Context()
	reports, _ := hrh.ReportUseCase.ReportRepository.GetReportSync(ctx, input.Since, input.Limit)
	log.Print(reports)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]any{
		"status": "success",
		"data":   reports,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
	}
}
