package midaz

import "github.com/LerianStudio/lerian-sdk-golang/pkg/core"

func ensureService[T any](svc *T) error {
	if svc == nil {
		return core.ErrNilService
	}

	return nil
}
