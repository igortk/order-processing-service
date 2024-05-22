package util

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"order-processing-service/dto/proto"
	"strings"
)

func ValidateCreateOrderRequest(errors *[]*proto.Error, req *proto.CreateOrderRequest) {
	if req.Id == "" {
		*errors = append(*errors, &proto.Error{
			Code:    400,
			Message: fmt.Sprintf("invalid format request id:[%s]", req.Id),
		})
	}
	if req.UserId == "" {
		*errors = append(*errors, &proto.Error{
			Code:    400,
			Message: fmt.Sprintf("invalid format user id:[%s]", req.UserId),
		})
	}
	if req.OrderId == "" {
		*errors = append(*errors, &proto.Error{
			Code:    400,
			Message: fmt.Sprintf("invalid format order id:[%s]", req.OrderId),
		})
	}
	if req.Pair == "" {
		*errors = append(*errors, &proto.Error{
			Code:    400,
			Message: fmt.Sprintf("invalid format pair in request:[%s]", req.Pair),
		})
	}
	if req.InitVolume <= 0 {
		*errors = append(*errors, &proto.Error{
			Code:    400,
			Message: fmt.Sprintf("wrong init volume(must be >= 0.0001) in request:[%g]", req.InitVolume),
		})
	}
	if req.InitPrice <= 0 {
		*errors = append(*errors, &proto.Error{
			Code:    400,
			Message: fmt.Sprintf("wrong init price(must be >= 0.0001) in request:[%g]", req.InitPrice),
		})
	}
	if req.Direction == proto.Direction_ORDER_DIRECTION_UNDEFINED {
		*errors = append(*errors, &proto.Error{
			Code:    400,
			Message: fmt.Sprintf("invalid direction:[%s]", req.Direction),
		})
	}
}

func ValidateUserBalance(errors *[]*proto.Error, req *proto.CreateOrderRequest, balances *proto.GetBalanceByUserIdResponse) {
	var balance float64
	var currency string
	pair := strings.Split(req.Pair, "/")

	if req.Direction == proto.Direction_ORDER_DIRECTION_BUY {
		currency = pair[1]
	} else if req.Direction == proto.Direction_ORDER_DIRECTION_SELL {
		currency = pair[0]
	} else {
		*errors = append(*errors, &proto.Error{
			Code:    400,
			Message: fmt.Sprintf("unsupported direction value:[%s]", req.Direction),
		})
		return
	}

	for _, userBalance := range balances.UserBalance.Balances {
		if userBalance.Currency == currency {
			balance = userBalance.Balance
			log.Printf("currency %s", currency)
			break
		}
	}
	log.Printf("balance %s", balance)
	log.Printf("req.InitVolume*req.InitPrice %s", req.InitVolume*req.InitPrice)
	if req.InitVolume*req.InitPrice > balance {
		*errors = append(*errors, &proto.Error{
			Code:    400,
			Message: fmt.Sprintf("user balance is low"),
		})
		log.Errorf("user balance is low")
		return
	}
}
