package fees

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// mockBackend — configurable core.Backend stub for service tests
// ---------------------------------------------------------------------------

// mockBackend delegates Call to a user-supplied callFn so tests can
// verify HTTP method, path, and body while injecting canned responses.
type mockBackend struct {
	callFn func(ctx context.Context, method, path string, body, result any) error
}

func (m *mockBackend) Do(ctx context.Context, req core.Request) (*core.Response, error) {
	if m.callFn == nil {
		return &core.Response{}, nil
	}

	var result any

	resultArg := any(&result)
	if req.ExpectNoResponse {
		resultArg = nil
	}

	if err := m.callFn(ctx, req.Method, req.Path, reqBody(req), resultArg); err != nil {
		return nil, err
	}

	if req.ExpectNoResponse {
		return &core.Response{}, nil
	}

	return jsonResponse(result)
}

// Compile-time check.
var _ core.Backend = (*mockBackend)(nil)

// jsonInto marshals src to JSON and unmarshals into dst. This is how the
// mock backend simulates the real backend populating result pointers.
func jsonInto(src, dst any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("mock marshal: %w", err)
	}

	return json.Unmarshal(b, dst)
}

func jsonResponse(result any) (*core.Response, error) {
	if result == nil {
		return &core.Response{}, nil
	}

	b, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &core.Response{Body: b}, nil
}

func reqBody(req core.Request) any {
	if len(req.BodyBytes) > 0 {
		return req.BodyBytes
	}

	return req.Body
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

// ---------------------------------------------------------------------------
// Shared fixtures
// ---------------------------------------------------------------------------

var testPackage = Package{
	ID:            "pkg-001",
	FeeGroupLabel: "standard_fees",
	Description:   strPtr("Default fee package for TED transfers"),
	SegmentID:     strPtr("seg-retail"),
	LedgerID:      "ledger-001",
	MinimumAmount: "0.00",
	MaximumAmount: "1000000.00",
	Fees: map[string]Fee{
		"ted_fee": {
			FeeLabel: "TED Transfer Fee",
			CalculationModel: &CalculationModel{
				ApplicationRule: "flatFee",
				Calculations: []Calculation{
					{Type: "flat", Value: "5.00"},
				},
			},
			ReferenceAmount:  "originalAmount",
			Priority:         1,
			IsDeductibleFrom: boolPtr(false),
			CreditAccount:    "platform-fee-account",
		},
	},
	Enable:    boolPtr(true),
	CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
}

// ---------------------------------------------------------------------------
// packagesServiceAPI.Create
// ---------------------------------------------------------------------------

func TestPackagesCreate(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/packages", path)
			assert.NotNil(t, body)

			return jsonInto(testPackage, result)
		},
	}

	svc := newPackagesService(backend)
	input := &CreatePackageInput{
		FeeGroupLabel: "standard_fees",
		LedgerID:      "ledger-001",
		MinimumAmount: "0.00",
		MaximumAmount: "1000000.00",
		Fees: map[string]Fee{
			"ted_fee": {
				FeeLabel: "TED Transfer Fee",
				CalculationModel: &CalculationModel{
					ApplicationRule: "flatFee",
					Calculations: []Calculation{
						{Type: "flat", Value: "5.00"},
					},
				},
				ReferenceAmount:  "originalAmount",
				IsDeductibleFrom: boolPtr(false),
				CreditAccount:    "platform-fee-account",
			},
		},
		Enable: boolPtr(true),
	}

	pkg, err := svc.Create(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, pkg)
	assert.Equal(t, "pkg-001", pkg.ID)
	assert.Equal(t, "standard_fees", pkg.FeeGroupLabel)
	assert.Equal(t, "ledger-001", pkg.LedgerID)
	assert.Contains(t, pkg.Fees, "ted_fee")
	assert.True(t, *pkg.Enable)
}

func TestPackagesCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})

	pkg, err := svc.Create(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPackagesCreateBackendError(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("network failure")
		},
	}

	svc := newPackagesService(backend)
	input := &CreatePackageInput{
		FeeGroupLabel: "test",
		LedgerID:      "ledger-001",
		MinimumAmount: "0",
		MaximumAmount: "100",
		Fees:          map[string]Fee{},
		Enable:        boolPtr(true),
	}

	pkg, err := svc.Create(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.Contains(t, err.Error(), "network failure")
}

