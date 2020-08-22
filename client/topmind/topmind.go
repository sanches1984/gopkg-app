package topmind

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/severgroup-tt/gopkg-app/runtime"
	errors "github.com/severgroup-tt/gopkg-errors"
	logger "github.com/severgroup-tt/gopkg-logger"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const endpoint = "https://api.topmind.io/api/v1/event"

type Action string

const (
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

type Type string

const (
	TypeAnalyticsUser         Type = "AnalyticsUser"
	TypeAnalyticsUserActivity Type = "AnalyticsUserActivity"
)

const (
	ParamCreatedAt = "created_at"
	ParamActionAt  = "action_at"
)

type client struct {
	Token string
	LogSuccess bool
	RetryMaxAttempt int
	RetryInterval time.Duration
}

type request struct {
	EntityType  Type              `json:"entity_type"`
	EntityID    string            `json:"entity_id"`
	TsAction    int64             `json:"ts_action"`
	TrackerSID  string            `json:"tracker_sid,omitempty"`
	Action      Action            `json:"action"`
	DataVersion int32             `json:"data_version"`
	Data        map[string]string `json:"data,omitempty"`
}

type response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewClient(token string, option ...Option) IClient {
	if token == "" {
		return &noop{}
	}
	c := &client{
		Token: token,
	}
	for _, f := range option {
		f(c)
	}
	logger.Info(logger.App, "Topmind: connected with token %s%s", strings.Repeat("*", len(token)-3), token[len(token)-3:])
	return c
}

func (c client) getTimeAction(dt time.Time) int64 {
	return dt.Unix() * 1000
}

func (c client) getTimeData(dt time.Time) string {
	return dt.Format(time.RFC3339)
}

func (c client) Create(ctx context.Context, entityType Type, entityID string, dataVersion int32, data map[string]string) {
	c.asyncSend(ctx, request{
		EntityType:  entityType,
		EntityID:    entityID,
		TsAction:    c.getTimeAction(runtime.Now()),
		Action:      ActionCreate,
		DataVersion: dataVersion,
		Data:        data,
	})
}

func (c client) Update(ctx context.Context, entityType Type, entityID string, dataVersion int32, data map[string]string) {
	c.asyncSend(ctx, request{
		EntityType:  entityType,
		EntityID:    entityID,
		TsAction:    c.getTimeAction(runtime.Now()),
		Action:      ActionUpdate,
		DataVersion: dataVersion,
		Data:        data,
	})
}

func (c client) Delete(ctx context.Context, entityType Type, entityID string, dataVersion int32, data map[string]string) {
	c.asyncSend(ctx, request{
		EntityType:  entityType,
		EntityID:    entityID,
		TsAction:    c.getTimeAction(runtime.Now()),
		Action:      ActionDelete,
		DataVersion: dataVersion,
		Data:        data,
	})
}

// TypeAnalyticsUser

func (c client) fillUserData(data map[string]string, dt time.Time) {
	if data == nil {
		data = make(map[string]string, 1)
	}
	data[ParamCreatedAt] = c.getTimeData(dt)
}

func (c client) CreateUser(ctx context.Context, userID int64, dataVersion int32, data map[string]string) {
	now := runtime.Now()
	c.fillUserData(data, now)
	c.asyncSend(ctx, request{
		EntityType:  TypeAnalyticsUser,
		EntityID:    strconv.FormatInt(userID, 10),
		TsAction:    c.getTimeAction(now),
		Action:      ActionCreate,
		DataVersion: dataVersion,
		Data:        data,
	})
}

func (c client) UpdateUser(ctx context.Context, userID int64, dataVersion int32, data map[string]string) {
	now := runtime.Now()
	c.fillUserData(data, now)
	c.asyncSend(ctx, request{
		EntityType:  TypeAnalyticsUser,
		EntityID:    strconv.FormatInt(userID, 10),
		TsAction:    c.getTimeAction(now),
		Action:      ActionUpdate,
		DataVersion: dataVersion,
		Data:        data,
	})
}

func (c client) DeleteUser(ctx context.Context, userID int64, dataVersion int32, data map[string]string) {
	now := runtime.Now()
	c.fillUserData(data, now)
	c.asyncSend(ctx, request{
		EntityType:  TypeAnalyticsUser,
		EntityID:    strconv.FormatInt(userID, 10),
		TsAction:    c.getTimeAction(now),
		Action:      ActionDelete,
		DataVersion: dataVersion,
		Data:        data,
	})
}

// TypeAnalyticsUserActivity

func (c client) fillUserActivityData(data map[string]string, dt time.Time) {
	if data == nil {
		data = make(map[string]string, 1)
	}
	data[ParamActionAt] = c.getTimeData(dt)
}

func (c client) CreateUserActivity(ctx context.Context, userID int64, dataVersion int32, data map[string]string) {
	now := runtime.Now()
	c.fillUserActivityData(data, now)
	c.asyncSend(ctx, request{
		EntityType:  TypeAnalyticsUserActivity,
		EntityID:    strconv.FormatInt(userID, 10),
		TsAction:    c.getTimeAction(now),
		Action:      ActionCreate,
		DataVersion: dataVersion,
		Data:        data,
	})
}

func (c client) UpdateUserActivity(ctx context.Context, userID int64, dataVersion int32, data map[string]string) {
	now := runtime.Now()
	c.fillUserActivityData(data, now)
	c.asyncSend(ctx, request{
		EntityType:  TypeAnalyticsUserActivity,
		EntityID:    strconv.FormatInt(userID, 10),
		TsAction:    c.getTimeAction(now),
		Action:      ActionUpdate,
		DataVersion: dataVersion,
		Data:        data,
	})
}

func (c client) DeleteUserActivity(ctx context.Context, userID int64, dataVersion int32, data map[string]string) {
	now := runtime.Now()
	c.fillUserActivityData(data, now)
	c.asyncSend(ctx, request{
		EntityType:  TypeAnalyticsUserActivity,
		EntityID:    strconv.FormatInt(userID, 10),
		TsAction:    c.getTimeAction(now),
		Action:      ActionDelete,
		DataVersion: dataVersion,
		Data:        data,
	})
}

func (c client) asyncSend(ctx context.Context, req request) {
	go func(ctx context.Context, req request) {
		if err := c.send(ctx, req); err != nil {
			logger.Error(ctx, "Error on send topmind message: %+v", err)
		}
	}(ctx, req)
}

func (c client) send(ctx context.Context, req request) error {
	bts, _ := json.Marshal(&req)
	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(bts))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.Token)

	client := &http.Client{}

	var httpResp *http.Response
	var httpErr error
	attempt := c.RetryMaxAttempt
	for {
		httpResp, httpErr = client.Do(httpReq)
		if httpErr == nil || attempt <= 1 {
			break
		}
		attempt--
		time.Sleep(c.RetryInterval)
	}
	if httpErr != nil {
		return errors.Internal.ErrWrap(ctx, "Can't send topmind request", httpErr).
			WithLogKV("request", req)
	}

	defer httpResp.Body.Close()

	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return errors.Internal.Err(ctx, "Can't read topmind response").WithLogKV("body", body)
	}

	var resp response
	if err = json.Unmarshal(body, &resp); err != nil {
		return errors.Internal.Err(ctx, "Can't unmarshal topmind response").WithLogKV("body", body)
	}

	if resp.Status != "success" {
		return errors.Internal.Err(ctx, "Error in topmind response").WithLogKV("message", resp.Message)
	}

	if c.LogSuccess {
		logger.Info(ctx, "Send topmind event %s %s:%s: v%d %v",
			req.Action, req.EntityType, req.EntityID, req.DataVersion, req.Data)
	}

	return nil
}
