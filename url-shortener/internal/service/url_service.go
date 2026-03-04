package service

import (
	"log"
	"time"

	"github.com/go-redis/redis/v8"

	"url-shortener/internal/database"
	"url-shortener/internal/repository"
	base62 "url-shortener/pkg/hash"
)
type URLService struct {
	Repo  *repository.URLRepository
	Redis *redis.Client
}

func NewURLService(repo *repository.URLRepository, redis *redis.Client) *URLService {
	return &URLService{Repo: repo, Redis: redis}
}
func (s *URLService) ShortenURL(original string) (string, error) {

	url, err := s.Repo.CreateURL(original)
	if err != nil {
		return "", err
	}

	shortCode := base62.Encode(url.ID)

	err = s.Repo.UpdateShortCode(url.ID, shortCode)
	if err != nil {
		return "", err
	}

	if err := s.Redis.Set(database.Ctx, shortCode, original, time.Hour*24).Err(); err != nil {
		log.Printf("Redis Set shortCode=%q: %v", shortCode, err)
	} else {
		log.Printf("Redis: cached shortCode=%q", shortCode)
	}

	return shortCode, nil
}
func (s *URLService) GetOriginalURL(code string) (string, error) {

	// 1️⃣ check cache
	val, err := s.Redis.Get(database.Ctx, code).Result()

	if err == nil {
		return val, nil
	}

	// 2️⃣ fallback to DB
	url, err := s.Repo.GetByShortCode(code)
	if err != nil {
		return "", err
	}

	// 3️⃣ store in cache
	if err := s.Redis.Set(database.Ctx, code, url.OriginalURL, 24*time.Hour).Err(); err != nil {
		log.Printf("Redis Set code=%q: %v", code, err)
	}

	return url.OriginalURL, nil
}
