package provider

import (
	"context"
	"github.com/severgroup-tt/gopkg-app/app"
)

type IProvider interface {
	Connect(fromAddress, fromName string, showInfo, showError bool) (ISender, app.PublicCloserFn, error)
}

type ISender interface {
	Send(ctx context.Context, msg *Message) error
}
