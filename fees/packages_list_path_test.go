package fees

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuildPackagesListPathNilOptions(t *testing.T) {
	t.Parallel()

	path := buildPackagesListPath(nil)
	assert.Equal(t, "/packages", path)
}

func TestBuildPackagesListPathWithAllOptions(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)

	path := buildPackagesListPath(&PackageListOptions{
		SegmentID:        "seg-001",
		LedgerID:         "ledger-001",
		TransactionRoute: "ted_out",
		Enabled:          boolPtr(true),
		PageSize:         25,
		PageNumber:       2,
		SortOrder:        "desc",
		CreatedFrom:      &start,
		CreatedTo:        &end,
	})

	assert.Contains(t, path, "/packages?")
	assert.Contains(t, path, "segmentId=seg-001")
	assert.Contains(t, path, "ledgerId=ledger-001")
	assert.Contains(t, path, "transactionRoute=ted_out")
	assert.Contains(t, path, "enable=true")
	assert.Contains(t, path, "limit=25")
	assert.Contains(t, path, "page=2")
	assert.Contains(t, path, "sort_order=desc")
	assert.Contains(t, path, "start_date=2026-01-01")
	assert.Contains(t, path, "end_date=2026-12-31")
}

func TestBuildPackagesListPathEmptyOptions(t *testing.T) {
	t.Parallel()

	path := buildPackagesListPath(&PackageListOptions{})
	assert.Equal(t, "/packages", path)
}
