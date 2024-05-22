package handlers

import (
	"context"
	gitProto "github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"order-processing-service/config"
	"order-processing-service/dto/proto"
	"order-processing-service/services/redis"
	"order-processing-service/services/rmq/senders"
	"order-processing-service/util"
	"time"
)

type CreateOrderHandler struct {
	sender   senders.Sender
	rCl      *redis.Client
	respChan chan []byte
}

func NewCreateOrderHandler(sender senders.Sender, client *redis.Client, respChan chan []byte) CreateOrderHandler {
	return CreateOrderHandler{
		sender:   sender,
		rCl:      client,
		respChan: respChan,
	}
}

func (h CreateOrderHandler) HandleMessage(body []byte) {
	req := &proto.CreateOrderRequest{}
	err := gitProto.Unmarshal(body, req)
	util.IsError(err, "Failed to unmarshal message")

	log.Info("Received CreateOrderRequest: %s\n", req)
	var errors []*proto.Error
	util.ValidateCreateOrderRequest(&errors, req)

	respBalance := h.getBalanceByUserId(req.UserId)
	log.Info(respBalance)
	util.ValidateUserBalance(&errors, req, respBalance)

	newOrderEvent := h.createFirstEvent(req, errors)

	if len(errors) != 0 {
		h.sendEvent(newOrderEvent)
		log.Error("request was not validated")
		return
	}

	h.rCl.SetOrderData(newOrderEvent)

	//TODO to complete the matching mechanism
	orders := h.rCl.GetOrdersFromRedis(context.Background())

	for _, order := range orders {
		if order.Pair == req.Pair && order.Direction != req.Direction && order.UserId != req.UserId {

			remainderOrderNew := newOrderEvent.Order.InitVolume - newOrderEvent.Order.FillVolume
			remainderOrderOld := order.InitVolume - order.FillVolume

			if remainderOrderNew > remainderOrderOld {
				newOrderEvent.Order.FillVolume += remainderOrderOld
				order.FillVolume += remainderOrderOld
			} else {
				newOrderEvent.Order.FillVolume += remainderOrderNew
				order.FillVolume += remainderOrderNew
			}
			if order.FillVolume == order.InitVolume {
				order.Status = proto.OrderStatus_ORDER_STATUS_MATCHED
			}
			oldOrderEvent := &proto.OrderUpdateEvent{
				Id:    uuid.NewString(),
				Order: order,
			}

			h.sendEvent(oldOrderEvent)
			h.rCl.SetOrderData(oldOrderEvent)

			if newOrderEvent.Order.InitVolume == newOrderEvent.Order.FillVolume {
				newOrderEvent.Order.Status = proto.OrderStatus_ORDER_STATUS_MATCHED
				break
			}
		}
	}

	h.rCl.SetOrderData(newOrderEvent)
	h.sendEvent(newOrderEvent)
}

func (h CreateOrderHandler) getBalanceByUserId(id string) *proto.GetBalanceByUserIdResponse {
	reqBal := &proto.GetBalanceByUserIdRequest{
		Id:     uuid.NewString(),
		UserId: id,
	}

	var respBalance proto.GetBalanceByUserIdResponse
	h.sender.SendMessage(config.RabbitBalanceExchange,
		config.GetBalanceByUserIdRequestRoutingKey,
		"topic",
		reqBal)

	for respBalance.Id != reqBal.Id {
		select {
		case body := <-h.respChan:
			err := gitProto.Unmarshal(body, &respBalance)
			util.IsError(err, "failed unmarshal message")
			log.Printf("resp: %s", respBalance.Id)
			return &respBalance
		case <-time.After(10 * time.Second):
			log.Error("response wasn't received")
			return nil
		}
	}

	return nil
}

func (h CreateOrderHandler) createFirstEvent(req *proto.CreateOrderRequest, errors []*proto.Error) *proto.OrderUpdateEvent {
	return &proto.OrderUpdateEvent{
		Id: req.OrderId,
		Order: &proto.Order{
			OrderId:     req.OrderId,
			UserId:      req.UserId,
			Pair:        req.Pair,
			InitVolume:  req.InitVolume,
			FillVolume:  0.0,
			InitPrice:   req.InitPrice,
			Status:      proto.OrderStatus_ORDER_STATUS_ACTIVE,
			Direction:   req.Direction,
			UpdatedDate: time.Now().Unix(),
			CreatedDate: time.Now().Unix(),
		},
		Error: errors,
	}
}

func (h CreateOrderHandler) sendEvent(event *proto.OrderUpdateEvent) {
	h.sender.SendMessage(
		config.RabbitEventsExchange,
		config.UpdatedOrderEventRoutingKey,
		"topic",
		event,
	)
}
