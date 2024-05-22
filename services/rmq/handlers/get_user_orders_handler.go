package handlers

import (
	"context"
	gitProto "github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"order-processing-service/config"
	"order-processing-service/dto/proto"
	"order-processing-service/services/redis"
	"order-processing-service/services/rmq/senders"
	"order-processing-service/util"
)

type GetUserOrdersHandler struct {
	sender senders.Sender
	rCl    *redis.Client
}

func NewGetUserOrdersHandler(sender senders.Sender, client *redis.Client) GetUserOrdersHandler {
	return GetUserOrdersHandler{
		sender: sender,
		rCl:    client,
	}
}

func (h GetUserOrdersHandler) HandleMessage(body []byte) {
	req := &proto.GetUserOrdersRequest{}
	err := gitProto.Unmarshal(body, req)
	util.IsError(err, "Failed to unmarshal message")

	log.Info("Received GetUserOrdersRequest: %s\n", req)

	orders := []*proto.Order{}
	allOrders := h.rCl.GetOrdersFromRedis(context.TODO())
	for _, order := range allOrders {
		if order.UserId == req.UserId {
			orders = append(orders, order)
		}
	}

	resp := &proto.GetUserOrdersResponse{
		Id:     req.Id,
		Orders: orders,
		Error:  nil,
	}
	//resp := &proto.GetUserOrdersResponse{
	//	Id: req.Id,
	//	Orders: []*proto.Order{
	//		{
	//			OrderId: "test",
	//		},
	//	},
	//	Error: &proto.Error{
	//		Message: "test",
	//	},
	//}
	//message, err := gitProto.Marshal(resp)
	//
	//req2 := &proto.GetUserOrdersResponse{}
	//err = gitProto.Unmarshal(message, req2)

	h.sender.SendMessage(
		config.RabbitOrderExchange,
		config.GetUserOrdersResponseRoutingKey,
		"topic",
		resp,
	)
}
