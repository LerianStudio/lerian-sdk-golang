package leriantest

import (
	"context"
	"reflect"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/fees"
	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// newFakeFeesClient constructs a [fees.Client] with all service fields
// backed by in-memory fakes.
func newFakeFeesClient(cfg *fakeConfig) *fees.Client {
	return &fees.Client{
		Packages:  &fakeFeesPackages{store: newFakeStore[fees.Package](), cfg: cfg},
		Estimates: &fakeFeesEstimates{cfg: cfg},
		Fees:      &fakeFeesFees{cfg: cfg},
	}
}

// ---------------------------------------------------------------------------
// Packages
// ---------------------------------------------------------------------------

type fakeFeesPackages struct {
	store *fakeStore[fees.Package]
	cfg   *fakeConfig
}

var _ fees.PackagesService = (*fakeFeesPackages)(nil)

func (f *fakeFeesPackages) Create(_ context.Context, input *fees.CreatePackageInput) (*fees.Package, error) {
	if err := f.cfg.injectedError("fees.Packages.Create"); err != nil {
		return nil, err
	}

	now := time.Now()

	p := fees.Package{
		ID:        generateID("fpkg"),
		Name:      input.Name,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	f.store.Set(p.ID, p)

	return &p, nil
}

func (f *fakeFeesPackages) Get(_ context.Context, id string) (*fees.Package, error) {
	p, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Packages.Get", "Package", id)
	}

	return &p, nil
}

func (f *fakeFeesPackages) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[fees.Package] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeFeesPackages) Update(_ context.Context, id string, _ *fees.UpdatePackageInput) (*fees.Package, error) {
	p, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Packages.Update", "Package", id)
	}

	p.UpdatedAt = time.Now()
	f.store.Set(id, p)

	return &p, nil
}

func (f *fakeFeesPackages) Delete(_ context.Context, id string) error {
	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("Packages.Delete", "Package", id)
	}

	f.store.Delete(id)

	return nil
}

// ---------------------------------------------------------------------------
// Estimates (RPC-style, no store)
// ---------------------------------------------------------------------------

type fakeFeesEstimates struct {
	cfg *fakeConfig
}

var _ fees.EstimatesService = (*fakeFeesEstimates)(nil)

func (f *fakeFeesEstimates) Calculate(_ context.Context, _ *fees.CalculateEstimateInput) (*fees.Estimate, error) {
	if err := f.cfg.injectedError("fees.Estimates.Calculate"); err != nil {
		return nil, err
	}

	return &fees.Estimate{
		ID: generateID("fest"),
	}, nil
}

// ---------------------------------------------------------------------------
// Fees (RPC-style, no store)
// ---------------------------------------------------------------------------

type fakeFeesFees struct {
	cfg *fakeConfig
}

var _ fees.FeesService = (*fakeFeesFees)(nil)

func (f *fakeFeesFees) Calculate(_ context.Context, _ *fees.CalculateFeeInput) (*fees.Fee, error) {
	if err := f.cfg.injectedError("fees.Fees.Calculate"); err != nil {
		return nil, err
	}

	return &fees.Fee{
		ID: generateID("ffee"),
	}, nil
}

func (f *fakeFeesFees) TransformTransaction(_ context.Context, input *fees.TransformTransactionInput) (*fees.TransformTransactionOutput, error) {
	if err := f.cfg.injectedError("fees.Fees.TransformTransaction"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Fees.TransformTransaction", "Fee", "input is required")
	}

	if input.LedgerID == "" {
		return nil, sdkerrors.NewValidation("Fees.TransformTransaction", "Fee", "ledger ID is required")
	}

	return &fees.TransformTransactionOutput{
		Transaction: cloneTransactionDSL(input.Transaction),
	}, nil
}

func cloneTransactionDSL(input fees.TransactionDSL) fees.TransactionDSL {
	output := input
	output.Metadata = cloneMap(input.Metadata)
	output.RouteID = cloneStringPointer(input.RouteID)
	output.TransactionDate = cloneAny(input.TransactionDate)
	output.Send = cloneTransactionDSLSend(input.Send)

	return output
}

