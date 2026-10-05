package application

// DTOS

type ProcessReportUseCase struct{}

func (uc *ProcessReportUseCase) Execute(ReportGeneratedEventDTO *ReportGeneratedEventDTO) error {

	return nil
}

type CancelReportUseCase struct{}

func (uc *CancelReportUseCase) Execute() error {
	return nil
}
