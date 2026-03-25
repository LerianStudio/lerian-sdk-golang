package midaz

// CRMGetOptions controls visibility behavior for CRM get operations.
type CRMGetOptions struct {
	IncludeDeleted bool
}

// CRMDeleteOptions controls delete behavior for CRM delete operations.
type CRMDeleteOptions struct {
	HardDelete bool
}

// CRMListOptions controls page-based listing for CRM resources.
type CRMListOptions struct {
	PageNumber     int
	PageSize       int
	SortOrder      string
	IncludeDeleted bool
}

// AliasListOptions controls alias listing for CRM resources.
type AliasListOptions struct {
	CRMListOptions
	HolderID string
}
