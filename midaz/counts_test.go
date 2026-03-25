package midaz

import (
	"context"
	"errors"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMidazCountEndpoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func(t *testing.T, mock *mockBackend)
		path string
	}{
		{
			name: "Organizations",
			path: "/organizations/metrics/count",
			call: func(t *testing.T, mock *mockBackend) {
				t.Helper()

				count, err := newOrganizationsService(mock).Count(context.Background())
				require.NoError(t, err)
				assert.Equal(t, 42, count)
			},
		},
		{
			name: "Ledgers",
			path: "/organizations/org-1/ledgers/metrics/count",
			call: func(t *testing.T, mock *mockBackend) {
				t.Helper()

				count, err := newLedgersService(mock).Count(context.Background(), "org-1")
				require.NoError(t, err)
				assert.Equal(t, 42, count)
			},
		},
		{
			name: "Accounts",
			path: "/organizations/org-1/ledgers/led-1/accounts/metrics/count",
			call: func(t *testing.T, mock *mockBackend) {
				t.Helper()

				count, err := newAccountsService(mock).Count(context.Background(), "org-1", "led-1")
				require.NoError(t, err)
				assert.Equal(t, 42, count)
			},
		},
		{
			name: "Assets",
			path: "/organizations/org-1/ledgers/led-1/assets/metrics/count",
			call: func(t *testing.T, mock *mockBackend) {
				t.Helper()

				count, err := newAssetsService(mock).Count(context.Background(), "org-1", "led-1")
				require.NoError(t, err)
				assert.Equal(t, 42, count)
			},
		},
		{
			name: "Portfolios",
			path: "/organizations/org-1/ledgers/led-1/portfolios/metrics/count",
			call: func(t *testing.T, mock *mockBackend) {
				t.Helper()

				count, err := newPortfoliosService(mock).Count(context.Background(), "org-1", "led-1")
				require.NoError(t, err)
				assert.Equal(t, 42, count)
			},
		},
		{
			name: "Segments",
			path: "/organizations/org-1/ledgers/led-1/segments/metrics/count",
			call: func(t *testing.T, mock *mockBackend) {
				t.Helper()

				count, err := newSegmentsService(mock).Count(context.Background(), "org-1", "led-1")
				require.NoError(t, err)
				assert.Equal(t, 42, count)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockBackend{callHeadFn: func(_ context.Context, path string) (map[string][]string, error) {
				assert.Equal(t, tc.path, path)
				return map[string][]string{"X-Total-Count": {"42"}}, nil
			}}

			tc.call(t, mock)
		})
	}
}

func TestAccountsCountValidation(t *testing.T) {
	t.Parallel()

	count, err := newAccountsService(&mockBackend{}).Count(context.Background(), "", "led-1")
	require.Error(t, err)
	assert.Equal(t, 0, count)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}
