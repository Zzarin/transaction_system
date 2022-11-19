package rabbitMQ

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Zzarin/transaction_system/internal"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"strconv"
)

type Distributor struct {
	Conn *amqp.Connection
	Ch   *amqp.Channel
}

func GetNewDistributor(urlConnection string) (*Distributor, error) {
	conn, err := amqp.Dial(urlConnection)
	if err != nil {
		return nil, fmt.Errorf("rabbitMQ connection %v", zap.Error(err))
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("rabbitMQ open channel %v", zap.Error(err))
	}

	err = ch.ExchangeDeclare(
		"clientsTransactions",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	return &Distributor{Conn: conn, Ch: ch}, nil
}

func (d *Distributor) SendMessage(ctx context.Context, domainStruct *internal.IncStruct) error {
	ctxSendMessage, cancel := context.WithCancel(ctx)
	defer cancel()
	messageInBytes, err := json.Marshal(domainStruct)

	err = d.Ch.PublishWithContext(ctxSendMessage,
		"clientsTransactions",               // exchange name
		strconv.Itoa(domainStruct.ClientID), // routing key
		false,                               // mandatory
		false,                               // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         messageInBytes,
		})
	if err != nil {
		return fmt.Errorf("sending new message with key:%d, %v", domainStruct.ClientID, zap.Error(err))
	}
	return nil
}

func (d *Distributor) ReadMessage(ctx context.Context, bindingKey string, finishedTask chan bool) (chan internal.IncStruct, error) {
	ctxReadMessage, cancel := context.WithCancel(ctx)
	defer cancel()
	q, err := d.Ch.QueueDeclare(
		"",    // name
		true,  // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("declaring new queue for bindingKey:%s, %v", bindingKey, zap.Error(err))
	}

	err = d.Ch.QueueBind(
		q.Name,                // queue name
		bindingKey,            // binding bindingKey
		"clientsTransactions", // exchange
		false,
		nil)
	if err != nil {
		return nil, fmt.Errorf("binding new queue with exchange, bindingKey:%s, %v", bindingKey, zap.Error(err))
	}

	err = d.Ch.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("messaging distribution control, bindingKey:%s, %v", bindingKey, zap.Error(err))
	}

	msgs, err := d.Ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("consume messages for key:%s, %v", bindingKey, zap.Error(err))
	}

	chMessag := make(chan internal.IncStruct)
	go func() error {
		for msg := range msgs {
			var domainStruct internal.IncStruct
			err := json.Unmarshal(msg.Body, domainStruct)
			if err != nil {
				return fmt.Errorf("decoding message for key:%s, %v", bindingKey, zap.Error(err))
			}
			chMessag <- domainStruct
			if <-finishedTask == true {
				err := msg.Ack(false)
				return fmt.Errorf("acknowledging task with key:%s, %v", bindingKey, zap.Error(err))
			}

		}
		close(chMessag)
		ctxReadMessage.Done()
		return nil
	}()

	return chMessag, nil
}
