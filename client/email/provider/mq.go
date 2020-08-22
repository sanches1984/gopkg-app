package provider

import (
	"context"
	"encoding/json"
	"github.com/severgroup-tt/gopkg-app/middleware"
	"os"

	"github.com/severgroup-tt/gopkg-app/app"
	errors "github.com/severgroup-tt/gopkg-errors"
	logger "github.com/severgroup-tt/gopkg-logger"
	"github.com/streadway/amqp"
)

type mqProvider struct {
	dsn        string
	channel    *amqp.Channel
	from       string
	exchange   string
	routingKey string
	showInfo   bool
	showError  bool
}

type mqData struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Content string `json:"content"`
}

func NewCoreMqProvider() IProvider {
	dsn, ok := os.LookupEnv("RABBITMQ_ADDRESS")
	if !ok {
		panic("Core services RabbitMQ not fount in env RABBITMQ_ADDRESS")
	}
	return NewMqProvider(dsn, "notifications", "send_email")
}

func NewMqProvider(dsn, exchange, routingKey string) IProvider {
	return &mqProvider{
		dsn:        dsn,
		exchange:   exchange,
		routingKey: routingKey,
	}
}

func (c mqProvider) Connect(fromAddress, fromName string, showInfo, showError bool) (ISender, app.PublicCloserFn, error) {
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
	c.from = Contact{Address: fromAddress, Name: fromName}.String()
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

func (c mqProvider) Send(ctx context.Context, msg *Message) error {
	body, err := json.Marshal(mqData{
		From:    c.from,
		To:      msg.To.String(),
		Subject: msg.Subject,
		Content: msg.bodyHTML,
	})
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
