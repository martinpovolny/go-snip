package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"whale/database"
	"whale/handlers"
	"whale/repository"

	"github.com/spf13/viper"
)

func main() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("db.path", "./app.db")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP") // APP_SERVER_PORT
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	port := viper.GetInt("server.port")
	bind := fmt.Sprintf(":%d", port)

	db := database.Init(viper.GetString("db.path"))

	repo := repository.NewPersonDBRepository(db)
	handle := handlers.NewPersonHandler(repo)

	http.HandleFunc("/", handle.GetPerson)
	http.HandleFunc("/save", handle.PostPerson)

	log.Fatal(http.ListenAndServe(bind, nil))
}
