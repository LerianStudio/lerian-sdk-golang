package reporter

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// DataSourcesService provides read-only access to Reporter data-source
// endpoints. Data sources represent the upstream systems (databases, APIs,
// streams) from which reports pull their data.
type DataSourcesService interface {
	// Get retrieves a single data source by its unique identifier.
	Get(ctx context.Context, id string) (*DataSource, error)

	// List returns a paginated iterator over all available data sources.
	List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[DataSource]
}

// dataSourcesService is the concrete implementation of [DataSourcesService].
type dataSourcesService struct {
	core.BaseService
}

// Compile-time interface compliance check.
var _ DataSourcesService = (*dataSourcesService)(nil)

// newDataSourcesService constructs a [DataSourcesService] backed by the
// given [core.Backend].
func newDataSourcesService(backend core.Backend) DataSourcesService {
	return &dataSourcesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Get retrieves a single data source by ID.
func (s *dataSourcesService) Get(ctx context.Context, id string) (*DataSource, error) {
	const operation = "DataSources.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "DataSource", "id is required")
	}

	return core.Get[DataSource](ctx, &s.BaseService, "/datasources/"+url.PathEscape(id))
}

// List returns a paginated iterator over data sources.
func (s *dataSourcesService) List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[DataSource] {
	return core.List[DataSource](ctx, &s.BaseService, "/datasources", opts)
}
