package fees

import (
	"context"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestPackagesServiceNilReceiver(t *testing.T) {
	t.Parallel()

	var svc *packagesService

	_, err := svc.Create(context.Background(), &CreatePackageInput{})
	assert.ErrorIs(t, err, core.ErrNilService)

	_, err = svc.Get(context.Background(), "pkg-001")
	assert.ErrorIs(t, err, core.ErrNilService)

	_, err = svc.List(context.Background(), nil)
	assert.ErrorIs(t, err, core.ErrNilService)

	_, err = svc.Update(context.Background(), "pkg-001", &UpdatePackageInput{})
	assert.ErrorIs(t, err, core.ErrNilService)

	err = svc.Delete(context.Background(), "pkg-001")
	assert.ErrorIs(t, err, core.ErrNilService)
}

func TestFeesServiceNilReceiver(t *testing.T) {
	t.Parallel()

	var svc *feesCalcService

	_, err := svc.Calculate(context.Background(), &FeeCalculate{})
	assert.ErrorIs(t, err, core.ErrNilService)
}

func TestEstimatesServiceNilReceiver(t *testing.T) {
	t.Parallel()

	var svc *estimatesService

	_, err := svc.Calculate(context.Background(), &FeeEstimateInput{})
	assert.ErrorIs(t, err, core.ErrNilService)
}
