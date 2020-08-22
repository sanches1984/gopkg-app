package topmind

import (
	"context"
)

type noop struct{}

func (c noop) Create(ctx context.Context, entityType Type, entityID string, dataVersion int32, data map[string]string) {
}

func (c noop) Update(ctx context.Context, entityType Type, entityID string, dataVersion int32, data map[string]string) {
}

func (c noop) Delete(ctx context.Context, entityType Type, entityID string, dataVersion int32, data map[string]string) {
}

func (c noop) CreateUser(ctx context.Context, userID int64, dataVersion int32, data map[string]string) {
}

func (c noop) UpdateUser(ctx context.Context, userID int64, dataVersion int32, data map[string]string) {
}

func (c noop) DeleteUser(ctx context.Context, userID int64, dataVersion int32, data map[string]string) {
}

func (c noop) CreateUserActivity(ctx context.Context, userID int64, dataVersion int32, data map[string]string) {
}

func (c noop) UpdateUserActivity(ctx context.Context, userID int64, dataVersion int32, data map[string]string) {
}

func (c noop) DeleteUserActivity(ctx context.Context, userID int64, dataVersion int32, data map[string]string) {
}