func TestPackagesCreateRequiresCoreFields(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})

	tests := []struct {
		name    string
		input   *CreatePackageInput
		message string
	}{
		{name: "missing fee group label", input: &CreatePackageInput{LedgerID: "ledger-001", MinimumAmount: "0", MaximumAmount: "100", Enable: boolPtr(true)}, message: "fee group label is required"},
		{name: "missing ledger id", input: &CreatePackageInput{FeeGroupLabel: "fees", MinimumAmount: "0", MaximumAmount: "100", Enable: boolPtr(true)}, message: "ledger ID is required"},
		{name: "missing minimum amount", input: &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MaximumAmount: "100", Enable: boolPtr(true)}, message: "minimum amount is required"},
		{name: "missing maximum amount", input: &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "0", Enable: boolPtr(true)}, message: "maximum amount is required"},
		{name: "missing enable flag", input: &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "0", MaximumAmount: "100"}, message: "enable flag is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pkg, err := svc.Create(context.Background(), tt.input)
			require.Error(t, err)
			assert.Nil(t, pkg)
			assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
			assert.Contains(t, err.Error(), tt.message)
		})
	}
}

func TestPackagesCreateNormalizesNilFeesMap(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/packages", path)

			input, ok := body.(*CreatePackageInput)
			require.True(t, ok)
			require.NotNil(t, input.Fees)
			assert.Empty(t, input.Fees)

			return jsonInto(Package{
				ID:            "pkg-001",
				FeeGroupLabel: input.FeeGroupLabel,
				LedgerID:      input.LedgerID,
				MinimumAmount: input.MinimumAmount,
				MaximumAmount: input.MaximumAmount,
				Fees:          input.Fees,
				Enable:        input.Enable,
			}, result)
		},
	}

	svc := newPackagesService(backend)
	pkg, err := svc.Create(context.Background(), &CreatePackageInput{
		FeeGroupLabel: "fees",
		LedgerID:      "ledger-001",
		MinimumAmount: "0.00",
		MaximumAmount: "100.00",
		Enable:        boolPtr(true),
	})
	require.NoError(t, err)
	require.NotNil(t, pkg)
	require.NotNil(t, pkg.Fees)
	assert.Empty(t, pkg.Fees)
}

func TestPackagesCreateValidatesNestedFeeContract(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})

	tests := []struct {
		name    string
		input   *CreatePackageInput
		message string
	}{
		{
			name:    "invalid minimum amount ordering",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "10.00", MaximumAmount: "5.00", Enable: boolPtr(true)},
			message: "minimum amount must be less than or equal to maximum amount",
		},
		{
			name:    "negative minimum amount",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "-1.00", MaximumAmount: "5.00", Enable: boolPtr(true)},
			message: "minimum amount must be greater than or equal to zero",
		},
		{
			name:    "negative maximum amount",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "0.00", MaximumAmount: "-5.00", Enable: boolPtr(true)},
			message: "maximum amount must be greater than or equal to zero",
		},
		{
			name:    "invalid reference amount",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "0.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "flatFee", Calculations: []Calculation{{Type: "flat", Value: "1.00"}}}, ReferenceAmount: "send.value", IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "reference amount must be originalAmount or afterFeesAmount",
		},
		{
			name:    "negative calculation value",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "0.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "flatFee", Calculations: []Calculation{{Type: "flat", Value: "-1.00"}}}, ReferenceAmount: "originalAmount", IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "calculation value must be greater than zero",
		},
		{
			name:    "zero calculation value",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "0.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "flatFee", Calculations: []Calculation{{Type: "flat", Value: "0"}}}, ReferenceAmount: "originalAmount", IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "calculation value must be greater than zero",
		},
		{
			name:    "deductible after fees amount",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "10.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "flatFee", Calculations: []Calculation{{Type: "flat", Value: "1.00"}}}, ReferenceAmount: "afterFeesAmount", IsDeductibleFrom: boolPtr(true), CreditAccount: "acct"}}},
			message: "deductible fees must use originalAmount",
		},
		{
			name:    "invalid calculation model",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "10.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "percentual", Calculations: []Calculation{{Type: "flat", Value: "1.00"}}}, ReferenceAmount: "originalAmount", IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "percentual requires a percentage calculation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pkg, err := svc.Create(context.Background(), tt.input)
			require.Error(t, err)
			assert.Nil(t, pkg)
			assert.Contains(t, err.Error(), tt.message)
		})
	}
}

// ---------------------------------------------------------------------------
// packagesServiceAPI.Get
// ---------------------------------------------------------------------------

func TestPackagesGet(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/packages/pkg-001", path)
			assert.Nil(t, body)

			return jsonInto(testPackage, result)
		},
	}

	svc := newPackagesService(backend)

	pkg, err := svc.Get(context.Background(), "pkg-001")
	require.NoError(t, err)
	require.NotNil(t, pkg)
	assert.Equal(t, "pkg-001", pkg.ID)
	assert.Equal(t, "ledger-001", pkg.LedgerID)
	assert.True(t, *pkg.Enable)
}

