package provider

import (
	"context"
	"github.com/sanches1984/gopkg-app/app"
)

type IProvider interface {
	Connect(fromAddress, fromName string, showInfo, showError bool) (ISender, app.PublicCloserFn, error)
}

type ISender interface {
	Send(ctx context.Context, msg *Message) error
}
