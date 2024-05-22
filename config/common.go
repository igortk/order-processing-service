package config

// RMQ exchange
const (
	RabbitEventsExchange  = "e.events.forward"
	RabbitOrderExchange   = "e.orders.forward"
	RabbitBalanceExchange = "e.balances.forward"
)

// RMQ rk
const (
	CreateOrderRequestRoutingKey         = "r.order-processing-service.order.#.CreateOrderRequest"
	GetBalanceByUserIdRequestRoutingKey  = "r.balance-service.balances.order-processing-service.GetBalanceByUserIdRequest"
	UpdatedOrderEventRoutingKey          = "r.event.order.OrderUpdateEvent"
	GetBalanceByUserIdResponseRoutingKey = "r.balance.GetBalanceByUserIdResponse"
	GetUserOrdersRequestRoutingKey       = "r.order-processing-service.order.#.GetUserOrdersRequest"
	RemoveUserOrderRequestRoutingKey     = "r.order-processing-service.order.#.RemoveOrderRequest"
	GetUserOrdersResponseRoutingKey      = "r.order-processing-service.order.GetUserOrdersResponse"
)

// Queue name
const (
	UpdatedOrderEventQueueName   = "q.balance-service.order.event"
	CreateOrderRequestQueueName  = "q.order-processing-service.order.create"
	BalanceInfoResponseQueueName = "q.order-processing-service.balance.info"
	UserOrdersResponseQueueName  = "q.order-processing-service.user.orders.info"
	RemoveOrderResponseQueueName = "q.order-processing-service.user.orders.delete"
)

// Errors
const (
	ErrLoadConfig = "Error load configuration"
	ErrParseLog   = "Error parse log level"
	ErrConnectDb  = "Error connect db"
)

// Url pattern
const (
	RmqUrlConnectionPattern = "amqp://%s:%s@%s:%d/"
	RedisConnectionPattern  = "%s:%d"
)