func TestPackagesCreateValidationAdditionalMatrix(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})

	tests := []struct {
		name    string
		input   *CreatePackageInput
		message string
	}{
		{
			name:    "whitespace fee group label",
			input:   &CreatePackageInput{FeeGroupLabel: "   ", LedgerID: "ledger-001", MinimumAmount: "0.00", MaximumAmount: "1.00", Enable: boolPtr(true)},
			message: "fee group label is required",
		},
		{
			name:    "comma decimal formatting",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "10,00", MaximumAmount: "100.00", Enable: boolPtr(true)},
			message: "minimum amount must use dot decimal formatting",
		},
		{
			name:    "missing calculation model",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "1.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", ReferenceAmount: "originalAmount", IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "calculation model is required",
		},
		{
			name:    "unsupported application rule",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "1.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "mystery", Calculations: []Calculation{{Type: "flat", Value: "1.00"}}}, ReferenceAmount: "originalAmount", IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "unsupported application rule",
		},
		{
			name:    "empty calculations",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "1.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "flatFee"}, ReferenceAmount: "originalAmount", IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "at least one calculation is required",
		},
		{
			name:    "max between needs two calculations",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "1.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "maxBetweenTypes", Calculations: []Calculation{{Type: "flat", Value: "1.00"}}}, ReferenceAmount: "originalAmount", IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "maxBetweenTypes requires at least two calculations",
		},
		{
			name:    "priority one must use original amount",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "1.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "flatFee", Calculations: []Calculation{{Type: "flat", Value: "1.00"}}}, ReferenceAmount: "afterFeesAmount", Priority: 1, IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "priority 1 fees must use originalAmount",
		},
		{
			name:    "deductible percentage above 100",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "1.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "percentual", Calculations: []Calculation{{Type: "percentage", Value: "101.00"}}}, ReferenceAmount: "originalAmount", IsDeductibleFrom: boolPtr(true), CreditAccount: "acct"}}},
			message: "deductible percentage cannot be greater than 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pkg, err := svc.Create(context.Background(), tt.input)
			require.Error(t, err)
			assert.Nil(t, pkg)
			assert.Contains(t, err.Error(), tt.message)
		})
	}
}

func TestPackagesGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})

	pkg, err := svc.Get(context.Background(), "")
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPackagesGetBackendError(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("not found")
		},
	}

	svc := newPackagesService(backend)

	pkg, err := svc.Get(context.Background(), "pkg-999")
	require.Error(t, err)
	assert.Nil(t, pkg)
}

// ---------------------------------------------------------------------------
// packagesServiceAPI.List
// ---------------------------------------------------------------------------

func TestPackagesList(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/packages")
			assert.Nil(t, body)

			resp := struct {
				Items []Package `json:"items"`
				Page  int       `json:"page"`
				Limit int       `json:"limit"`
				Total int       `json:"total"`
			}{
				Items: []Package{testPackage},
				Page:  1,
				Limit: 10,
				Total: 1,
			}

			return jsonInto(resp, result)
		},
	}

	svc := newPackagesService(backend)

	resp, err := svc.List(context.Background(), &PackageListOptions{PageSize: 10})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "pkg-001", resp.Items[0].ID)
	assert.Equal(t, 1, resp.TotalItems)
	assert.Equal(t, 10, resp.PageSize)
}

func TestPackagesListNilOptions(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/packages", path) // no query params

			resp := struct {
				Items []Package `json:"items"`
				Page  int       `json:"page"`
				Limit int       `json:"limit"`
				Total int       `json:"total"`
			}{
				Items: []Package{},
				Page:  1,
				Limit: 10,
				Total: 0,
			}

			return jsonInto(resp, result)
		},
	}

	svc := newPackagesService(backend)

	resp, err := svc.List(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Items)
}

func TestPackagesListWithFilters(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "ledgerId=ledger-001")
			assert.Contains(t, path, "segmentId=seg-retail")
			assert.Contains(t, path, "enable=true")

			resp := struct {
				Items []Package `json:"items"`
				Page  int       `json:"page"`
				Limit int       `json:"limit"`
				Total int       `json:"total"`
			}{
				Items: []Package{testPackage},
				Page:  1,
				Limit: 25,
				Total: 1,
			}

			return jsonInto(resp, result)
		},
	}

	svc := newPackagesService(backend)

	resp, err := svc.List(context.Background(), &PackageListOptions{
		LedgerID:  "ledger-001",
		SegmentID: "seg-retail",
		Enabled:   boolPtr(true),
		PageSize:  25,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Items, 1)
}

