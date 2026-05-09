package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"whale/database"
	"whale/handlers"
	"whale/kafka"
	"whale/repository"

	"github.com/spf13/viper"
)

func main() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("db.path", "./app.db")
	viper.SetDefault("kafka.brokers", "localhost:9092")
	viper.SetDefault("kafka.topic", "people")
	viper.SetDefault("kafka.group", "whale")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP") // APP_SERVER_PORT
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	port := viper.GetInt("server.port")
	bind := fmt.Sprintf(":%d", port)

	db := database.Init(viper.GetString("db.path"))

	repo := repository.NewPersonDBRepository(db)
	handle := handlers.NewPersonHandler(repo)

	brokers := strings.Split(viper.GetString("kafka.brokers"), ",")
	topic := viper.GetString("kafka.topic")
	group := viper.GetString("kafka.group")
	kafka.EnsureTopic(context.Background(), brokers, topic)
	consumer, err := kafka.NewConsumer(brokers, topic, group, repo)
	if err != nil {
		log.Fatalf("failed to create kafka consumer: %v", err)
	}
	go consumer.Run(context.Background())
	log.Printf("kafka consumer started brokers=%v topic=%s group=%s", brokers, topic, group)

	http.HandleFunc("/", handle.GetPerson)
	http.HandleFunc("/save", handle.PostPerson)

	log.Fatal(http.ListenAndServe(bind, nil))
}
