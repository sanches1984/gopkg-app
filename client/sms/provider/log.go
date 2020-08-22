package provider

import (
	"context"
	"github.com/severgroup-tt/gopkg-app/app"
	logger "github.com/severgroup-tt/gopkg-logger"
)

type logProvider struct{}

func NewLogProvider() IProvider {
	return &logProvider{}
}

func (c logProvider) Connect(_, _ bool) (ISender, app.PublicCloserFn, error) {
	return c, nil, nil
}

func (c logProvider) Send(ctx context.Context, phone int64, message string) error {
	logger.Info(ctx, "Send SMS '%s' to %v", message, phone)
	return nil
}
