package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type RequestPayload struct {
	Action string    `json:"action"`
	Log    LogPayload `json:"log"`
}

type BenchmarkResult struct {
	Method      string
	TotalTime   time.Duration
	Requests    int
	Successes   int
	Failures    int
	AvgLatency  time.Duration
	MinLatency  time.Duration
	MaxLatency  time.Duration
	Throughput  float64 // requests per second
}

func main() {
	brokerURL := "http://broker-service"
	if len(os.Args) > 1 {
		brokerURL = os.Args[1]
	}

	requests := 100
	if len(os.Args) > 2 {
		fmt.Sscanf(os.Args[2], "%d", &requests)
	}

	concurrency := 10
	if len(os.Args) > 3 {
		fmt.Sscanf(os.Args[3], "%d", &concurrency)
	}

	fmt.Printf("Benchmarking broker->logger communication methods\n")
	fmt.Printf("Broker URL: %s\n", brokerURL)
	fmt.Printf("Requests: %d\n", requests)
	fmt.Printf("Concurrency: %d\n\n", concurrency)

	// Warm up
	fmt.Println("Warming up...")
	warmup(brokerURL)

	// Benchmark each method
	fmt.Println("\n=== Benchmarking HTTP (Direct) ===")
	httpResult := benchmarkHTTP(brokerURL, requests, concurrency)
	printResult(httpResult)

	fmt.Println("\n=== Benchmarking RabbitMQ ===")
	rabbitResult := benchmarkRabbitMQ(brokerURL, requests, concurrency)
	printResult(rabbitResult)

	fmt.Println("\n=== Benchmarking gRPC ===")
	grpcResult := benchmarkGRPC(brokerURL, requests, concurrency)
	printResult(grpcResult)

	// Summary
	fmt.Println("\n=== SUMMARY ===")
	fmt.Printf("%-15s %12s %10s %12s %12s\n", "Method", "Avg Latency", "Throughput", "Success", "Failure")
	fmt.Println("----------------------------------------------------------------")
	printSummary(httpResult)
	printSummary(rabbitResult)
	printSummary(grpcResult)
}

