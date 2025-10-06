package main

import (
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"log"
	"os"
	pstgre "tg-bot/internal/adapters/db"
	"tg-bot/internal/adapters/rabbitmq"
	"tg-bot/internal/adapters/telegram"
	"tg-bot/internal/repository"
	"tg-bot/internal/service"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	//Конфигская параша
	if err := initConfig(); err != nil {
		logrus.Fatal("error initializing configs", err)
	}
	if err := godotenv.Load(".env"); err != nil {
		logrus.Fatal("error initializing configs", err)
	}
	//db fignia
	db, err := pstgre.NewPostgresDB(
		viper.GetString("db.host"),
		viper.GetString("db.port"),
		viper.GetString("db.username"),
		os.Getenv("POSTGRES_PASSWORD"),
		viper.GetString("db.dbname"),
		viper.GetString("db.sslmode"))
	if err != nil {
		logrus.Fatalf("failed to init postgres: %s", err.Error())
	}
	err = pstgre.RunMigrations(
		viper.GetString("db.host"),
		viper.GetString("db.port"),
		viper.GetString("db.username"),
		os.Getenv("POSTGRES_PASSWORD"),
		viper.GetString("db.dbname"),
		viper.GetString("db.sslmode"),
	)
	if err != nil {
		logrus.Fatalf("failed to run migrations: %s", err.Error())
	}

	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("DBCLOSE error")
		}
	}(db)

	rmq, err := rabbitmq.NewRabbitMQ()
	if err != nil {
		logrus.Fatalf("RabbitMQ connect error: %s", err)
	}
	defer rmq.Close()
	if _, err := rmq.DeclareQueue("user.events"); err != nil {
		logrus.Fatalf("queue declare error: %s", err)
	}

	repos := repository.NewRepository(db)
	services := service.NewService(repos, rmq)

	//Ботяра Жарас
	bot, err := telegram.InitBot(os.Getenv("TOKEN_BOT"), services)
	if err != nil {
		log.Fatalf("Error with InitBot: %s", err.Error())
	}
	if err := telegram.CheckBot(bot); err != nil {
		log.Fatalf("Error with CheckBot: %s", err.Error())
	}
	c := cron.New()

	// Каждый день в 10:00 проверяем новые события
	_, err = c.AddFunc("0 10 * * *", func() {
		CheckAndUpdateEvent(services)
	})
	if err != nil {
		logrus.Fatalf("Error with cron: %s", err.Error())
		return
	}

	c.Start()
	defer c.Stop()
	//Зайчик
	go startConsumer(rmq, services)
	CheckAndUpdateEvent(services)
	go bot.Start()
	select {}
}
func CheckAndUpdateEvent(services *service.Service) {
	logrus.Info("🕙 Проверяем Ticketon на новые события...")
	if err := services.Events.CheckAndUpdateEvents(); err != nil {
		logrus.Errorf("Ошибка при обновлении событий: %v", err)
	}
}
func startConsumer(rmq *rabbitmq.RabbitMQ, services *service.Service) {
	msgs, err := rmq.Consume("user.events")
	if err != nil {
		logrus.Fatalf("Consume error: %s", err)
	}

	for msg := range msgs {
		logrus.Infof("Received event: %s", string(msg.Body))

		// 📊 отправляем в stats сервис
		if err := services.Stats.HandleEvent(msg.Body); err != nil {
			logrus.Errorf("failed to handle event: %s", err)
		}

		// подтверждаем обработку
		_ = msg.Ack(false)
	}
}
func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
