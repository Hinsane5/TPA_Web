package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"gopkg.in/gomail.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/notifications/internal/models"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/notifications/internal/repository"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/notifications/internal/ws"
)

type EmailTask struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

func main() {
	// 1. Load Environment Variables
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	emailUser := os.Getenv("EMAIL_USER")
	emailPass := os.Getenv("EMAIL_APP_PASSWORD")
	dbDSN := os.Getenv("DATABASE_URL")
	rabbitMQURL := "amqp://admin:rabbitmq_password_123@rabbitmq:5672/"

	if smtpHost == "" || smtpPortStr == "" || emailUser == "" || emailPass == "" {
		log.Fatal("FATAL: SMTP environment variables are not set")
	}

	smtpPort, err := strconv.Atoi(smtpPortStr)
	failOnError(err, "Invalid SMTP_PORT")

	// 2. Initialize Database (PostgreSQL)
	db, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{})
	failOnError(err, "Failed to connect to Database")
	
	err = db.AutoMigrate(&models.Notification{})
	failOnError(err, "Failed to migrate database")
	
	repo := repository.NewNotificationRepository(db)
	log.Println("Database connected and migrated.")

	// 3. Initialize WebSocket Hub
	hub := ws.NewHub()

	// 4. Initialize SMTP Dialer
	dialer := gomail.NewDialer(smtpHost, smtpPort, emailUser, emailPass)
	
	// Test SMTP connection
	dialerConn, err := dialer.Dial()
	failOnError(err, "Failed to connect to SMTP server")
	dialerConn.Close()
	log.Println("SMTP connection successful.")

	// 5. Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitMQURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare("email_exchange", "direct", true, false, false, false, nil)
	failOnError(err, "Failed to declare email_exchange")

	// OTP Queue
	qOTP, err := ch.QueueDeclare("email_queue", true, false, false, false, nil)
	failOnError(err, "Failed to declare email_queue")
	err = ch.QueueBind(qOTP.Name, "send_email", "email_exchange", false, nil)
	failOnError(err, "Failed to bind email_queue")

	// Welcome Queue
	qWelcome, err := ch.QueueDeclare("welcome_email_queue", true, false, false, false, nil)
	failOnError(err, "Failed to declare welcome_email_queue")
	err = ch.QueueBind(qWelcome.Name, "email.welcome", "email_exchange", false, nil)
	failOnError(err, "Failed to bind welcome_email_queue")

	err = ch.ExchangeDeclare("notification_exchange", "topic", true, false, false, false, nil)
	failOnError(err, "Failed to declare notification_exchange")

	qNotif, err := ch.QueueDeclare("notification_service_queue", true, false, false, false, nil)
	failOnError(err, "Failed to declare notification_service_queue")

	err = ch.QueueBind(qNotif.Name, "notification.*", "notification_exchange", false, nil)
	failOnError(err, "Failed to bind notification_service_queue")

	otpMsgs, err := ch.Consume(qOTP.Name, "otp_consumer", false, false, false, false, nil)
	failOnError(err, "Failed to register OTP consumer")

	welcomeMsgs, err := ch.Consume(qWelcome.Name, "welcome_consumer", false, false, false, false, nil)
	failOnError(err, "Failed to register Welcome consumer")

	notifMsgs, err := ch.Consume(qNotif.Name, "app_notif_consumer", true, false, false, false, nil)
	failOnError(err, "Failed to register Notification consumer")

	// --- OTP EMAIL CONSUMER (With Original Logs) ---
	go func() {
		for d := range otpMsgs {
			// RESTORED: Logging the body so you can see the OTP content
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

	// --- WELCOME EMAIL CONSUMER (With Original Logs) ---
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
	
	// --- REAL-TIME NOTIFICATION CONSUMER ---
	go func() {
		for d := range notifMsgs {
			log.Printf(" [NOTIF] Received event: %s", d.Body)

			var event models.NotificationEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf("Error: Failed to parse notification event: %v", err)
				continue
			}

			// 1. Save to Database
			notif := models.Notification{
				RecipientID: event.RecipientID,
				SenderID:    event.SenderID,
				SenderName:  event.SenderName,
				SenderImage: event.SenderImage,
				Type:        event.Type,
				EntityID:    event.EntityID,
				Message:     event.Message,
				IsRead:      false,
			}
			
			if err := repo.Create(&notif); err != nil {
				log.Printf("Error saving notification to DB: %v", err)
			}

			// 2. Send to WebSocket
			hub.SendNotification(event.RecipientID, notif)
		}
	}()

	log.Println("Notification Service Started. Listening for Emails and Events...")

	// ---------------------------------------------------------
	// D. Start HTTP Server
	// ---------------------------------------------------------
	
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/notifications/:userID", func(c *gin.Context) {
		userIDStr := c.Param("userID")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
			return
		}

		notifs, err := repo.GetByUserID(uint(userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
			return
		}

		c.JSON(http.StatusOK, notifs)
	})

	r.GET("/ws", func(c *gin.Context) {
		token := c.Query("token")
		// In production: validate token. 
		// For development/demo, we allow passing user_id directly if needed, or parse token.
		userIDStr := c.Query("user_id") 
		if userIDStr == "" {
			// Logic to extract from token would go here
			log.Printf("Warning: connecting WS with token: %s", token)
		}
		
		// Fallback for simple testing
		if userIDStr == "" {
			userIDStr = "1" 
		}

		userID, _ := strconv.Atoi(userIDStr)
		ws.ServeWs(hub, c.Writer, c.Request, uint(userID))
	})

	if err := r.Run(":8084"); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}