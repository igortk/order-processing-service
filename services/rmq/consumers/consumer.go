package consumers

import (
	gitProto "github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"order-processing-service/services/rmq/handlers"
	"order-processing-service/util"
	"time"

	"github.com/streadway/amqp"
)

type Consumer struct {
	Connection  *amqp.Connection
	Channel     *amqp.Channel
	Queue       amqp.Queue
	Handler     handlers.MessageHandler
	MessageChan chan []byte
}

func NewConsumer(connection *amqp.Connection, exchange, routingKey, queueName string, handler handlers.MessageHandler, ch chan []byte) *Consumer {
	channel, err := connection.Channel()
	util.IsError(err, "failed open channel")
	err = channel.ExchangeDeclare(
		exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	util.IsError(err, "failed exchange declare")
	queue, err := channel.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	util.IsError(err, "failed queue declare")

	err = channel.QueueBind(
		queue.Name,
		routingKey,
		exchange,
		false,
		nil,
	)
	util.IsError(err, "failed queue bind")

	return &Consumer{
		Connection:  connection,
		Channel:     channel,
		Queue:       queue,
		Handler:     handler,
		MessageChan: ch,
	}
}

func (c *Consumer) ConsumeMessages() {
	mes, err := c.Channel.Consume(
		c.Queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	util.IsError(err, "Failed to register a consumer")

	go func() {
		log.Printf(c.Queue.Name)
		for d := range mes {
			if c.Handler != nil {
				c.Handler.HandleMessage(d.Body)
			} else {
				c.MessageChan <- d.Body
			}
		}
	}()
}

func (c *Consumer) GetMessageByCondition(condition func(message []byte) bool, mes gitProto.Message, seconds time.Duration) *gitProto.Message {
	for condition(<-c.MessageChan) {
		select {
		case body := <-c.MessageChan:
			err := gitProto.Unmarshal(body, mes)
			util.IsError(err, "failed unmarshal received message")

			log.Printf("received message: %s", mes)
			return &mes

		case <-time.After(seconds * time.Second):
			log.Error("response wasn't received")
			return nil
		}
	}

	return nil
}

func (c *Consumer) Close() {
	err := c.Channel.Close()
	util.IsError(err, "failed close channel")

	err = c.Connection.Close()
	util.IsError(err, "failed close connection")
}
