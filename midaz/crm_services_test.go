package midaz

import (
	"context"
	"errors"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHoldersGetIncludesOrgHeaderAndQuery(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callWithHdrsFn: func(_ context.Context, method, path string, headers map[string]string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/holders/holder-1?include_deleted=true", path)
			assert.Equal(t, "org-1", headers["X-Organization-Id"])
			assert.Nil(t, body)

			return unmarshalInto(Holder{ID: "holder-1", Name: "Alice"}, result)
		},
	}

	svc := newHoldersService(mock)
	holder, err := svc.Get(context.Background(), "org-1", "holder-1", &CRMGetOptions{IncludeDeleted: true})

	require.NoError(t, err)
	require.NotNil(t, holder)
	assert.Equal(t, "holder-1", holder.ID)
}

func TestHoldersDeleteIncludesHardDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callWithHdrsFn: func(_ context.Context, method, path string, headers map[string]string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/holders/holder-1?hard_delete=true", path)
			assert.Equal(t, "org-1", headers["X-Organization-Id"])
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newHoldersService(mock)
	err := svc.Delete(context.Background(), "org-1", "holder-1", &CRMDeleteOptions{HardDelete: true})
	require.NoError(t, err)
}

func TestHoldersCreateUsesOrgHeader(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callWithHdrsFn: func(_ context.Context, method, path string, headers map[string]string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/holders", path)
			assert.Equal(t, "org-1", headers["X-Organization-Id"])

			input, ok := body.(*CreateHolderInput)
			require.True(t, ok)
			assert.Equal(t, "Alice", input.Name)

			return unmarshalInto(Holder{ID: "holder-1", Name: "Alice"}, result)
		},
	}

	svc := newHoldersService(mock)
	holder, err := svc.Create(context.Background(), "org-1", &CreateHolderInput{Name: "Alice", Type: "individual", Document: "123"})
	require.NoError(t, err)
	require.NotNil(t, holder)
	assert.Equal(t, "holder-1", holder.ID)
}

func TestHoldersUpdateUsesOrgHeader(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callWithHdrsFn: func(_ context.Context, method, path string, headers map[string]string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/holders/holder-1", path)
			assert.Equal(t, "org-1", headers["X-Organization-Id"])

			return unmarshalInto(Holder{ID: "holder-1", Name: "Alice Updated"}, result)
		},
	}

	svc := newHoldersService(mock)
	holder, err := svc.Update(context.Background(), "org-1", "holder-1", &UpdateHolderInput{Name: ptr("Alice Updated")})
	require.NoError(t, err)
	require.NotNil(t, holder)
	assert.Equal(t, "Alice Updated", holder.Name)
}

func TestAliasesCreateUsesHolderScopedPath(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callWithHdrsFn: func(_ context.Context, method, path string, headers map[string]string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/holders/holder-1/aliases", path)
			assert.Equal(t, "org-1", headers["X-Organization-Id"])

			input, ok := body.(*CreateAliasInput)
			require.True(t, ok)
			assert.Equal(t, "ledger-1", input.LedgerID)

			return unmarshalInto(Alias{ID: "alias-1", HolderID: "holder-1", LedgerID: "ledger-1"}, result)
		},
	}

	svc := newAliasesService(mock)
	alias, err := svc.Create(context.Background(), "org-1", "holder-1", &CreateAliasInput{LedgerID: "ledger-1", AccountID: "acc-1"})

	require.NoError(t, err)
	require.NotNil(t, alias)
	assert.Equal(t, "alias-1", alias.ID)
}

func TestAliasesGetUsesHolderScopedPath(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callWithHdrsFn: func(_ context.Context, method, path string, headers map[string]string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/holders/holder-1/aliases/alias-1", path)
			assert.Equal(t, "org-1", headers["X-Organization-Id"])
			assert.Nil(t, body)

			return unmarshalInto(Alias{ID: "alias-1", HolderID: "holder-1"}, result)
		},
	}

	svc := newAliasesService(mock)
	alias, err := svc.Get(context.Background(), "org-1", "holder-1", "alias-1", nil)
	require.NoError(t, err)
	require.NotNil(t, alias)
	assert.Equal(t, "alias-1", alias.ID)
}

func TestAliasesUpdateUsesHolderScopedPath(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callWithHdrsFn: func(_ context.Context, method, path string, headers map[string]string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/holders/holder-1/aliases/alias-1", path)
			assert.Equal(t, "org-1", headers["X-Organization-Id"])

			return unmarshalInto(Alias{ID: "alias-1", HolderID: "holder-1", Metadata: models.Metadata{"updated": true}}, result)
		},
	}

	svc := newAliasesService(mock)
	alias, err := svc.Update(context.Background(), "org-1", "holder-1", "alias-1", &UpdateAliasInput{Metadata: models.Metadata{"updated": true}})
	require.NoError(t, err)
	require.NotNil(t, alias)
	assert.Equal(t, true, alias.Metadata["updated"])
}

func TestHoldersListIteratorFetchesAllPages(t *testing.T) {
	t.Parallel()

	requestCount := 0
	mock := &mockBackend{
		callWithHdrsFn: func(_ context.Context, method, path string, headers map[string]string, body, result any) error {
			requestCount++

			assert.Equal(t, "GET", method)
			assert.Equal(t, "org-1", headers["X-Organization-Id"])
			assert.Nil(t, body)

			switch requestCount {
			case 1:
				assert.Equal(t, "/holders?limit=2", path)

				return unmarshalInto(models.ListResponse[Holder]{
					Items:      []Holder{{ID: "holder-1"}, {ID: "holder-2"}},
					Pagination: models.Pagination{Total: 3, Limit: 2},
				}, result)
			case 2:
				assert.Equal(t, "/holders?limit=2&page=2", path)

				return unmarshalInto(models.ListResponse[Holder]{
					Items:      []Holder{{ID: "holder-3"}},
					Pagination: models.Pagination{Total: 3, Page: 2, Limit: 2},
				}, result)
			default:
				return errors.New("unexpected extra CRM page fetch")
			}
		},
	}

	svc := newHoldersService(mock)
	iter := svc.List(context.Background(), "org-1", &CRMListOptions{PageSize: 2})
	holders, err := iter.Collect(context.Background())

	require.NoError(t, err)
	assert.Len(t, holders, 3)
	assert.Equal(t, []string{"holder-1", "holder-2", "holder-3"}, []string{holders[0].ID, holders[1].ID, holders[2].ID})
	assert.Equal(t, 2, requestCount)
}

