package main

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	pstgre "tg-bot/internal/adapters/db"
	"tg-bot/internal/adapters/rabbitmq"
	"tg-bot/internal/adapters/telegram"
	"tg-bot/internal/handler"
	"tg-bot/internal/repository"
	"tg-bot/internal/service"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := loadConfig(); err != nil {
		logrus.Fatal("error initializing configs", err)
	}
	db := mustInitDB()
	rmq := mustInitRabbitMQ()
	repos := repository.NewRepository(db)
	services := service.NewService(repos, rmq)
	botAdapter := mustInitBot()
	handlers := handler.NewHandlers(botAdapter.Tg, services)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		startConsumer(ctx, rmq, services)
	}()

	startCron(services)
	wg.Add(1)
	go func() {
		defer wg.Done()
		handlers.Run(ctx)
	}()

	<-ctx.Done()
	logrus.Info("shutdown signal received, waiting goroutines...")
	rmq.Close()
	_ = db.Close()
	wg.Wait()
	logrus.Info("shutdown complete")
}

// Инициализация конфигов и env
func loadConfig() error {
	if err := initConfig(); err != nil {
		return err
	}
	return godotenv.Load(".env")
}

// Инициализация БД
func mustInitDB() *sqlx.DB {
	db, err := pstgre.NewPostgresDB(
		viper.GetString("db.host"),
		viper.GetString("db.port"),
		viper.GetString("db.username"),
		os.Getenv("POSTGRES_PASSWORD"),
		viper.GetString("db.dbname"),
		viper.GetString("db.sslmode"),
	)
	if err != nil {
		logrus.Fatalf("failed to init postgres: %s", err.Error())
	}
	if err := pstgre.RunMigrations(
		viper.GetString("db.host"),
		viper.GetString("db.port"),
		viper.GetString("db.username"),
		os.Getenv("POSTGRES_PASSWORD"),
		viper.GetString("db.dbname"),
		viper.GetString("db.sslmode"),
	); err != nil {
		logrus.Fatalf("failed to run migrations: %s", err.Error())
	}
	return db
}

// Инициализация RabbitMQ
func mustInitRabbitMQ() *rabbitmq.RabbitMQ {
	rmq, err := rabbitmq.NewRabbitMQ()
	if err != nil {
		logrus.Fatalf("RabbitMQ connect error: %s", err)
	}
	if _, err := rmq.DeclareQueue("user.events"); err != nil {
		logrus.Fatalf("queue declare error: %s", err)
	}
	return rmq
}

// Инициализация Telegram Bot
func mustInitBot() *telegram.BotAdapter {
	botAdapter, err := telegram.NewBot(os.Getenv("TOKEN_BOT"))
	if err != nil {
		log.Fatalf("Error creating telegram bot: %s", err)
	}
	return botAdapter
}

// Запуск Cron-задач
func startCron(services *service.Service) {
	c := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))
	_, err := c.AddFunc("0 10 * * *", func() {
		logrus.Info("cron: running CheckAndUpdateEvents")
		// services.Events.CheckAndUpdateEvents() // пример вызова бизнес-логики
	})
	if err != nil {
		logrus.Fatalf("cron add error: %v", err)
	}
	c.Start()
	go func() {
		<-context.Background().Done()
		c.Stop()
	}()
}

// Консьюмер RabbitMQ
func startConsumer(ctx context.Context, rmq *rabbitmq.RabbitMQ, services *service.Service) {
	q, err := rmq.DeclareQueue("user.events")
	if err != nil {
		logrus.Fatalf("DeclareQueue in consumer failed: %v", err)
	}
	msgs, err := rmq.Consume(q.Name)
	if err != nil {
		logrus.Fatalf("Consume error: %s", err)
	}
	for {
		select {
		case <-ctx.Done():
			logrus.Info("consumer: context canceled, stop consuming")
			return
		case msg, ok := <-msgs:
			if !ok {
				logrus.Warn("consumer: messages channel closed")
				return
			}
			logrus.Infof("Received event: %s", string(msg.Body))
			if err := services.Stats.HandleEvent(msg.Body); err != nil {
				logrus.Errorf("failed to handle event: %v", err)
				_ = msg.Nack(false, true)
			} else {
				_ = msg.Ack(false)
			}
		}
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
