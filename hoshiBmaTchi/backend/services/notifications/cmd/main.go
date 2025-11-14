package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
	"gopkg.in/gomail.v2"
)

type EmailTask struct{
	Email string `json:"email"`
	Subject string `json:"subject"`
	Body string `json:"body"`
}

func failOnError(err error, msg string){
	if err != nil{
		log.Fatalf("%s: %v", msg, err)
	}
}

func main(){

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	emailUser := os.Getenv("EMAIL_USER")
	emailPass := os.Getenv("EMAIL_APP_PASSWORD")

	if smtpHost == "" || smtpPortStr == "" || emailUser == "" || emailPass == "" {
		log.Fatal("FATAL: SMTP environment variables are not set")
	}

	smtpPort, err := strconv.Atoi(smtpPortStr)
	failOnError(err, "Invalid SMTP_PORT")

	dialer := gomail.NewDialer(smtpHost, smtpPort, emailUser, emailPass)

	dialerConn, err := dialer.Dial()
	failOnError(err, "Failed to connect to SMTP server")
	dialerConn.Close()
	log.Println("SMTP connection successful.")

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
		welcomeQueue.Name,    
		"welcome_consumer", 
		false,              
		false,              
		false,              
		false,              
		nil,                
	)
	failOnError(err, "Failed to register Welcome consumer")

	log.Println("Notification service started. Waiting for email tasks...")

	var forever chan struct{}

	go func() {
		for d := range otpMsgs {
			log.Printf(" [OTP] Received email task for: %s", d.Body)
			
			var task EmailTask
			if err := json.Unmarshal(d.Body, &task); err != nil {
				log.Printf("Error: Failed to parse task body: %v", err)
				d.Nack(false, false)
				continue
			}

			m := gomail.NewMessage()
			m.SetHeader("From", emailUser)
			m.SetHeader("To", task.Email)
			m.SetHeader("Subject", task.Subject)
			m.SetBody("text/html", task.Body)

			if err := dialer.DialAndSend(m); err != nil {
				log.Printf("Error: Failed to send OTP email: %v", err)
				d.Nack(false, true) 
			} else {
				log.Printf(" [OTP] Successfully sent email to %s", task.Email)
				d.Ack(false) 
			}
		}
	}()

	go func() {
		for d := range welcomeMsgs {
			log.Printf(" [WELCOME] Received email task: %s", d.Body)

			var task EmailTask
			if err := json.Unmarshal(d.Body, &task); err != nil {
				log.Printf("Error: Failed to parse task body: %v", err)
				d.Nack(false, false)
				continue
			}

			m := gomail.NewMessage()
			m.SetHeader("From", emailUser)
			m.SetHeader("To", task.Email)
			m.SetHeader("Subject", task.Subject)
			m.SetBody("text/html", task.Body)

			if err := dialer.DialAndSend(m); err != nil {
				log.Printf("Error: Failed to send welcome email: %v", err)
				d.Nack(false, true)
			} else {
				log.Printf(" [WELCOME] Successfully sent email to %s", task.Email)
				d.Ack(false)
			}
		}
	}()

	<-forever
}