package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPageLimit = 20
	maxPageLimit     = 100
)

// paginationParams reads optional limit/offset query params. Missing or
// invalid values fall back to sane defaults instead of erroring, since these
// endpoints predate pagination and existing clients never send them.
func paginationParams(c *gin.Context) (limit, offset int) {
	limit = defaultPageLimit
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > maxPageLimit {
		limit = maxPageLimit
	}

	offset = 0
	if raw := c.Query("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}

// setPaginationHeaders exposes the total item count and the page window via
// headers so the JSON body keeps returning a plain array and existing
// clients that only read `data` as a list keep working unmodified.
func setPaginationHeaders(c *gin.Context, total int64, limit, offset int) {
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.Header("X-Limit", strconv.Itoa(limit))
	c.Header("X-Offset", strconv.Itoa(offset))
}
