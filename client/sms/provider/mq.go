package provider

import (
	"context"
	"encoding/json"
	"os"
	"strconv"

	"github.com/severgroup-tt/gopkg-app/app"
	"github.com/severgroup-tt/gopkg-app/middleware"
	errors "github.com/severgroup-tt/gopkg-errors"
	logger "github.com/severgroup-tt/gopkg-logger"
	"github.com/streadway/amqp"
)

type mqProvider struct {
	channel    *amqp.Channel
	dsn        string
	exchange   string
	routingKey string
	showInfo   bool
	showError  bool
}

type mqData struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

func NewCoreMqProvider() IProvider {
	dsn, ok := os.LookupEnv("RABBITMQ_ADDRESS")
	if !ok {
		panic("Core services RabbitMQ not fount in env RABBITMQ_ADDRESS")
	}
	return NewMqProvider(dsn, "notifications", "send_sms")
}

func NewMqProvider(dsn, exchange, routingKey string) IProvider {
	return &mqProvider{
		dsn:        dsn,
		exchange:   exchange,
		routingKey: routingKey,
	}
}

func (c mqProvider) Connect(showInfo, showError bool) (ISender, app.PublicCloserFn, error) {
	conn, err := amqp.Dial(c.dsn)
	if err != nil {
		return nil, func() error {
				return nil
			},
			errors.Internal.ErrWrap(context.Background(), "sms: error on AMQP connect", err).WithLogKV("dsn", c.dsn)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, func() error {
				logger.Info(logger.App, "sms: close AMQP connect")
				return conn.Close()
			},
			errors.Internal.ErrWrap(context.Background(), "sms: error on amqp channel connect", err).WithLogKV("dsn", c.dsn)
	}

	logger.Info(logger.App, "sms: use AMPQ connect %v", c.dsn)

	c.channel = channel
	c.showInfo = showInfo
	c.showError = showError
	closer := func() error {
		logger.Info(logger.App, "sms: close AMQP connect")
		if err := conn.Close(); err != nil {
			return err
		}
		return channel.Close()
	}
	return c, closer, nil
}

func (c mqProvider) Send(ctx context.Context, phone int64, message string) error {
	body, err := json.Marshal(mqData{Phone: strconv.FormatInt(phone, 10), Message: message})
	if err != nil {
		return err
	}
	err = c.channel.Publish(
		c.exchange,
		c.routingKey,
		true,
		false,
		amqp.Publishing{
			MessageId:   middleware.GetRequestId(ctx),
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		if c.showError {
			logger.Error(ctx, "Error on publish message %s to %v.%v: %v", body, c.exchange, c.routingKey, err)
		}
		return err
	}
	if c.showInfo {
		logger.Info(ctx, "Success publish message %s to %v.%v", body, c.exchange, c.routingKey)
	}

	return nil
}
