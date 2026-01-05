package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"broker/logs"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const webPort = "80"

type Config struct {
	Rabbit     *amqp.Connection
	GRPCClient logs.LogServiceClient
	GRPCConn   *grpc.ClientConn
	HTTPClient *http.Client
}

func main() {
	rabbitconn, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitconn.Close()

	// Connect to gRPC service (reuse connection)
	grpcConn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Warning: Failed to connect to gRPC service: %v\n", err)
		log.Println("gRPC features will be unavailable")
	} else {
		log.Println("Connected to gRPC logger service")
	}
	defer func() {
		if grpcConn != nil {
			grpcConn.Close()
		}
	}()

	var grpcClient logs.LogServiceClient
	if grpcConn != nil {
		grpcClient = logs.NewLogServiceClient(grpcConn)
	}

	// Create HTTP client with connection pooling
	httpClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	app := Config{
		Rabbit:     rabbitconn,
		GRPCClient: grpcClient,
		GRPCConn:   grpcConn,
		HTTPClient: httpClient,
	}

	log.Printf("Starting broker service on port %s\n", webPort)

	// define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start the server
	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	// dont continue until rabbit is ready
	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("RabbitMQ not yet ready...")
			counts++
		} else {
			connection = c
			log.Println("connected to rabbitmq")
			break
		}
		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off....")
		time.Sleep(backOff)
		continue
	}

	return connection, nil
}
