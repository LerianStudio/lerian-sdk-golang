// Package leriantest provides in-memory fake implementations of all Lerian
// SDK product services for consumer testing.
//
// Instead of requiring real HTTP services or complex mock setups, consumers
// can create a fully functional [lerian.Client] backed by simple map stores.
// All CRUD operations work correctly: Create assigns an ID, Get/Update/Delete
// operate on the stored items, and List returns paginated iterators.
//
// # Basic Usage
//
//	client := leriantest.NewFakeClient()
//
//	// Use the client exactly like a real one:
//	org, err := client.Midaz.Organizations.Create(ctx, &midaz.CreateOrganizationInput{
//	    LegalName:     "Test Corp",
//	    LegalDocument: "123",
//	})
//
//	// Retrieve it back:
//	got, err := client.Midaz.Organizations.Get(ctx, org.ID)
//
// # Seeding Data and Injecting Errors
//
// Options allow seeding data and injecting errors for specific operations:
//
//	client := leriantest.NewFakeClient(
//	    leriantest.WithSeedOrganizations(midaz.Organization{ID: "org-1", LegalName: "Acme"}),
//	    leriantest.WithErrorOn("midaz.Organizations.Create", someErr),
//	)
package leriantest
