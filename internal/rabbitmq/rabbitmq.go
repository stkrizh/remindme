package rabbitmq

import (
	"context"
	"fmt"
	"remindme/internal/core/domain/logging"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const delay = 3 // reconnect after delay seconds

// Connection amqp.Connection wrapper
type Connection struct {
	*amqp.Connection
	log logging.Logger
}

// Channel wrap amqp.Connection.Channel, get a auto reconnect channel
func (c *Connection) Channel() (*Channel, error) {
	ch, err := c.Connection.Channel()
	if err != nil {
		return nil, err
	}

	channel := &Channel{
		Channel: ch,
		log:     c.log,
	}

	go func() {
		for {
			reason, ok := <-channel.Channel.NotifyClose(make(chan *amqp.Error))
			// exit this goroutine if closed by developer
			if !ok || channel.IsClosed() {
				channel.Close() // close again, ensure closed flag set when connection closed
				break
			}

			c.log.Warning(context.Background(), "RabbitMQ channel closed.", logging.Entry("reason", *reason))
			for {
				time.Sleep(delay * time.Second)

				ch, err := c.Connection.Channel()
				if err == nil {
					c.log.Info(context.Background(), "Channel recreate success.")
					channel.Channel = ch
					break
				}

				c.log.Error(context.Background(), "Channel recreate failed.", logging.Entry("err", err))
			}
		}

	}()

	return channel, nil
}

func Dial(url string, log logging.Logger) (*Connection, error) {
	if log == nil {
		return nil, fmt.Errorf("log argument must not be nil")
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	connection := &Connection{
		Connection: conn,
		log:        log,
	}

	go func() {
		for {
			reason, ok := <-connection.Connection.NotifyClose(make(chan *amqp.Error))
			if !ok {
				log.Info(context.Background(), "RabbitMQ connection closed.")
				break
			}

			log.Warning(context.Background(), "RabbitMQ connection closed.", logging.Entry("reason", *reason))
			for {
				time.Sleep(delay * time.Second)

				conn, err := amqp.Dial(url)
				if err == nil {
					connection.Connection = conn
					log.Info(context.Background(), "RabbitMQ reconnect success.")
					break
				}
				log.Error(context.Background(), "RabbitMQ reconnect failed.", logging.Entry("err", err))
			}
		}
	}()

	return connection, nil
}

// Channel amqp.Channel wapper
type Channel struct {
	*amqp.Channel
	closed int32
	log    logging.Logger
}

// IsClosed indicate closed by developer
func (ch *Channel) IsClosed() bool {
	return (atomic.LoadInt32(&ch.closed) == 1)
}

// Close ensure closed flag set
func (ch *Channel) Close() error {
	if ch.IsClosed() {
		return amqp.ErrClosed
	}

	atomic.StoreInt32(&ch.closed, 1)

	return ch.Channel.Close()
}

// Consume wrap amqp.Channel.Consume, the returned delivery will end only when channel closed by developer
func (ch *Channel) Consume(
	queue, consumer string,
	autoAck, exclusive, noLocal, noWait bool,
	args amqp.Table,
) (<-chan amqp.Delivery, error) {
	deliveries := make(chan amqp.Delivery)

	go func() {
		for {
			d, err := ch.Channel.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
			if err != nil {
				ch.log.Error(context.Background(), "Consume failed.", logging.Entry("err", err))
				time.Sleep(delay * time.Second)
				continue
			}

			for msg := range d {
				deliveries <- msg
			}

			// sleep before IsClose call. closed flag may not set before sleep.
			time.Sleep(delay * time.Second)

			if ch.IsClosed() {
				ch.log.Info(context.Background(), "Channel is closed, stop consuming.", logging.Entry("queue", queue))
				break
			}
		}
	}()

	return deliveries, nil
}
