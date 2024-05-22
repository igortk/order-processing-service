package rmq

import (
	"order-processing-service/services/rmq/consumers"
	"order-processing-service/services/rmq/handlers"
	"order-processing-service/services/rmq/senders"
	"sync"
)

type Client struct {
	sender       senders.Sender
	consumer     consumers.Consumer
	handler      handlers.MessageHandler
	deliveryChan chan []byte
}

func NewClient(sender senders.Sender, consumer consumers.Consumer, handler handlers.MessageHandler) *Client {
	return &Client{
		sender:       sender,
		consumer:     consumer,
		handler:      handler,
		deliveryChan: consumer.MessageChan,
	}
}

func (c *Client) StartHandler(wg *sync.WaitGroup) {
	defer wg.Done()

	c.handler.HandleMessage(<-c.deliveryChan)
}

func (c *Client) StartConsumer(wg *sync.WaitGroup) {
	defer wg.Done()

	c.consumer.ConsumeMessages()
}
func StartHandler(consumer *consumers.Consumer) {

}
