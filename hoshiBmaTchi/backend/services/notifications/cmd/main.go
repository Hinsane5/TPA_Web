package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string){
	if err != nil{
		log.Fatalf("%s: %v", msg, err)
	}
}

func main(){
	conn, err := amqp.Dial("amqp://admin:rabbitmq_password_123@rabbitmq:5672/")

	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"email_exchange",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"email_queue",
		true,
		false,
		false,
		false,
		nil,
	)

	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,
		"send_email",
		"email_exchange",
		false,
		nil,
	)

	failOnError(err, "Failed to bind a queue")

	welcomeQueue, err := ch.QueueDeclare(
		"welcome_email_queue",
		true,
		false,
		false,
		false,
		nil,
	)

	failOnError(err, "Failed to declare Welcome queue")

	err = ch.QueueBind(
		welcomeQueue.Name,
		"email.welcome",  
		"email_exchange",  
		false,
		nil,
	)
	failOnError(err, "Failed to bind Welcome queue")

	otpMsgs, err := ch.Consume(
		q.Name, 
		"otp_customer",
		false,
		false,
		false,
		false,
		nil,
	)

	failOnError(err, "Failed to register Welcome consumer")

	welcomeMsgs, err := ch.Consume(
		welcomeQueue.Name,    // queue
		"welcome_consumer", // consumer tag
		false,              // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	failOnError(err, "Failed to register Welcome consumer")

	log.Println("Notification service started. Waiting for email tasks...")

	var forever chan struct{}

	go func() {
		for d := range otpMsgs {
			log.Printf(" [OTP] Received email task: %s", d.Body)
			d.Ack(false)
		}
	}()

	go func() {
		for d := range welcomeMsgs {
			log.Printf(" [WELCOME] Received email task: %s", d.Body)
			

			d.Ack(false)
		}
	}()

	<-forever
}