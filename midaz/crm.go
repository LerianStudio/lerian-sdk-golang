// crm.go provides helper functions for CRM API services that require
// the X-Organization-Id header on every request. Unlike the onboarding and
// transaction APIs which encode the organization in the URL path, the CRM
// API uses a dedicated header for organization context.
package midaz

import (
	"net/url"
	"strconv"
	"strings"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

// crmHeaders builds the standard CRM request headers containing the
// organization context.
func crmHeaders(orgID string) map[string]string {
	return map[string]string{
		"X-Organization-Id": orgID,
	}
}

func validateCRMOrgID(operation, resource, orgID string) (string, error) {
	trimmed := strings.TrimSpace(orgID)
	if trimmed == "" {
		return "", sdkerrors.NewValidation(operation, resource, "organization id is required")
	}

	return trimmed, nil
}

func validateCRMIdentifier(operation, resource, value, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", sdkerrors.NewValidation(operation, resource, field+" is required")
	}

	return trimmed, nil
}

func buildCRMQueryPath(path string, params url.Values) string {
	if len(params) == 0 {
		return path
	}

	return path + "?" + params.Encode()
}

func crmCollectionPath(collection string) string {
	return "/" + collection
}

func crmItemPath(collection, id string) string {
	return crmCollectionPath(collection) + "/" + url.PathEscape(id)
}

func crmNestedCollectionPath(parentCollection, parentID, collection string) string {
	return crmItemPath(parentCollection, parentID) + "/" + collection
}

func crmNestedItemPath(parentCollection, parentID, collection, id string) string {
	return crmNestedCollectionPath(parentCollection, parentID, collection) + "/" + url.PathEscape(id)
}

func applyCRMListOptions(params url.Values, page int, opts *CRMListOptions) {
	if opts != nil {
		if opts.PageSize > 0 {
			params.Set("limit", strconv.Itoa(opts.PageSize))
		}

		if opts.SortOrder != "" {
			params.Set("sort_order", opts.SortOrder)
		}

		if opts.IncludeDeleted {
			params.Set("include_deleted", "true")
		}
	}

	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}
}

// buildCRMListPath constructs CRM list query parameters.
func buildCRMListPath(path string, opts *CRMListOptions, page int) string {
	params := url.Values{}
	applyCRMListOptions(params, page, opts)

	return buildCRMQueryPath(path, params)
}

func buildCRMAliasListPath(path string, opts *AliasListOptions, page int) string {
	params := url.Values{}
	if opts != nil {
		applyCRMListOptions(params, page, &opts.CRMListOptions)

		if strings.TrimSpace(opts.HolderID) != "" {
			params.Set("holder_id", opts.HolderID)
		}
	} else {
		applyCRMListOptions(params, page, nil)
	}

	return buildCRMQueryPath(path, params)
}

func initialCRMPage(opts *CRMListOptions) int {
	if opts != nil && opts.PageNumber > 0 {
		return opts.PageNumber
	}

	return 0
}

func initialCRMAliasPage(opts *AliasListOptions) int {
	if opts != nil && opts.PageNumber > 0 {
		return opts.PageNumber
	}

	return 0
}

func applyCRMGetOptions(path string, opts *CRMGetOptions) string {
	if opts == nil || !opts.IncludeDeleted {
		return path
	}

	return buildCRMQueryPath(path, url.Values{"include_deleted": []string{"true"}})
}

func applyCRMDeleteOptions(path string, opts *CRMDeleteOptions) string {
	if opts == nil || !opts.HardDelete {
		return path
	}

	return buildCRMQueryPath(path, url.Values{"hard_delete": []string{"true"}})
}
