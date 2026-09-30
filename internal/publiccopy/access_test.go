package publiccopy

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNewRequiresReaderOnlyWhenEnabled(t *testing.T) {
	if _, err := New(nil, nil, false, ""); err == nil {
		t.Fatal("nil database accepted")
	}
	pool := &pgxpool.Pool{}
	if _, err := New(pool, nil, false, ""); err != nil {
		t.Fatalf("disabled access required object credentials: %v", err)
	}
	if _, err := New(pool, nil, true, "https://alerts.example"); err == nil {
		t.Fatal("enabled access accepted no object reader")
	}
}
