package leriantest

import (
	"context"
	"strings"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

func newCRMListIterator[T any](items []T, opts *models.PageListOptions) *pagination.Iterator[T] {
	pageSize := defaultFakePageSize
	if opts != nil && opts.PageSize > 0 {
		pageSize = opts.PageSize
	}

	initialPage := 0
	if opts != nil && opts.PageNumber > 0 {
		initialPage = opts.PageNumber
	}

	return pagination.NewPageIterator[T](initialPage, func(_ context.Context, page int) ([]T, int, int, int, error) {
		if page < 1 {
			page = 1
		}

		start := (page - 1) * pageSize
		if start >= len(items) {
			return nil, len(items), pageSize, page, nil
		}

		end := start + pageSize
		if end > len(items) {
			end = len(items)
		}

		pageItems := append([]T(nil), items[start:end]...)
		return pageItems, len(items), pageSize, page, nil
	})
}

func wantsDeletedRecords(opts *midaz.CRMListOptions) bool {
	if opts == nil {
		return false
	}

	return opts.IncludeDeleted
}

func wantsDescendingCRMOrder(sortOrder string) bool {
	if sortOrder == "" {
		return false
	}

	return strings.EqualFold(sortOrder, "desc")
}

func optsPageNumber(opts *midaz.CRMListOptions) int {
	if opts == nil {
		return 0
	}
	return opts.PageNumber
}

func optsPageSize(opts *midaz.CRMListOptions) int {
	if opts == nil {
		return 0
	}
	return opts.PageSize
}

func optsSortOrder(opts *midaz.CRMListOptions) string {
	if opts == nil {
		return ""
	}
	return opts.SortOrder
}

func aliasOptsPageNumber(opts *midaz.AliasListOptions) int {
	if opts == nil {
		return 0
	}
	return opts.PageNumber
}

func aliasOptsPageSize(opts *midaz.AliasListOptions) int {
	if opts == nil {
		return 0
	}
	return opts.PageSize
}

func aliasOptsSortOrder(opts *midaz.AliasListOptions) string {
	if opts == nil {
		return ""
	}
	return opts.SortOrder
}

func aliasOptsIncludeDeleted(opts *midaz.AliasListOptions) bool {
	if opts == nil {
		return false
	}
	return opts.IncludeDeleted
}

func aliasOptsHolderID(opts *midaz.AliasListOptions) string {
	if opts == nil {
		return ""
	}
	return opts.HolderID
}

func validateCRMOrgIDFake(operation, resource, orgID string) (string, error) {
	trimmed := strings.TrimSpace(orgID)
	if trimmed == "" {
		return "", sdkerrors.NewValidation(operation, resource, "organization id is required")
	}

	return trimmed, nil
}

// ---------------------------------------------------------------------------
// Holders (CRM)
// ---------------------------------------------------------------------------

type fakeHolders struct {
	store   *fakeStore[midaz.Holder]
	orgByID map[string]string
	cfg     *fakeConfig
}

func newFakeHolders(cfg *fakeConfig) *fakeHolders {
	return &fakeHolders{store: newFakeStore[midaz.Holder](), orgByID: make(map[string]string), cfg: cfg}
}

func (f *fakeHolders) Create(_ context.Context, orgID string, input *midaz.CreateHolderInput) (*midaz.Holder, error) {
	if err := f.cfg.injectedError("midaz.Holders.Create"); err != nil {
		return nil, err
	}

	trimmedOrgID, err := validateCRMOrgIDFake("Holders.Create", "Holder", orgID)
	if err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Holders.Create", "Holder", "input is required")
	}

	now := time.Now()

	holder := midaz.Holder{
		ID:            generateID("holder"),
		Name:          input.Name,
		Type:          input.Type,
		Document:      input.Document,
		Addresses:     input.Addresses,
		Contact:       input.Contact,
		NaturalPerson: input.NaturalPerson,
		LegalPerson:   input.LegalPerson,
		Metadata:      input.Metadata,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if input.ExternalID != nil {
		holder.ExternalID = *input.ExternalID
	}

	f.store.Set(holder.ID, holder)
	f.orgByID[holder.ID] = trimmedOrgID

	return &holder, nil
}

func (f *fakeHolders) Get(_ context.Context, orgID string, id string, opts *midaz.CRMGetOptions) (*midaz.Holder, error) {
	if err := f.cfg.injectedError("midaz.Holders.Get"); err != nil {
		return nil, err
	}

	trimmedOrgID, err := validateCRMOrgIDFake("Holders.Get", "Holder", orgID)
	if err != nil {
		return nil, err
	}

	if err := validateRequiredField("Holders.Get", "Holder", id, "id"); err != nil {
		return nil, err
	}

	holder, ok := f.store.Get(id)
	if !ok || f.orgByID[id] != trimmedOrgID {
		return nil, sdkerrors.NewNotFound("Holders.Get", "Holder", id)
	}

	includeDeleted := opts != nil && opts.IncludeDeleted
	if holder.DeletedAt != nil && !includeDeleted {
		return nil, sdkerrors.NewNotFound("Holders.Get", "Holder", id)
	}

	return &holder, nil
}

func (f *fakeHolders) List(_ context.Context, orgID string, opts *midaz.CRMListOptions) *pagination.Iterator[midaz.Holder] {
	if err := f.cfg.injectedError("midaz.Holders.List"); err != nil {
		return pagination.NewErrorIterator[midaz.Holder](err)
	}

	trimmedOrgID, err := validateCRMOrgIDFake("Holders.List", "Holder", orgID)
	if err != nil {
		return pagination.NewErrorIterator[midaz.Holder](err)
	}

	includeDeleted := wantsDeletedRecords(opts)
	holders := make([]midaz.Holder, 0, f.store.Len())
	for _, holder := range f.store.List() {
		if f.orgByID[holder.ID] != trimmedOrgID {
			continue
		}
		if !includeDeleted && holder.DeletedAt != nil {
			continue
		}
		holders = append(holders, holder)
	}

	if wantsDescendingCRMOrder(optsSortOrder(opts)) {
		for left, right := 0, len(holders)-1; left < right; left, right = left+1, right-1 {
			holders[left], holders[right] = holders[right], holders[left]
		}
	}

	return newCRMListIterator(holders, &models.PageListOptions{PageNumber: optsPageNumber(opts), PageSize: optsPageSize(opts), SortOrder: optsSortOrder(opts)})
}

func (f *fakeHolders) Update(_ context.Context, orgID string, id string, input *midaz.UpdateHolderInput) (*midaz.Holder, error) {
	if err := f.cfg.injectedError("midaz.Holders.Update"); err != nil {
		return nil, err
	}

	trimmedOrgID, err := validateCRMOrgIDFake("Holders.Update", "Holder", orgID)
	if err != nil {
		return nil, err
	}

	if err := validateRequiredField("Holders.Update", "Holder", id, "id"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Holders.Update", "Holder", "input is required")
	}

	holder, ok := f.store.Get(id)
	if !ok || f.orgByID[id] != trimmedOrgID {
		return nil, sdkerrors.NewNotFound("Holders.Update", "Holder", id)
	}

	if input.Name != nil {
		holder.Name = *input.Name
	}
	if input.ExternalID != nil {
		holder.ExternalID = *input.ExternalID
	}
	if input.Addresses != nil {
		holder.Addresses = input.Addresses
	}
	if input.Contact != nil {
		holder.Contact = input.Contact
	}
	if input.NaturalPerson != nil {
		holder.NaturalPerson = input.NaturalPerson
	}
	if input.LegalPerson != nil {
		holder.LegalPerson = input.LegalPerson
	}
	if input.Metadata != nil {
		holder.Metadata = input.Metadata
	}

	holder.UpdatedAt = time.Now()
	f.store.Set(id, holder)
	return &holder, nil
}

func (f *fakeHolders) Delete(_ context.Context, orgID string, id string, opts *midaz.CRMDeleteOptions) error {
	if err := f.cfg.injectedError("midaz.Holders.Delete"); err != nil {
		return err
	}

	trimmedOrgID, err := validateCRMOrgIDFake("Holders.Delete", "Holder", orgID)
	if err != nil {
		return err
	}

	if err := validateRequiredField("Holders.Delete", "Holder", id, "id"); err != nil {
		return err
	}

	holder, ok := f.store.Get(id)
	if !ok || f.orgByID[id] != trimmedOrgID {
		return sdkerrors.NewNotFound("Holders.Delete", "Holder", id)
	}

	if opts != nil && opts.HardDelete {
		f.store.Delete(id)
		delete(f.orgByID, id)
		return nil
	}

	now := time.Now()
	holder.DeletedAt = &now
	holder.UpdatedAt = now
	f.store.Set(id, holder)
	return nil
}

// ---------------------------------------------------------------------------
// Aliases (CRM)
// ---------------------------------------------------------------------------

type fakeAliases struct {
	store         *fakeStore[midaz.Alias]
	orgByID       map[string]string
	holders       *fakeStore[midaz.Holder]
	holderOrgByID map[string]string
	cfg           *fakeConfig
}

func newFakeAliases(cfg *fakeConfig, holders *fakeStore[midaz.Holder], holderOrgByID map[string]string) *fakeAliases {
	return &fakeAliases{
		store:         newFakeStore[midaz.Alias](),
		orgByID:       make(map[string]string),
		holders:       holders,
		holderOrgByID: holderOrgByID,
		cfg:           cfg,
	}
}

func (f *fakeAliases) holderVisible(orgID, holderID string) bool {
	holder, ok := f.holders.Get(holderID)
	if !ok {
		return false
	}
	if f.holderOrgByID[holderID] != orgID {
		return false
	}
	return holder.DeletedAt == nil
}

func (f *fakeAliases) Create(_ context.Context, orgID string, holderID string, input *midaz.CreateAliasInput) (*midaz.Alias, error) {
	if err := f.cfg.injectedError("midaz.Aliases.Create"); err != nil {
		return nil, err
	}

	trimmedOrgID, err := validateCRMOrgIDFake("Aliases.Create", "Alias", orgID)
	if err != nil {
		return nil, err
	}
	if err := validateRequiredField("Aliases.Create", "Alias", holderID, "holder id"); err != nil {
		return nil, err
	}
	if input == nil {
		return nil, sdkerrors.NewValidation("Aliases.Create", "Alias", "input is required")
	}
	if !f.holderVisible(trimmedOrgID, holderID) {
		return nil, sdkerrors.NewNotFound("Aliases.Create", "Holder", holderID)
	}

	now := time.Now()
	alias := midaz.Alias{
		ID:               generateID("alias"),
		HolderID:         holderID,
		LedgerID:         input.LedgerID,
		AccountID:        input.AccountID,
		BankingDetails:   input.BankingDetails,
		RegulatoryFields: input.RegulatoryFields,
		RelatedParties:   input.RelatedParties,
		Metadata:         input.Metadata,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	f.store.Set(alias.ID, alias)
	f.orgByID[alias.ID] = trimmedOrgID
	return &alias, nil
}

func (f *fakeAliases) Get(_ context.Context, orgID, holderID string, aliasID string, opts *midaz.CRMGetOptions) (*midaz.Alias, error) {
	if err := f.cfg.injectedError("midaz.Aliases.Get"); err != nil {
		return nil, err
	}

	trimmedOrgID, err := validateCRMOrgIDFake("Aliases.Get", "Alias", orgID)
	if err != nil {
		return nil, err
	}
	if err := validateRequiredField("Aliases.Get", "Alias", holderID, "holder id"); err != nil {
		return nil, err
	}
	if err := validateRequiredField("Aliases.Get", "Alias", aliasID, "alias id"); err != nil {
		return nil, err
	}
	if !f.holderVisible(trimmedOrgID, holderID) {
		return nil, sdkerrors.NewNotFound("Aliases.Get", "Holder", holderID)
	}

	alias, ok := f.store.Get(aliasID)
	includeDeleted := opts != nil && opts.IncludeDeleted
	if !ok || f.orgByID[aliasID] != trimmedOrgID || alias.HolderID != holderID || (alias.DeletedAt != nil && !includeDeleted) {
		return nil, sdkerrors.NewNotFound("Aliases.Get", "Alias", aliasID)
	}
	return &alias, nil
}

func (f *fakeAliases) List(_ context.Context, orgID string, opts *midaz.AliasListOptions) *pagination.Iterator[midaz.Alias] {
	if err := f.cfg.injectedError("midaz.Aliases.List"); err != nil {
		return pagination.NewErrorIterator[midaz.Alias](err)
	}

	trimmedOrgID, err := validateCRMOrgIDFake("Aliases.List", "Alias", orgID)
	if err != nil {
		return pagination.NewErrorIterator[midaz.Alias](err)
	}

	includeDeleted := aliasOptsIncludeDeleted(opts)
	holderIDFilter := aliasOptsHolderID(opts)

	aliases := make([]midaz.Alias, 0, f.store.Len())
	for _, alias := range f.store.List() {
		if f.orgByID[alias.ID] != trimmedOrgID {
			continue
		}
		if holderIDFilter != "" && alias.HolderID != holderIDFilter {
			continue
		}
		if !includeDeleted && alias.DeletedAt != nil {
			continue
		}
		aliases = append(aliases, alias)
	}

	if wantsDescendingCRMOrder(aliasOptsSortOrder(opts)) {
		for left, right := 0, len(aliases)-1; left < right; left, right = left+1, right-1 {
			aliases[left], aliases[right] = aliases[right], aliases[left]
		}
	}

	return newCRMListIterator(aliases, &models.PageListOptions{PageNumber: aliasOptsPageNumber(opts), PageSize: aliasOptsPageSize(opts), SortOrder: aliasOptsSortOrder(opts)})
}

func (f *fakeAliases) Update(_ context.Context, orgID, holderID string, aliasID string, input *midaz.UpdateAliasInput) (*midaz.Alias, error) {
	if err := f.cfg.injectedError("midaz.Aliases.Update"); err != nil {
		return nil, err
	}

	trimmedOrgID, err := validateCRMOrgIDFake("Aliases.Update", "Alias", orgID)
	if err != nil {
		return nil, err
	}
	if err := validateRequiredField("Aliases.Update", "Alias", holderID, "holder id"); err != nil {
		return nil, err
	}
	if err := validateRequiredField("Aliases.Update", "Alias", aliasID, "alias id"); err != nil {
		return nil, err
	}
	if input == nil {
		return nil, sdkerrors.NewValidation("Aliases.Update", "Alias", "input is required")
	}
	if !f.holderVisible(trimmedOrgID, holderID) {
		return nil, sdkerrors.NewNotFound("Aliases.Update", "Holder", holderID)
	}

	alias, ok := f.store.Get(aliasID)
	if !ok || f.orgByID[aliasID] != trimmedOrgID || alias.HolderID != holderID {
		return nil, sdkerrors.NewNotFound("Aliases.Update", "Alias", aliasID)
	}

	if input.BankingDetails != nil {
		alias.BankingDetails = input.BankingDetails
	}
	if input.RegulatoryFields != nil {
		alias.RegulatoryFields = input.RegulatoryFields
	}
	if input.RelatedParties != nil {
		alias.RelatedParties = *input.RelatedParties
	}
	if input.Metadata != nil {
		alias.Metadata = input.Metadata
	}

	alias.UpdatedAt = time.Now()
	f.store.Set(aliasID, alias)
	return &alias, nil
}

func (f *fakeAliases) Delete(_ context.Context, orgID, holderID, aliasID string, opts *midaz.CRMDeleteOptions) error {
	if err := f.cfg.injectedError("midaz.Aliases.Delete"); err != nil {
		return err
	}

	trimmedOrgID, err := validateCRMOrgIDFake("Aliases.Delete", "Alias", orgID)
	if err != nil {
		return err
	}
	if err := validateRequiredField("Aliases.Delete", "Alias", holderID, "holder id"); err != nil {
		return err
	}
	if err := validateRequiredField("Aliases.Delete", "Alias", aliasID, "alias id"); err != nil {
		return err
	}
	if !f.holderVisible(trimmedOrgID, holderID) {
		return sdkerrors.NewNotFound("Aliases.Delete", "Holder", holderID)
	}

	alias, ok := f.store.Get(aliasID)
	if !ok || f.orgByID[aliasID] != trimmedOrgID || alias.HolderID != holderID {
		return sdkerrors.NewNotFound("Aliases.Delete", "Alias", aliasID)
	}

	if opts != nil && opts.HardDelete {
		f.store.Delete(aliasID)
		delete(f.orgByID, aliasID)
		return nil
	}

	now := time.Now()
	alias.DeletedAt = &now
	alias.UpdatedAt = now
	f.store.Set(aliasID, alias)
	return nil
}

func (f *fakeAliases) DeleteRelatedParty(_ context.Context, orgID, holderID, aliasID, relatedPartyID string) error {
	if err := f.cfg.injectedError("midaz.Aliases.DeleteRelatedParty"); err != nil {
		return err
	}

	trimmedOrgID, err := validateCRMOrgIDFake("Aliases.DeleteRelatedParty", "Alias", orgID)
	if err != nil {
		return err
	}
	if err := validateRequiredField("Aliases.DeleteRelatedParty", "Alias", holderID, "holder id"); err != nil {
		return err
	}
	if err := validateRequiredField("Aliases.DeleteRelatedParty", "Alias", aliasID, "alias id"); err != nil {
		return err
	}
	if err := validateRequiredField("Aliases.DeleteRelatedParty", "Alias", relatedPartyID, "related party id"); err != nil {
		return err
	}
	if !f.holderVisible(trimmedOrgID, holderID) {
		return sdkerrors.NewNotFound("Aliases.DeleteRelatedParty", "Holder", holderID)
	}

	alias, ok := f.store.Get(aliasID)
	if !ok || f.orgByID[aliasID] != trimmedOrgID || alias.HolderID != holderID {
		return sdkerrors.NewNotFound("Aliases.DeleteRelatedParty", "Alias", aliasID)
	}

	filtered := alias.RelatedParties[:0]
	removed := false
	for _, relatedParty := range alias.RelatedParties {
		if relatedParty.ID == relatedPartyID {
			removed = true
			continue
		}
		filtered = append(filtered, relatedParty)
	}

	if !removed {
		return sdkerrors.NewNotFound("Aliases.DeleteRelatedParty", "RelatedParty", relatedPartyID)
	}

	alias.RelatedParties = append([]midaz.RelatedParty(nil), filtered...)
	alias.UpdatedAt = time.Now()
	f.store.Set(aliasID, alias)
	return nil
}
