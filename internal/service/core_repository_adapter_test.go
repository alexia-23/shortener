package service

import "context"

// Keep generated mocks tied to the small interface they actually exercise.
type coreRepositoryAdapter struct {
	ShortLinkRepository
}

func (coreRepositoryAdapter) GetByUserID(
	context.Context,
	string,
) ([]ShortLink, error) {
	panic("unexpected GetByUserID call in a core service test")
}

func (coreRepositoryAdapter) DeleteBatch(
	context.Context,
	[]DeleteShortLink,
) error {
	panic("unexpected DeleteBatch call in a core service test")
}

var _ URLRepository = coreRepositoryAdapter{}
