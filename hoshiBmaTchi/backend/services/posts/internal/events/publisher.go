package events

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type NotificationEvent struct {
	RecipientID uint   `json:"recipient_id"`
	SenderID    uint   `json:"sender_id"`
	SenderName  string `json:"sender_name"`
	SenderImage string `json:"sender_image"`
	Type        string `json:"type"`      
	EntityID    uint   `json:"entity_id"` 
	Message     string `json:"message"`
}

type RabbitMQPublisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbitMQPublisher(url string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare("notification_exchange", "topic", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return &RabbitMQPublisher{conn: conn, ch: ch}, nil
}

func (p *RabbitMQPublisher) Publish(event NotificationEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	routingKey := "notification." + event.Type
	
	err = p.ch.Publish(
		"notification_exchange",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("Failed to publish event: %v", err)
		return err
	}
	log.Printf("Published event: %s", routingKey)
	return nil
}

func (p *RabbitMQPublisher) Close() {
	p.ch.Close()
	p.conn.Close()
}