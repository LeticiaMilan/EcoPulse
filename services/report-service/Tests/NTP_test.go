package Tests

import (
	"report-service/internal/infraestructure/external/BrasilNtp"
	"testing"
)

func TestNow(t *testing.T) {
	service, err := BrasilNtp.NewBrasilNtpClock()
	if err != nil {
		t.Fatal(err)
	}
	ntpTime := service.Now()

	t.Logf("Horário recebido do NTP: %v", ntpTime)

	if ntpTime.IsZero() {
		t.Fatal("Recebeu valor zero-value")
	}
}