func cloneTransactionDSLSend(input fees.TransactionDSLSend) fees.TransactionDSLSend {
	return fees.TransactionDSLSend{
		Asset:      input.Asset,
		Value:      cloneAny(input.Value),
		Source:     cloneTransactionDSLSource(input.Source),
		Distribute: cloneTransactionDSLDistribute(input.Distribute),
	}
}

func cloneTransactionDSLSource(input fees.TransactionDSLSource) fees.TransactionDSLSource {
	output := fees.TransactionDSLSource{
		Remaining: input.Remaining,
		From:      make([]fees.TransactionDSLLeg, 0, len(input.From)),
	}

	for _, leg := range input.From {
		output.From = append(output.From, cloneTransactionDSLLeg(leg))
	}

	return output
}

func cloneTransactionDSLDistribute(input fees.TransactionDSLDistribute) fees.TransactionDSLDistribute {
	output := fees.TransactionDSLDistribute{
		Remaining: input.Remaining,
		To:        make([]fees.TransactionDSLLeg, 0, len(input.To)),
	}

	for _, leg := range input.To {
		output.To = append(output.To, cloneTransactionDSLLeg(leg))
	}

	return output
}

func cloneTransactionDSLLeg(input fees.TransactionDSLLeg) fees.TransactionDSLLeg {
	output := input
	output.Amount = cloneTransactionDSLAmount(input.Amount)
	output.Share = cloneTransactionDSLShare(input.Share)
	output.Rate = cloneTransactionDSLRate(input.Rate)
	output.RouteID = cloneStringPointer(input.RouteID)
	output.Metadata = cloneMap(input.Metadata)

	return output
}

func cloneTransactionDSLAmount(input *fees.TransactionDSLAmount) *fees.TransactionDSLAmount {
	if input == nil {
		return nil
	}

	output := *input
	output.Value = cloneAny(input.Value)

	return &output
}

func cloneTransactionDSLShare(input *fees.TransactionDSLShare) *fees.TransactionDSLShare {
	if input == nil {
		return nil
	}

	output := *input

	return &output
}

func cloneTransactionDSLRate(input *fees.TransactionDSLRate) *fees.TransactionDSLRate {
	if input == nil {
		return nil
	}

	output := *input
	output.Value = cloneAny(input.Value)

	return &output
}

func cloneStringPointer(input *string) *string {
	if input == nil {
		return nil
	}

	value := *input

	return &value
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}

	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = cloneAny(value)
	}

	return output
}

func cloneAny(input any) any {
	if input == nil {
		return nil
	}

	return cloneReflectValue(reflect.ValueOf(input)).Interface()
}

func cloneReflectValue(input reflect.Value) reflect.Value {
	if !input.IsValid() {
		return input
	}

	switch input.Kind() {
	case reflect.Interface:
		if input.IsNil() {
			return reflect.Zero(input.Type())
		}

		cloned := cloneReflectValue(input.Elem())
		output := reflect.New(input.Type()).Elem()
		output.Set(cloned)

		return output
	case reflect.Pointer:
		if input.IsNil() {
			return reflect.Zero(input.Type())
		}

		output := reflect.New(input.Type().Elem())
		output.Elem().Set(cloneReflectValue(input.Elem()))

		return output
	case reflect.Map:
		if input.IsNil() {
			return reflect.Zero(input.Type())
		}

		output := reflect.MakeMapWithSize(input.Type(), input.Len())
		iter := input.MapRange()
		for iter.Next() {
			output.SetMapIndex(iter.Key(), cloneReflectValue(iter.Value()))
		}

		return output
	case reflect.Slice:
		if input.IsNil() {
			return reflect.Zero(input.Type())
		}

		output := reflect.MakeSlice(input.Type(), input.Len(), input.Len())
		for i := 0; i < input.Len(); i++ {
			output.Index(i).Set(cloneReflectValue(input.Index(i)))
		}

		return output
	case reflect.Array:
		output := reflect.New(input.Type()).Elem()
		for i := 0; i < input.Len(); i++ {
			output.Index(i).Set(cloneReflectValue(input.Index(i)))
		}

		return output
	default:
		return input
	}
}
