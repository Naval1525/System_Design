package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"url-shortener/internal/database"
	"url-shortener/internal/handler"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	// database
	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := database.RunMigrations(); err != nil {
		log.Fatal("Migrations: ", err)
	}

	// repository
	urlRepo := repository.NewURLRepository(db)

	redis := database.NewRedisClient()
	defer redis.Close()
	log.Printf("Redis: connecting to %s", database.RedisAddr())
	if err := redis.Ping(database.Ctx).Err(); err != nil {
		log.Printf("Warning: Redis unreachable at %s (%v); cache will be skipped", database.RedisAddr(), err)
	} else {
		log.Printf("Redis: connected")
	}

	// service
	urlService := service.NewURLService(urlRepo, redis)

	// handler
	urlHandler := handler.NewURLHandler(urlService)

	// router
	router := handler.SetupRouter(urlHandler)

	log.Println("Server running on port 8080")

	http.ListenAndServe(":8080", router)
}
