package services

const (
	DefaultPageLimit = 20
	MaxPageLimit     = 100
)

// Pagination bounds a list query. Limit <= 0 falls back to DefaultPageLimit
// and is capped at MaxPageLimit; Offset < 0 is treated as 0.
type Pagination struct {
	Limit  int
	Offset int
}

func (p Pagination) normalized() (limit, offset int) {
	limit = p.Limit
	if limit <= 0 {
		limit = DefaultPageLimit
	}
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}

	offset = p.Offset
	if offset < 0 {
		offset = 0
	}

	return limit, offset
}
