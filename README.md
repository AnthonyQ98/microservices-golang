# Performance Analysis: Broker → Logger Communication Methods

This document analyzes the performance characteristics of three communication methods between the broker and logger micro services.

## Communication Methods

### 1. Direct HTTP (Synchronous)
- **Path**: Broker → Logger (direct HTTP POST)
- **Protocol**: HTTP/1.1 with JSON
- **Connection**: Reused HTTP client with connection pooling
- **Blocking**: Yes (waits for response)

### 2. RabbitMQ (Asynchronous)
- **Path**: Broker → RabbitMQ → Listener Service → Logger (HTTP)
- **Protocol**: AMQP → HTTP/1.1 with JSON
- **Blocking**: No (broker returns immediately after queueing)
- **Characteristics**: Decoupled, message queue with intermediate service

### 3. gRPC (Synchronous)
- **Path**: Broker → Logger (direct gRPC call)
- **Protocol**: gRPC with Protocol Buffers
- **Connection**: Reused persistent connection
- **Blocking**: Yes (waits for response)

## Actual Benchmark Results

Benchmarks run with 1000 requests, 50 concurrent connections:

### Latency (Time to Complete)

1. **gRPC** - **Fastest** ⚡
   - **Average Latency**: 5.99ms
   - **Min Latency**: 0.68ms
   - **Max Latency**: 17.25ms
   - Binary protocol (Protobuf) is more efficient than JSON
   - Persistent connection (no handshake overhead)
   - Lower overhead than HTTP

2. **Direct HTTP** - **Medium** 🚀
   - **Average Latency**: 9.70ms
   - **Min Latency**: 0.74ms
   - **Max Latency**: 44.76ms
   - Text-based JSON encoding/decoding
   - HTTP headers add overhead
   - Connection pooling reduces overhead

3. **RabbitMQ** - **Slowest** (but non-blocking) 🐰
   - **Average Latency**: 18.47ms
   - **Min Latency**: 6.14ms
   - **Max Latency**: 38.79ms
   - Multiple hops: Broker → RabbitMQ → Listener → Logger
   - Message serialization/deserialization
   - Queue operations
   - HTTP call from listener to logger
   - **Note**: Broker returns immediately after queueing (async)

### Throughput (Requests per Second)

1. **gRPC** - **Highest** 🏆
   - **Throughput**: 7,778 req/s
   - Efficient binary protocol
   - Connection reuse
   - Lower CPU overhead

2. **Direct HTTP** - **Medium**
   - **Throughput**: 4,923 req/s
   - JSON parsing overhead
   - HTTP protocol overhead

3. **RabbitMQ** - **Lowest**
   - **Throughput**: 2,644 req/s
   - Depends on queue processing speed
   - Can handle bursts better (decoupling)
   - **Note**: Broker throughput is higher since it returns immediately

## Trade-offs

### gRPC
✅ **Pros:**
- Fastest latency
- Highest throughput
- Type-safe with Protobuf
- Connection reuse (fixed)

❌ **Cons:**
- Synchronous (blocks until response)
- Requires both services to be available
- More complex setup

### Direct HTTP
✅ **Pros:**
- Simple and standard
- Easy to debug
- Works through firewalls/proxies easily
- Connection pooling (fixed)

❌ **Cons:**
- Slower than gRPC
- JSON parsing overhead
- Synchronous (blocks until response)

### RabbitMQ
✅ **Pros:**
- **Non-blocking** (broker returns immediately)
- Decoupled (services can be temporarily unavailable)
- Can handle traffic spikes
- Message persistence
- Retry capabilities

❌ **Cons:**
- Highest latency (end-to-end)
- More complex architecture
- Additional service (listener) required
- Eventual consistency

## When to Use Each Method

### Use gRPC when:
- You need **lowest latency** and **highest throughput**
- Both services are always available
- You need synchronous responses
- Performance is critical

### Use Direct HTTP when:
- You need simplicity and standard protocols
- Debugging and observability are important
- You need to work through proxies/firewalls
- Moderate performance is acceptable

### Use RabbitMQ when:
- You need **asynchronous processing**
- Services may be temporarily unavailable
- You need to handle traffic spikes
- You want decoupling between services
- You need message persistence/retry

## Running Benchmarks

To compare actual performance, use the benchmark tool:

```bash
cd benchmark
go run main.go [broker-url] [requests] [concurrency]
```

Example:
```bash
go run main.go http://localhost:8080 1000 50
```

This will test all three methods and show:
- Average latency
- Throughput (req/s)
- Success/failure rates
- Min/max latencies

### Latest Benchmark Results (1000 requests, 50 concurrency)

```
Method           Avg Latency Throughput      Success      Failure
----------------------------------------------------------------
HTTP              9.703734ms    4923.34        1000           0
RabbitMQ         18.469035ms    2643.52        1000           0
gRPC              5.985075ms    7778.22        1000           0
```

**Key Findings:**
- **gRPC is 1.6x faster** than HTTP in latency
- **gRPC is 1.6x higher** in throughput than HTTP
- **gRPC is 2.9x faster** than RabbitMQ in latency
- **gRPC is 2.9x higher** in throughput than RabbitMQ
- **HTTP is 1.9x faster** than RabbitMQ in latency
- **HTTP is 1.9x higher** in throughput than RabbitMQ

## Performance Summary

Based on actual benchmarks:

| Method | Avg Latency | Throughput | Best For |
|--------|-------------|------------|----------|
| **gRPC** | 5.99ms | 7,778 req/s | High performance, low latency needs |
| **HTTP** | 9.70ms | 4,923 req/s | Standard protocols, simplicity |
| **RabbitMQ** | 18.47ms | 2,644 req/s | Async processing, decoupling |

**gRPC wins on both latency and throughput**, making it the best choice for synchronous high-performance scenarios.

## Conclusion

- **For synchronous, high-performance needs**: Use **gRPC**
- **For simplicity and standard protocols**: Use **Direct HTTP**
- **For asynchronous, decoupled processing**: Use **RabbitMQ**

The choice depends on your specific requirements for latency, throughput, availability, and coupling.

