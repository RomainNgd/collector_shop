package services

import (
	"context"
	"poc-gin/pkg/constants"
)

func withDBTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}

	return context.WithTimeout(parent, constants.DBTimeout)
}
