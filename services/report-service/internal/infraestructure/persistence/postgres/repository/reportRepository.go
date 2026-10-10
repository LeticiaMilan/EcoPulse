package repository

import (
	"context"
	"log"
	"report-service/internal/domain"
	"report-service/internal/infraestructure/persistence/postgres/db"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportPostgresModel struct {
	ID        uuid.UUID `db:"id"`
	Latitude  float64   `db:"latitude"`
	Longitude float64   `db:"longitude"`
	Type      string    `db:"type"`
	UserID    uuid.UUID `db:"user_id"`
	Datetime  time.Time `db:"datetime"`
	Status    string    `db:"status"`
	//Area      AffectedArea
}

type ReportRepositoryImpl struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func newCreateReportParams(report *domain.Report) db.CreateReportParams {
	var lat, lng pgtype.Float8
	var Type pgtype.Text
	var datetime pgtype.Timestamptz
	var status pgtype.Text
	if err := lat.Scan(report.Latitude); err != nil {
		log.Print("Error parsing latitude")
	}
	if err := lng.Scan(report.Longitude); err != nil {
		log.Print("Error parsing longitude")
	}
	if err := Type.Scan(report.Type); err != nil {
		log.Print("Error parsing type")
	}
	if err := datetime.Scan(report.Datetime); err != nil {
		log.Print("Error parsing datetime")
	}
	if err := status.Scan(report.Status); err != nil {
		log.Print("Error parsing status")
	}
	return db.CreateReportParams{
		ID:        report.ID,
		Latitude:  lat,
		Longitude: lng,
		Type:      Type,
		Status:    status,
		UserID:    report.UserID,
		Datetime:  datetime,
	}
}

func NewReportRepository(pool *pgxpool.Pool) *ReportRepositoryImpl {
	return &ReportRepositoryImpl{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (r *ReportRepositoryImpl) Create(ctx context.Context, report *domain.Report) error {
	ReportParams := newCreateReportParams(report)
	createReport, err := r.queries.CreateReport(ctx, ReportParams)
	if err != nil {
		return err
	}
	log.Printf("Create Report %+v", createReport)
	return nil
}

func (r *ReportRepositoryImpl) Update(ctx context.Context, report *domain.Report) error {
	return nil
}
