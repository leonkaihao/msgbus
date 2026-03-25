# msgbus

[![CI](https://github.com/leonkaihao/msgbus/actions/workflows/ci.yml/badge.svg)](https://github.com/leonkaihao/msgbus/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/leonkaihao/msgbus)](https://github.com/leonkaihao/msgbus)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A unified message bus abstraction library for Go that provides a consistent interface across multiple messaging protocols including NATS, MQTT, and in-process messaging.

## Features

- 🔌 **Unified Interface**: Single API for multiple messaging backends (NATS, MQTT3, in-process)
- 🚀 **Fire & Forget / Request-Response**: Support for both async fire-and-forget and synchronous request-response patterns
- 📦 **Multiple Payload Formats**: Built-in support for JSON (msgbus), binary (edgebus), and raw payloads
- 🔄 **Consumer Groups**: Automatic load balancing across consumer groups
- 🎯 **Topic Wildcards**: Pattern-based topic subscription
- 🛡️ **Production Ready**: Battle-tested with comprehensive test coverage
- 🐳 **Kubernetes Ready**: Includes deployment manifests and example applications

## Supported Protocols

| Protocol | Client Type | Description |
|----------|-------------|-------------|
| **NATS** | `client.CLI_NATS` | High-performance cloud-native messaging |
| **MQTT v3** | `client.CLI_MQTT3` | IoT-focused pub/sub messaging |
| **In-Process** | `client.CLI_INPROC` | Lock-free in-memory messaging for testing/local use |

## Installation

```bash
go get github.com/leonkaihao/msgbus@latest
```

## Quick Start

### Producer (Fire and Forget)

```go
package main

import (
    "github.com/leonkaihao/msgbus/pkg/client"
    "github.com/leonkaihao/msgbus/pkg/model"
    log "github.com/sirupsen/logrus"
)

func main() {
    // Create a NATS client
    cli, err := client.NewBuilder(client.CLI_NATS).Build()
    if err != nil {
        log.Fatal(err)
    }
    defer cli.Close()

    // Connect to broker
    broker := cli.Broker("nats://localhost:4222")
    
    // Create producer
    producer := broker.Producer("MyProducer", "events.user.created")
    
    // Set payload type (optional, defaults to msgbus format)
    producer.SetOption(model.PRD_OPTION_PAYLOAD, model.PLTYPE_MSGBUS)
    
    // Send message (fire and forget)
    data := []byte(`{"userId": 123, "name": "John"}`)
    metadata := map[string]string{"source": "api-server"}
    
    err = producer.Fire(data, metadata)
    if err != nil {
        log.Errorf("Failed to send: %v", err)
    }
}
```

### Producer (Request-Response)

```go
package main

import (
    "context"
    "time"
    "github.com/leonkaihao/msgbus/pkg/client"
    log "github.com/sirupsen/logrus"
)

func main() {
    cli, _ := client.NewBuilder(client.CLI_NATS).Build()
    defer cli.Close()

    broker := cli.Broker("nats://localhost:4222")
    producer := broker.Producer("MyProducer", "rpc.getUser")
    
    // Send request and wait for response
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    request := []byte(`{"userId": 123}`)
    response, err := producer.Request(ctx, request, nil)
    if err != nil {
        log.Errorf("Request failed: %v", err)
        return
    }
    
    log.Infof("Got response: %s", string(response))
}
```

### Consumer

```go
package main

import (
    "github.com/leonkaihao/msgbus/pkg/client"
    "github.com/leonkaihao/msgbus/pkg/model"
    log "github.com/sirupsen/logrus"
)

func main() {
    cli, _ := client.NewBuilder(client.CLI_NATS).Build()
    defer cli.Close()

    broker := cli.Broker("nats://localhost:4222")
    
    // Create consumer with topic and consumer group
    consumer := broker.Consumer("MyConsumer", "events.user.*", "user-service")
    
    // Subscribe to messages
    msgChan, err := consumer.Subscribe()
    if err != nil {
        log.Fatal(err)
    }
    
    // Process messages
    for msg := range msgChan {
        log.Infof("Received from %s [seq:%d]: %s", 
            msg.Source(), msg.Seq(), string(msg.Data()))
        
        // For request-response: send acknowledgment
        if msg.Dest() != "" {
            msg.Ack([]byte(`{"status": "processed"}`))
        }
    }
}
```

### Service Pattern (Recommended)

For production use, leverage the built-in Service abstraction for automatic lifecycle management:

```go
package main

import (
    "github.com/leonkaihao/msgbus/pkg/client"
    "github.com/leonkaihao/msgbus/pkg/common"
    "github.com/leonkaihao/msgbus/pkg/model"
    log "github.com/sirupsen/logrus"
)

type MyServiceHandler struct{}

func (h *MyServiceHandler) OnReceive(consumer model.Consumer, msg model.Messager) {
    log.Infof("[%s] Received message on topic %s: %s", 
        consumer.Name(), msg.Topic(), string(msg.Data()))
    
    // Process message...
    
    // Acknowledge if needed
    if msg.Dest() != "" {
        msg.Ack([]byte(model.CONSUMER_ACK))
    }
}

func (h *MyServiceHandler) OnError(inst interface{}, err error) {
    log.Errorf("Error: %v", err)
}

func (h *MyServiceHandler) OnDestroy(inst interface{}) {
    log.Info("Service destroyed")
}

func main() {
    cli, _ := client.NewBuilder(client.CLI_NATS).Build()
    defer cli.Close()

    // Create service with callbacks
    svc := common.NewService(&MyServiceHandler{})
    defer svc.Close()

    // Add consumers
    broker := cli.Broker("nats://localhost:4222")
    consumers := []model.Consumer{
        broker.Consumer("Worker1", "tasks.process", "workers"),
        broker.Consumer("Worker2", "tasks.analyze", "workers"),
    }
    svc.AddConsumers(consumers)

    // Start processing
    svc.Serve()
}
```

## Architecture

### Core Interfaces

**Client → Broker → Producer/Consumer**

```
Client
  └── Broker (connected to message broker endpoint)
       ├── Producer (publishes to topics)
       └── Consumer (subscribes to topics with consumer group)
```

### Key Interfaces

```go
// Client: Entry point for creating brokers
type Client interface {
    Broker(url string) Broker
    Remove(brk Broker) error
    Close() error
}

// Broker: Manages producers and consumers
type Broker interface {
    URL() string
    Consumer(name, topic, group string) Consumer
    Producer(name, topic string) Producer
    Close() error
}

// Producer: Publishes messages
type Producer interface {
    Fire(data []byte, metadata map[string]string) error
    Request(ctx context.Context, data []byte, metadata map[string]string) ([]byte, error)
    SetOption(k, v string) error
    Close() error
}

// Consumer: Subscribes to messages
type Consumer interface {
    Subscribe() (<-chan Messager, error)
    Close() error
}
```

## Payload Formats

msgbus supports three payload encoding formats:

| Format | Type Constant | Description | Use Case |
|--------|---------------|-------------|----------|
| **msgbus** | `model.PLTYPE_MSGBUS` | JSON wrapper with data + metadata | Default, human-readable |
| **edgebus** | `model.PLTYPE_EDGEBUS` | Binary format with header `0xA1 0x60` | High-performance edge devices |
| **raw** | `model.PLTYPE_RAW` | Pass-through, no encoding | Direct byte payload |

Set payload type on producer:

```go
producer.SetOption(model.PRD_OPTION_PAYLOAD, model.PLTYPE_MSGBUS)
```

## Configuration File Format

For running the example apps, use a JSON config file:

```json
{
  "loglevel": "info",
  "producer": {
    "endpoint": "nats://localhost:4222",
    "events": [
      {
        "topic": "events.user.created",
        "repeat": 100,
        "delay": 0,
        "sleep": 1000,
        "mode": "fire",
        "payload": "msgbus",
        "data": {"userId": 123, "action": "created"},
        "metadata": {"source": "api"}
      }
    ]
  },
  "consumers": [
    {
      "endpoint": "nats://localhost:4222",
      "show": "all",
      "topics": [
        {
          "topic": "events.user.*",
          "group": "user-service"
        }
      ]
    }
  ]
}
```

Run examples:

```bash
./app/producer/producer config.json
./app/consumer/consumer config.json
```

## Development Setup

### Prerequisites

- Go 1.18+
- Docker
- kubectl
- kind (Kubernetes in Docker)

### Running with Kubernetes (Full Stack)

1. **Create Kubernetes cluster**:
   ```bash
   kind create cluster
   ```

2. **Build and push images**:
   ```bash
   make build-consumer
   make build-producer
   kind load docker-image leonkaihao/nats-consumer:1.1.1
   kind load docker-image leonkaihao/nats-producer:1.1.1
   ```

3. **Deploy NATS stack**:
   ```bash
   make deploy-nats-stack
   ```

4. **View logs**:
   ```bash
   kubectl logs -f deployment/producer
   kubectl logs -f deployment/consumer
   ```

5. **Monitor with Grafana** (optional):
   ```bash
   make deploy-nats-monitor
   kubectl port-forward deployment/grafana 3000:3000
   # Open http://localhost:3000 (admin/admin)
   # Add datasource: http://prometheus:9090
   # Import dashboard: 2279
   ```

### Running Tests

```bash
# Unit tests
make test

# Integration tests (requires NATS server)
make test/integration

# Benchmark tests
make test/bench
```

## Production Considerations

- **Connection Resilience**: All clients implement automatic reconnection
- **Consumer Groups**: Use consumer groups for load balancing and high availability
- **Context Timeouts**: Always use context with timeout for `Request()` calls
- **Graceful Shutdown**: Properly close clients, brokers, and consumers on shutdown
- **Monitoring**: Track message rates via `Producer.CurrentSeq()` and `Messager.Seq()`

## API Reference

### Producer Methods

| Method | Description |
|--------|-------------|
| `Fire(data, metadata)` | Fire-and-forget publish |
| `Request(ctx, data, metadata)` | Request-response publish with timeout |
| `SetOption(key, value)` | Set producer options (e.g., payload type) |
| `Topic()` | Get bound topic |
| `CurrentSeq()` | Get current sequence number |
| `Close()` | Release resources |

### Consumer Methods

| Method | Description |
|--------|-------------|
| `Subscribe()` | Subscribe and get message channel |
| `Topic()` | Get subscribed topic pattern |
| `Group()` | Get consumer group name |
| `Close()` | Unsubscribe and cleanup |

### Message Methods

| Method | Description |
|--------|-------------|
| `Data()` | Get message payload |
| `Metadata()` | Get metadata map |
| `Topic()` | Get actual topic (resolved from wildcard) |
| `Source()` | Get producer ID |
| `Seq()` | Get message sequence number |
| `Ack(data)` | Send acknowledgment (for request-response) |

## Examples

See the `/app` directory for complete examples:

- **[Producer](/app/producer)**: Configurable message producer with fire/request modes
- **[Consumer](/app/consumer)**: Service-based consumer with callbacks
- **[MQTT Example](/app/mqtt-example)**: MQTT-specific implementation
- **[Service Example](/app/svcexample)**: Advanced service pattern usage

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Roadmap

- [ ] MQTT v5 support
- [ ] Kafka client implementation
- [ ] Observability (OpenTelemetry integration)
- [ ] Schema validation
- [ ] Message replay/dead letter queue

## Related Projects

- [NATS](https://nats.io/) - Cloud-native messaging system
- [Eclipse Paho](https://www.eclipse.org/paho/) - MQTT client library
