package provider

import (
	"context"
	"github.com/severgroup-tt/gopkg-app/app"
)

type IProvider interface {
	Connect(showInfo, showError bool) (ISender, app.PublicCloserFn, error)
}

type ISender interface {
	Send(ctx context.Context, phone int64, message string) error
}
