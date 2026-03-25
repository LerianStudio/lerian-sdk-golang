package fees

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

// packagesServiceAPI manages fee packages. A package groups one or more
// [Fee] definitions that are evaluated together when calculating
// fees for a transaction.
type packagesServiceAPI interface {
	// Create creates a new fee package from the provided input.
	Create(ctx context.Context, input *CreatePackageInput) (*Package, error)

	// Get retrieves a fee package by its unique identifier.
	Get(ctx context.Context, id string) (*Package, error)

	// List returns a normalized page of fee packages matching the given filter
	// options.
	List(ctx context.Context, opts *PackageListOptions) (*PackagePage, error)

	// Update partially updates an existing fee package.
	Update(ctx context.Context, id string, input *UpdatePackageInput) (*Package, error)

	// Delete removes a fee package by its unique identifier.
	Delete(ctx context.Context, id string) error
}

// packagesService is the concrete implementation of [packagesServiceAPI].
type packagesService struct {
	core.BaseService
}

// Compile-time interface compliance check.
var _ packagesServiceAPI = (*packagesService)(nil)

// newPackagesService constructs a [packagesServiceAPI] backed by the given
// [core.Backend].
func newPackagesService(backend core.Backend) packagesServiceAPI {
	return &packagesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Create creates a new fee package.
func (s *packagesService) Create(ctx context.Context, input *CreatePackageInput) (*Package, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	return core.Create[Package, CreatePackageInput](ctx, &s.BaseService, "/packages", input)
}

// Get retrieves a fee package by ID.
func (s *packagesService) Get(ctx context.Context, id string) (*Package, error) {
	const operation = "Packages.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Package", "id is required")
	}

	return core.Get[Package](ctx, &s.BaseService, "/packages/"+url.PathEscape(id))
}

// List returns a paginated list of fee packages. The plugin uses a flat
// pagination envelope and supports domain-specific query parameters
// (segmentId, ledgerId, transactionRoute, enable) alongside standard
// pagination params.
func (s *packagesService) List(ctx context.Context, opts *PackageListOptions) (*PackagePage, error) {
	if s == nil {
		return nil, core.ErrNilService
	}

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	backend, err := core.ResolveBackend(&s.BaseService)
	if err != nil {
		return nil, err
	}

	queryPath := buildPackagesListPath(opts)

	res, err := backend.Do(ctx, core.Request{Method: "GET", Path: queryPath})
	if err != nil {
		return nil, err
	}

	var raw struct {
		Items []Package `json:"items"`
		Page  int       `json:"page"`
		Limit int       `json:"limit"`
		Total int       `json:"total"`
	}
	if err := json.Unmarshal(res.Body, &raw); err != nil {
		return nil, sdkerrors.NewInternal("fees", "Packages.List", "failed to unmarshal response body", err)
	}

	totalPages := 0
	if raw.Limit > 0 {
		totalPages = (raw.Total + raw.Limit - 1) / raw.Limit
	}

	return &PackagePage{
		Items:      raw.Items,
		PageNumber: raw.Page,
		PageSize:   raw.Limit,
		TotalItems: raw.Total,
		TotalPages: totalPages,
	}, nil
}

// Update partially updates an existing fee package.
func (s *packagesService) Update(ctx context.Context, id string, input *UpdatePackageInput) (*Package, error) {
	if id == "" {
		return nil, sdkerrors.NewValidation("Packages.Update", "Package", "id is required")
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	return core.Update[Package, UpdatePackageInput](ctx, &s.BaseService, "/packages/"+url.PathEscape(id), input)
}

// Delete removes a fee package by ID.
func (s *packagesService) Delete(ctx context.Context, id string) error {
	const operation = "Packages.Delete"

	if id == "" {
		return sdkerrors.NewValidation(operation, "Package", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/packages/"+url.PathEscape(id))
}

// buildPackagesListPath constructs the query string for the packages list
// endpoint. The plugin expects bare query parameters (not filter[key]=value)
// and uses snake_case for some pagination params (sort_order, start_date,
// end_date).
func buildPackagesListPath(opts *PackageListOptions) string {
	const basePath = "/packages"

	if opts == nil {
		return basePath
	}

	params := url.Values{}

	// Domain-specific filters
	if opts.SegmentID != "" {
		params.Set("segmentId", opts.SegmentID)
	}

	if opts.LedgerID != "" {
		params.Set("ledgerId", opts.LedgerID)
	}

	if opts.TransactionRoute != "" {
		params.Set("transactionRoute", opts.TransactionRoute)
	}

	if opts.Enabled != nil {
		params.Set("enable", strconv.FormatBool(*opts.Enabled))
	}

	// Pagination params (plugin uses snake_case for some)
	if opts.PageSize > 0 {
		params.Set("limit", strconv.Itoa(opts.PageSize))
	}

	if opts.PageNumber > 0 {
		params.Set("page", strconv.Itoa(opts.PageNumber))
	}

	if opts.SortOrder != "" {
		params.Set("sort_order", opts.SortOrder)
	}

	if opts.CreatedFrom != nil {
		params.Set("start_date", opts.CreatedFrom.Format("2006-01-02"))
	}

	if opts.CreatedTo != nil {
		params.Set("end_date", opts.CreatedTo.Format("2006-01-02"))
	}

	if len(params) == 0 {
		return basePath
	}

	return basePath + "?" + params.Encode()
}