func TestPackagesListNilBackendUsesCoreError(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(nil)

	resp, err := svc.List(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, core.ErrNilBackend)
}

func TestPackagesListRejectsIncompleteDateRange(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := newPackagesService(&mockBackend{})

	resp, err := svc.List(context.Background(), &PackageListOptions{CreatedFrom: &start})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	assert.Contains(t, err.Error(), "createdFrom and createdTo must both be provided")
}

func TestPackagesListRejectsInvalidRangeAndSortOrder(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := newPackagesService(&mockBackend{})

	resp, err := svc.List(context.Background(), &PackageListOptions{CreatedFrom: &start, CreatedTo: &end})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "createdFrom must be before or equal to createdTo")

	resp, err = svc.List(context.Background(), &PackageListOptions{SortOrder: "sideways"})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "sortOrder must be either asc or desc")
}

func TestPackagesListBackendError(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("service unavailable")
		},
	}

	svc := newPackagesService(backend)

	resp, err := svc.List(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "service unavailable")
}

// ---------------------------------------------------------------------------
// packagesServiceAPI.Update
// ---------------------------------------------------------------------------

func TestPackagesUpdate(t *testing.T) {
	t.Parallel()

	updated := testPackage
	updated.FeeGroupLabel = "premium_fees"

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/packages/pkg-001", path)
			assert.NotNil(t, body)

			return jsonInto(updated, result)
		},
	}

	svc := newPackagesService(backend)
	input := &UpdatePackageInput{FeeGroupLabel: "premium_fees"}

	pkg, err := svc.Update(context.Background(), "pkg-001", input)
	require.NoError(t, err)
	require.NotNil(t, pkg)
	assert.Equal(t, "premium_fees", pkg.FeeGroupLabel)
}

func TestPackagesUpdateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})
	input := &UpdatePackageInput{FeeGroupLabel: "test"}

	pkg, err := svc.Update(context.Background(), "", input)
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPackagesUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})

	pkg, err := svc.Update(context.Background(), "pkg-001", nil)
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPackagesUpdateValidatesNestedFeeContract(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})
	minAmount := "10.00"
	negativeMin := "-1.00"

	pkg, err := svc.Update(context.Background(), "pkg-001", &UpdatePackageInput{
		MinimumAmount: &minAmount,
		Fees: map[string]Fee{
			"fee": {
				CalculationModel: &CalculationModel{ApplicationRule: "flatFee", Calculations: []Calculation{{Type: "flat", Value: "20.00"}}},
				IsDeductibleFrom: boolPtr(true),
			},
		},
	})
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.Contains(t, err.Error(), "deductible flat fee cannot exceed minimum amount")

	pkg, err = svc.Update(context.Background(), "pkg-001", &UpdatePackageInput{
		MinimumAmount: &negativeMin,
	})
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.Contains(t, err.Error(), "minimum amount must be greater than or equal to zero")

	pkg, err = svc.Update(context.Background(), "pkg-001", &UpdatePackageInput{
		Fees: map[string]Fee{
			"fee": {
				CalculationModel: &CalculationModel{ApplicationRule: "flatFee", Calculations: []Calculation{{Type: "flat", Value: "0"}}},
			},
		},
	})
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.Contains(t, err.Error(), "calculation value must be greater than zero")
}

func TestPackagesUpdateBackendError(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("server error")
		},
	}

	svc := newPackagesService(backend)
	input := &UpdatePackageInput{FeeGroupLabel: "test"}

	pkg, err := svc.Update(context.Background(), "pkg-001", input)
	require.Error(t, err)
	assert.Nil(t, pkg)
}

// ---------------------------------------------------------------------------
// packagesServiceAPI.Delete
// ---------------------------------------------------------------------------

func TestPackagesDelete(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/packages/pkg-001", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newPackagesService(backend)

	err := svc.Delete(context.Background(), "pkg-001")
	require.NoError(t, err)
}

func TestPackagesDeleteEmptyID(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})

	err := svc.Delete(context.Background(), "")
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPackagesDeleteBackendError(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("delete failed")
		},
	}

	svc := newPackagesService(backend)

	err := svc.Delete(context.Background(), "pkg-001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
}

// ---------------------------------------------------------------------------
// buildPackagesListPath — query string builder
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Compile-time interface assertion
// ---------------------------------------------------------------------------

var _ packagesServiceAPI = (*packagesService)(nil)
