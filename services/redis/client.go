package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"order-processing-service/dto/proto"
	"order-processing-service/util"
	"sort"
	"strconv"
	"time"
)

type Client struct {
	client *redis.Client
}

func NewRedisClient(addr, password string, db int) *Client {
	return &Client{
		client: redis.NewClient(&redis.Options{
			Addr: "localhost:7000",
			DB:   0,
		}),
	}
}

//func NewRedisClient(addr, password string, db int) *Client {
//	return &Client{
//		client: redis.NewClusterClient(&redis.ClusterOptions{
//			Addrs: []string{
//				"node1:6379", // Add the addresses of your Redis nodes here
//				"node2:6379",
//				// Add more nodes as needed
//			},
//			Password: password,
//		}),
//	}
//}

func (rc *Client) SetOrderData(event *proto.OrderUpdateEvent) {

	err := rc.client.SAdd(context.TODO(), "orders", event.Order.OrderId).Err()
	util.IsError(err, "error adding to set")

	err = rc.client.HSet(context.TODO(), fmt.Sprintf("order:%s", event.Order.OrderId), map[string]interface{}{
		"user_id":      event.Order.UserId,
		"pair":         event.Order.Pair,
		"init_volume":  event.Order.InitVolume,
		"fill_volume":  event.Order.FillVolume,
		"init_price":   event.Order.InitPrice,
		"status":       int(event.Order.Status.Number()),
		"direction":    int(event.Order.Direction.Number()),
		"updated_date": time.Now().Unix(),
		"created_date": event.Order.CreatedDate,
	}).Err()

	util.IsError(err, "err set data to redis")
}

func (rc *Client) GetOrderData(orderId string) map[string]string {
	userData, err := rc.client.HGetAll(context.TODO(), fmt.Sprintf("order:%s", orderId)).Result()

	util.IsError(err, "err get data from redis")

	return userData
}

func (rc *Client) GetOrdersFromRedis(ctx context.Context) []*proto.Order {
	userIDs, err := rc.client.SMembers(ctx, "orders").Result()
	util.IsError(err, "failed get all orders")

	log.Println(userIDs)
	var orders []*proto.Order
	for _, orderId := range userIDs {
		log.Infof("sercher +%s", orderId)
		orderData, err := rc.client.HGetAll(ctx, fmt.Sprintf("order:%s", orderId)).Result()
		util.IsError(err, fmt.Sprintf("failed get order [id: %s]", orderId))

		order := &proto.Order{
			OrderId:     orderId,
			UserId:      orderData["user_id"],
			Pair:        orderData["pair"],
			InitVolume:  convertToFloat(orderData["init_volume"]),
			FillVolume:  convertToFloat(orderData["fill_volume"]),
			InitPrice:   convertToFloat(orderData["init_price"]),
			Status:      proto.OrderStatus(convertToInt64(orderData["status"])),
			Direction:   proto.Direction(convertToInt64(orderData["direction"])),
			UpdatedDate: convertToInt64(orderData["updated_date"]),
			CreatedDate: convertToInt64(orderData["created_date"]),
		}
		orders = append(orders, order)

	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedDate < orders[j].CreatedDate
	})

	return orders
}
func convertToFloat(value string) float64 {
	result, err := strconv.ParseFloat(value, 64)
	util.IsError(err, "failed pars to float")
	return result
}

func convertToInt64(value string) int64 {
	result, err := strconv.ParseInt(value, 10, 64)
	util.IsError(err, "failed pars to int")
	return result
}
