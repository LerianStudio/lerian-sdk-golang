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
//   - [ListOptions] -- configures pagination, sorting, and filtering for
//     List operations across all products
//   - [ListResponse] -- generic paginated response envelope with typed items
//   - [Pagination] -- page metadata (total count, cursors, limits)
//   - [Status] -- entity status with optional description
//   - [Metadata] -- free-form key-value map attached to entities
//   - [Address] -- physical address representation
package models
