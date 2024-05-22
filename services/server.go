package services

import (
	"fmt"
	"github.com/streadway/amqp"
	"order-processing-service/config"
	"order-processing-service/services/redis"
	"order-processing-service/services/rmq/consumers"
	"order-processing-service/services/rmq/handlers"
	"order-processing-service/services/rmq/senders"
	"order-processing-service/util"
)

type Server struct {
	clRedis   *redis.Client
	rmqConn   *amqp.Connection
	senders   senders.Sender
	handlers  map[string]handlers.MessageHandler
	consumers map[string]*consumers.Consumer
	ch        chan []byte
}

func NewServer(cfg *config.Config) *Server {
	conn, err := amqp.Dial(fmt.Sprintf(config.RmqUrlConnectionPattern,
		cfg.RabbitConfig.Username,
		cfg.RabbitConfig.Password,
		cfg.RabbitConfig.Host,
		cfg.RabbitConfig.Port,
	))
	util.IsError(err, "err")

	rCl := redis.NewRedisClient(
		fmt.Sprintf(config.RedisConnectionPattern, cfg.RedisConfig.Host, cfg.RedisConfig.Port),
		cfg.RedisConfig.Password,
		8,
	)
	/*f := rCl.GetOrdersFromRedis(context.TODO())
	log.Info(f)*/
	return &Server{
		clRedis:   rCl,
		rmqConn:   conn,
		senders:   senders.NewSender(conn),
		handlers:  map[string]handlers.MessageHandler{},
		consumers: map[string]*consumers.Consumer{},
		ch:        make(chan []byte),
	}
}

func (s *Server) InitHandlers() {
	s.handlers["CreateOrderHandler"] = handlers.NewCreateOrderHandler(s.senders, s.clRedis, s.ch)
	s.handlers["GetUserOrdersHandler"] = handlers.NewGetUserOrdersHandler(s.senders, s.clRedis)
	s.handlers["DeleteOrderHandler"] = handlers.NewDeleteOrderHandler(s.senders, s.clRedis)
}

func (s *Server) InitConsumers() {
	s.consumers["CreateOrderRequestConsumer"] = consumers.NewConsumer(
		s.rmqConn,
		config.RabbitOrderExchange,
		config.CreateOrderRequestRoutingKey,
		config.CreateOrderRequestQueueName,
		s.handlers["CreateOrderHandler"],
		nil)

	s.consumers["GetBalanceByUserIdResponse"] = consumers.NewConsumer(
		s.rmqConn,
		config.RabbitBalanceExchange,
		config.GetBalanceByUserIdResponseRoutingKey,
		config.BalanceInfoResponseQueueName,
		nil,
		s.ch)

	s.consumers["GetUserOrdersRequestConsumer"] = consumers.NewConsumer(
		s.rmqConn,
		config.RabbitOrderExchange,
		config.GetUserOrdersRequestRoutingKey,
		config.UserOrdersResponseQueueName,
		s.handlers["GetUserOrdersHandler"],
		nil)

	s.consumers["RemoveOrderRequestConsumer"] = consumers.NewConsumer(
		s.rmqConn,
		config.RabbitOrderExchange,
		config.RemoveUserOrderRequestRoutingKey,
		config.RemoveOrderResponseQueueName,
		s.handlers["DeleteOrderHandler"],
		nil)
}

func (s *Server) Run() {
	forever := make(chan bool)
	s.runAllConsumers()
	<-forever
}

func (s *Server) runAllConsumers() {
	for _, client := range s.consumers {
		go client.ConsumeMessages()
	}
}
