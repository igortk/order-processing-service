package senders

import (
	"fmt"
	gitProto "github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"order-processing-service/util"
	"sync"
)

type Sender struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

func NewSender(connection *amqp.Connection) Sender {
	channel, err := connection.Channel()
	util.IsError(err, "err create chanel for senders")

	return Sender{
		Connection: connection,
		Channel:    channel,
	}
}

func (s Sender) SendMessage(exchange, routingKey, kind string, mes gitProto.Message) {
	message, err := gitProto.Marshal(mes)
	util.IsError(err, fmt.Sprintf("failed Marshal message[%s] ", mes))

	err = s.Channel.ExchangeDeclare(
		exchange,
		kind,
		true,
		false,
		false,
		false,
		nil,
	)
	util.IsError(err, fmt.Sprintf("failed exchange[%s] declare ", exchange))

	err = s.Channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			Body: message,
		},
	)
	util.IsError(err, "failed send message")
}

func (s Sender) SendMessageAndWaitForResponse(exchange, routingKey, requestId string, message []byte) []byte {

	responseQueue, err := s.Channel.QueueDeclare(
		"q.w.e.r.t.y.u.i.o.p.a", // name
		false,                   // durable
		true,                    // delete when unused
		true,                    // exclusive
		false,                   // no-wait
		nil,                     // arguments
	)
	util.IsError(err, "err1")

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)

	// Определение консьюмера для ожидания ответа
	msgs, err := s.Channel.Consume(
		responseQueue.Name, // queue
		"",                 // consumers
		true,               // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	util.IsError(err, "err2")

	// Публикация сообщения с уникальным correlation ID
	err = s.Channel.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          message,
			CorrelationId: requestId,
			ReplyTo:       responseQueue.Name,
		},
	)
	util.IsError(err, "err3")
	responseChan := make(chan []byte, 1)
	// Ожидание ответа
	go func() {
		for d := range msgs {
			if d.CorrelationId == requestId {
				response := d.Body
				log.Printf("Received response: %s\n", response)
				// Отправляем ответ в канал
				responseChan <- response
				// Закрываем канал, чтобы завершить горутину после получения ответа
				close(responseChan)
				return
			}
		}
	}()

	// Ожидание завершения получения ответа
	response, ok := <-responseChan
	if !ok {
		return nil
	}

	// Закрытие временной очереди
	_, err = s.Channel.QueueDelete(responseQueue.Name, false, false, false)
	if err != nil {
		log.Printf("Failed to delete temporary queue: %v", err)
	}

	return response
}

func (s Sender) Close() {
	s.Channel.Close()
	s.Connection.Close()
}
