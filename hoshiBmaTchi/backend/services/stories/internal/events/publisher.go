package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type EventPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewEventPublisher(url string) (*EventPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		"notifications_exchange", 
		"topic",                 
		true,                     
		false,                   
		false,                    
		false,                    
		nil,                      
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &EventPublisher{conn: conn, channel: ch}, nil
}

func (p *EventPublisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}

type NotificationEvent struct {
	Type      string    `json:"type"`       
	ActorID   string    `json:"actor_id"`   
	TargetID  string    `json:"target_id"`  
	ResourceID string   `json:"resource_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func (p *EventPublisher) PublishNotification(ctx context.Context, event NotificationEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	routingKey := "notification.story" 

	return p.channel.PublishWithContext(ctx,
		"notifications_exchange",
		routingKey,               
		false,                    
		false,                    
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)
}