func TestAliasesListUsesOrgHeaderAndOptions(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callWithHdrsFn: func(_ context.Context, method, path string, headers map[string]string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/aliases?")
			assert.Contains(t, path, "holder_id=holder-1")
			assert.Contains(t, path, "include_deleted=true")
			assert.Contains(t, path, "limit=25")
			assert.Contains(t, path, "page=2")
			assert.Contains(t, path, "sort_order=desc")
			assert.Equal(t, "org-1", headers["X-Organization-Id"])
			assert.Nil(t, body)

			return unmarshalInto(models.ListResponse[Alias]{
				Items:      []Alias{{ID: "alias-1", HolderID: "holder-1"}},
				Pagination: models.Pagination{Total: 1, Page: 2, Limit: 25},
			}, result)
		},
	}

	svc := newAliasesService(mock)
	iter := svc.List(context.Background(), "org-1", &AliasListOptions{
		CRMListOptions: CRMListOptions{PageSize: 25, PageNumber: 2, SortOrder: "desc", IncludeDeleted: true},
		HolderID:       "holder-1",
	})
	items, err := iter.Collect(context.Background())

	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestAliasesDeleteIncludesHardDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callWithHdrsFn: func(_ context.Context, method, path string, headers map[string]string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/holders/holder-1/aliases/alias-1?hard_delete=true", path)
			assert.Equal(t, "org-1", headers["X-Organization-Id"])
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newAliasesService(mock)
	err := svc.Delete(context.Background(), "org-1", "holder-1", "alias-1", &CRMDeleteOptions{HardDelete: true})
	require.NoError(t, err)
}

func TestAliasesDeleteRelatedPartyPath(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callWithHdrsFn: func(_ context.Context, method, path string, headers map[string]string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/holders/holder-1/aliases/alias-1/related-parties/rp-1", path)
			assert.Equal(t, "org-1", headers["X-Organization-Id"])
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newAliasesService(mock)
	err := svc.DeleteRelatedParty(context.Background(), "org-1", "holder-1", "alias-1", "rp-1")
	require.NoError(t, err)
}

func TestCRMServicesTrimOrganizationID(t *testing.T) {
	t.Parallel()

	t.Run("holders get", func(t *testing.T) {
		t.Parallel()

		mock := &mockBackend{callWithHdrsFn: func(_ context.Context, method, path string, headers map[string]string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/holders/holder-1", path)
			assert.Equal(t, "org-1", headers["X-Organization-Id"])

			return unmarshalInto(Holder{ID: "holder-1"}, result)
		}}

		svc := newHoldersService(mock)
		holder, err := svc.Get(context.Background(), "  org-1\t", "holder-1", nil)
		require.NoError(t, err)
		require.NotNil(t, holder)
	})

	t.Run("aliases list rejects whitespace only", func(t *testing.T) {
		t.Parallel()

		svc := newAliasesService(&mockBackend{})
		iter := svc.List(context.Background(), " \t ", nil)
		items, err := iter.Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("holders get rejects whitespace id", func(t *testing.T) {
		t.Parallel()

		svc := newHoldersService(&mockBackend{})
		holder, err := svc.Get(context.Background(), "org-1", "   ", nil)
		require.Error(t, err)
		assert.Nil(t, holder)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("aliases get rejects whitespace ids", func(t *testing.T) {
		t.Parallel()

		svc := newAliasesService(&mockBackend{})
		alias, err := svc.Get(context.Background(), "org-1", "   ", " \t ", nil)
		require.Error(t, err)
		assert.Nil(t, alias)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestCRMServicesRejectNilReceivers(t *testing.T) {
	t.Parallel()

	t.Run("holders get", func(t *testing.T) {
		t.Parallel()

		var svc *holdersService

		holder, err := svc.Get(context.Background(), "org-1", "holder-1", nil)
		require.Error(t, err)
		assert.Nil(t, holder)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("aliases delete", func(t *testing.T) {
		t.Parallel()

		var svc *aliasesService

		err := svc.Delete(context.Background(), "org-1", "holder-1", "alias-1", nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("holders list", func(t *testing.T) {
		t.Parallel()

		var svc *holdersService

		iter := svc.List(context.Background(), "org-1", nil)
		items, err := iter.Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, core.ErrNilService)
	})
}

func TestBuildCRMListPath(t *testing.T) {
	t.Parallel()

	t.Run("nil options", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "/holders", buildCRMListPath("/holders", nil, 0))
	})

	t.Run("with page sort order and filters", func(t *testing.T) {
		t.Parallel()

		path := buildCRMAliasListPath("/aliases", &AliasListOptions{
			CRMListOptions: CRMListOptions{PageSize: 10, SortOrder: "asc"},
			HolderID:       "holder-1",
		}, 3)

		assert.Contains(t, path, "/aliases?")
		assert.Contains(t, path, "holder_id=holder-1")
		assert.Contains(t, path, "limit=10")
		assert.Contains(t, path, "page=3")
		assert.Contains(t, path, "sort_order=asc")
	})
}
