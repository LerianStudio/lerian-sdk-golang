// Package models provides shared cross-product types used across all Lerian
// SDK products.
//
// These types define the common data structures that appear in multiple
// product APIs, such as pagination envelopes, list options, status
// representations, metadata maps, and address structures. Product-specific
// types live in their respective packages (e.g. [midaz], [matcher]).
//
// # Key Types
//
//   - [CursorListOptions] -- configures cursor-based pagination, sorting, and
//     filtering for list operations across products
//   - [PageListOptions] -- configures page-based pagination for APIs without
//     cursor semantics
//   - [ListResponse] -- generic paginated response envelope with typed items
//   - [Pagination] -- page metadata (total count, cursors, limits)
//   - [Status] -- entity status with optional description
//   - [Metadata] -- free-form key-value map attached to entities
//   - [Address] -- physical address representation
package models
