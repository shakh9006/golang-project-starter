package app

import (
	"context"
	"example.com/config"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
)

var (
	ctx context.Context
)

func initConfig() (config.Config, error) {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}

func initMongoDB(cfg config.Config) (*mongo.Client, error) {
	mongoConn := options.Client().ApplyURI(cfg.DBUri)
	mongoClient, err := mongo.Connect(ctx, mongoConn)

	if err != nil {
		return nil, err
	}

	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	fmt.Println("MongoDB successfully connected...")

	return mongoClient, nil
}

func initRedis(cfg config.Config) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisUri,
	})

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	err := redisClient.Set(ctx, "test", "Welcome to Goland with Redis and MongoDB", 0).Err()
	if err != nil {
		return nil, err
	}

	fmt.Println("Redis client connected successfully...")
	return redisClient, err
}

func Run() {
	ctx = context.TODO()
	cfg, err := initConfig()
	if err != nil {
		log.Fatal("Could not load environment variables")
	}

	mongoClient, err := initMongoDB(cfg)
	if err != nil {
		panic(err)
	}
	defer mongoClient.Disconnect(ctx)

	redisClient, err := initRedis(cfg)
	if err != nil {
		panic(err)
	}
	value, err := redisClient.Get(ctx, "test").Result()
	if err == redis.Nil {
		fmt.Println("key: test does not exist")
	} else if err != nil {
		panic(err)
	}

	server := gin.Default()
	router := server.Group("/api")
	router.GET("/healthchecker", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": value})
	})

	log.Fatal(server.Run(fmt.Sprintf(":%s", cfg.Port)))
}
