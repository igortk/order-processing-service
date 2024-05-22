package handlers

import (
	"context"
	gitProto "github.com/golang/protobuf/proto"
	"order-processing-service/config"
	"order-processing-service/dto/proto"
	"order-processing-service/services/redis"
	"order-processing-service/services/rmq/senders"
	"order-processing-service/util"
)

type DeleteOrderHandler struct {
	sender senders.Sender
	rCl    *redis.Client
}

func NewDeleteOrderHandler(sender senders.Sender, client *redis.Client) DeleteOrderHandler {
	return DeleteOrderHandler{
		sender: sender,
		rCl:    client,
	}
}

func (h DeleteOrderHandler) HandleMessage(body []byte) {
	req := &proto.RemoveOrderRequest{}
	err := gitProto.Unmarshal(body, req)
	util.IsError(err, "Failed to unmarshal message")

	orderRemoved := &proto.Order{}
	allOrders := h.rCl.GetOrdersFromRedis(context.TODO())
	for _, order := range allOrders {
		if order.OrderId == req.OrderId {
			orderRemoved = order
			break
		}
	}

	event := &proto.OrderUpdateEvent{
		Id:    req.Id,
		Order: orderRemoved,
	}

	if req.OrderId == "" {
		event.Error = append(event.Error, &proto.Error{Code: 403,
			Message: "err delete. Order id is empty"})
		h.sendEvent(event)
		return
	}

	event.Order.Status = proto.OrderStatus_ORDER_STATUS_REMOVED
	h.rCl.SetOrderData(event)

	h.sendEvent(event)
}

func (h DeleteOrderHandler) sendEvent(event *proto.OrderUpdateEvent) {
	h.sender.SendMessage(
		config.RabbitEventsExchange,
		config.UpdatedOrderEventRoutingKey,
		"topic",
		event,
	)
}
