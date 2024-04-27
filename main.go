package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

const (
	dbUsername = "bekzatbatyrkhanov"
	dbPassword = ""
	dbHost     = "localhost"
	dbPort     = 5432
	dbName     = "postgres"
)

var (
	redisClient *redis.Client
)

type Product struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

func initDB() (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUsername, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return db, nil
}

func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
}

func SetProductCache(product Product) error {
	productJSON, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal product to JSON: %v", err)
	}

	err = redisClient.HSet(ctx, "products", product.ID, productJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to set product in cache: %v", err)
	}

	return nil
}

func GetProductFromCache(productID string) (*Product, error) {
	productJSON, err := redisClient.HGet(ctx, "products", productID).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get product from cache: %v", err)
	}

	var product Product
	err = json.Unmarshal([]byte(productJSON), &product)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal product JSON: %v", err)
	}

	return &product, nil
}

func main() {
	initRedis()
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	product := Product{
		ID:    "1",
		Name:  "Example Product",
		Price: 100,
	}

	err = SetProductCache(product)
	if err != nil {
		log.Fatalf("Failed to set product in cache: %v", err)
	}

	retrievedProduct, err := GetProductFromCache("1")
	if err != nil {
		log.Fatalf("Failed to get product from cache: %v", err)
	}

	log.Printf("Retrieved product from cache: %+v\n", *retrievedProduct)
}
