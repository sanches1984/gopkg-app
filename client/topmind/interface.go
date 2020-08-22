package topmind

import "context"

type IClient interface {
	Create(ctx context.Context, entityType Type, entityID string, dataVersion int32, data map[string]string)
	Update(ctx context.Context, entityType Type, entityID string, dataVersion int32, data map[string]string)
	Delete(ctx context.Context, entityType Type, entityID string, dataVersion int32, data map[string]string)
	CreateUser(ctx context.Context, userID int64, dataVersion int32, data map[string]string)
	UpdateUser(ctx context.Context, userID int64, dataVersion int32, data map[string]string)
	DeleteUser(ctx context.Context, userID int64, dataVersion int32, data map[string]string)
	CreateUserActivity(ctx context.Context, userID int64, dataVersion int32, data map[string]string)
	UpdateUserActivity(ctx context.Context, userID int64, dataVersion int32, data map[string]string)
	DeleteUserActivity(ctx context.Context, userID int64, dataVersion int32, data map[string]string)
}