func warmup(brokerURL string) {
	payload := RequestPayload{
		Action: "log",
		Log: LogPayload{
			Name: "warmup",
			Data: "warmup data",
		},
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", brokerURL+"/log-http", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	client.Do(req)
	time.Sleep(100 * time.Millisecond)
}

func benchmarkHTTP(brokerURL string, totalRequests, concurrency int) BenchmarkResult {
	start := time.Now()
	latencies := make([]time.Duration, 0, totalRequests)
	var mu sync.Mutex
	var wg sync.WaitGroup
	successes := 0
	failures := 0

	requestsPerGoroutine := totalRequests / concurrency

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: 5 * time.Second}

			for j := 0; j < requestsPerGoroutine; j++ {
				payload := RequestPayload{
					Action: "log",
					Log: LogPayload{
						Name: fmt.Sprintf("http-test-%d", j),
						Data: fmt.Sprintf("HTTP benchmark data %d", j),
					},
				}

				jsonData, _ := json.Marshal(payload)
				reqStart := time.Now()
				req, err := http.NewRequest("POST", brokerURL+"/log-http", bytes.NewBuffer(jsonData))
				if err != nil {
					mu.Lock()
					failures++
					mu.Unlock()
					continue
				}

				req.Header.Set("Content-Type", "application/json")
				resp, err := client.Do(req)
				latency := time.Since(reqStart)

				mu.Lock()
				if err != nil || resp == nil || resp.StatusCode != http.StatusAccepted {
					failures++
				} else {
					successes++
					latencies = append(latencies, latency)
					resp.Body.Close()
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	totalTime := time.Since(start)

	return calculateResult("HTTP", totalTime, latencies, successes, failures, totalRequests)
}

func benchmarkRabbitMQ(brokerURL string, totalRequests, concurrency int) BenchmarkResult {
	start := time.Now()
	latencies := make([]time.Duration, 0, totalRequests)
	var mu sync.Mutex
	var wg sync.WaitGroup
	successes := 0
	failures := 0

	requestsPerGoroutine := totalRequests / concurrency

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: 5 * time.Second}

			for j := 0; j < requestsPerGoroutine; j++ {
				payload := RequestPayload{
					Action: "log",
					Log: LogPayload{
						Name: fmt.Sprintf("rabbit-test-%d", j),
						Data: fmt.Sprintf("RabbitMQ benchmark data %d", j),
					},
				}

				jsonData, _ := json.Marshal(payload)
				reqStart := time.Now()
				req, err := http.NewRequest("POST", brokerURL+"/log-rabbit", bytes.NewBuffer(jsonData))
				if err != nil {
					mu.Lock()
					failures++
					mu.Unlock()
					continue
				}

				req.Header.Set("Content-Type", "application/json")
				resp, err := client.Do(req)
				latency := time.Since(reqStart)

				mu.Lock()
				if err != nil || resp == nil || resp.StatusCode != http.StatusAccepted {
					failures++
				} else {
					successes++
					latencies = append(latencies, latency)
					resp.Body.Close()
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	totalTime := time.Since(start)

	return calculateResult("RabbitMQ", totalTime, latencies, successes, failures, totalRequests)
}

func benchmarkGRPC(brokerURL string, totalRequests, concurrency int) BenchmarkResult {
	start := time.Now()
	latencies := make([]time.Duration, 0, totalRequests)
	var mu sync.Mutex
	var wg sync.WaitGroup
	successes := 0
	failures := 0

	requestsPerGoroutine := totalRequests / concurrency

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: 5 * time.Second}

			for j := 0; j < requestsPerGoroutine; j++ {
				payload := RequestPayload{
					Action: "log",
					Log: LogPayload{
						Name: fmt.Sprintf("grpc-test-%d", j),
						Data: fmt.Sprintf("gRPC benchmark data %d", j),
					},
				}

				jsonData, _ := json.Marshal(payload)
				reqStart := time.Now()
				req, err := http.NewRequest("POST", brokerURL+"/log-grpc", bytes.NewBuffer(jsonData))
				if err != nil {
					mu.Lock()
					failures++
					mu.Unlock()
					continue
				}

				req.Header.Set("Content-Type", "application/json")
				resp, err := client.Do(req)
				latency := time.Since(reqStart)

				mu.Lock()
				if err != nil || resp == nil || resp.StatusCode != http.StatusAccepted {
					failures++
				} else {
					successes++
					latencies = append(latencies, latency)
					resp.Body.Close()
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	totalTime := time.Since(start)

	return calculateResult("gRPC", totalTime, latencies, successes, failures, totalRequests)
}

func calculateResult(method string, totalTime time.Duration, latencies []time.Duration, successes, failures, totalRequests int) BenchmarkResult {
	if len(latencies) == 0 {
		return BenchmarkResult{
			Method:    method,
			TotalTime: totalTime,
			Requests:  totalRequests,
			Successes: successes,
			Failures:  failures,
		}
	}

	var sum time.Duration
	min := latencies[0]
	max := latencies[0]

	for _, lat := range latencies {
		sum += lat
		if lat < min {
			min = lat
		}
		if lat > max {
			max = lat
		}
	}

	avgLatency := sum / time.Duration(len(latencies))
	throughput := float64(successes) / totalTime.Seconds()

	return BenchmarkResult{
		Method:     method,
		TotalTime:  totalTime,
		Requests:   totalRequests,
		Successes:  successes,
		Failures:   failures,
		AvgLatency: avgLatency,
		MinLatency: min,
		MaxLatency: max,
		Throughput: throughput,
	}
}

func printResult(result BenchmarkResult) {
	fmt.Printf("Total Time:     %v\n", result.TotalTime)
	fmt.Printf("Requests:       %d\n", result.Requests)
	fmt.Printf("Successes:      %d\n", result.Successes)
	fmt.Printf("Failures:       %d\n", result.Failures)
	if result.Successes > 0 {
		fmt.Printf("Avg Latency:    %v\n", result.AvgLatency)
		fmt.Printf("Min Latency:    %v\n", result.MinLatency)
		fmt.Printf("Max Latency:    %v\n", result.MaxLatency)
		fmt.Printf("Throughput:     %.2f req/s\n", result.Throughput)
	}
}

func printSummary(result BenchmarkResult) {
	if result.Successes > 0 {
		fmt.Printf("%-15s %12s %10.2f %11d %11d\n",
			result.Method,
			result.AvgLatency.String(),
			result.Throughput,
			result.Successes,
			result.Failures)
	} else {
		fmt.Printf("%-15s %12s %10s %11d %11d\n",
			result.Method,
			"N/A",
			"N/A",
			result.Successes,
			result.Failures)
	}
}

