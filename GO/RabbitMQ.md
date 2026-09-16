---
id: rabbitmq_go
aliases:
  - RabbitMQ in Go
  - AMQP in Go
tags:
  - go
  - rabbitmq
  - amqp
  - message-queues
  - event-driven
dg-publish: true
---

# RabbitMQ in Go: Exchanges, Queues & Workers

**RabbitMQ** is a robust, widely adopted AMQP message broker. In Go, the official active library is **`github.com/rabbitmq/amqp091-go`** (superseding the deprecated `streadway/amqp`).

```
+----------+             +------------+  Routing Key   +-------+           +----------+
| Producer | ──Publish─> |  Exchange  | ─────────────> | Queue | ──Consume─>| Consumer |
+----------+             +------------+  ("orders.new") +-------+           +----------+
                               │                               ▲
                               └───────── Binding ─────────────┘
```

---

## 1. Core Concepts & Exchange Types

| Exchange Type | Routing Behavior | Typical Use Case |
|---|---|---|
| **Direct** | Exact match between routing key and binding key | Task queues, 1-to-1 routing |
| **Fanout** | Broadcasts blindly to all queues bound to it | Notification systems, cache invalidation |
| **Topic** | Pattern matching using `*` (1 word) and `#` (0 or more words) | Pub/sub routing (`logs.error.auth`, `orders.*`) |
| **Headers** | Routes messages based on header attributes rather than routing keys | Complex multi-attribute routing |

---

## 2. Driver Installation

```bash
go get github.com/rabbitmq/amqp091-go
```

---

## 3. Publisher Implementation

The producer establishes a TCP connection, opens an AMQP channel, ensures the queue exists, and publishes a persistent message.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQProducer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewProducer(amqpURI string) (*RabbitMQProducer, error) {
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQProducer{conn: conn, ch: ch}, nil
}

func (p *RabbitMQProducer) PublishTask(ctx context.Context, queueName string, body []byte) error {
	// Declare queue (idempotent: created if not present)
	q, err := p.ch.QueueDeclare(
		queueName, // name
		true,      // durable (survives broker restart)
		false,     // delete when unused
		false,     // exclusive (used by only one connection)
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("queue declare error: %w", err)
	}

	// Publish message with delivery mode = Persistent
	err = p.ch.PublishWithContext(ctx,
		"",     // exchange (empty string = default direct exchange)
		q.Name, // routing key (queue name)
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // message durability
			ContentType:  "application/json",
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish error: %w", err)
	}

	fmt.Printf("Sent message to queue '%s': %s\n", queueName, string(body))
	return nil
}

func (p *RabbitMQProducer) Close() {
	if p.ch != nil {
		p.ch.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
```

---

## 4. Consumer Implementation (Worker with Manual Ack)

> [!IMPORTANT]
> Always set `autoAck: false` and call `d.Ack(false)` once processing finishes. If the consumer crashes midway, RabbitMQ requeues the message for another worker.

```go
package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func StartConsumer(amqpURI string, queueName string) {
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		log.Fatalf("Consumer connection error: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Consumer channel error: %v", err)
	}
	defer ch.Close()

	// Ensure queue is declared
	_, err = ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("Queue declaration error: %v", err)
	}

	// Fair dispatch: don't give more than 1 message to a worker until it's acked
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Fatalf("Qos configuration failed: %v", err)
	}

	msgs, err := ch.Consume(
		queueName,
		"",    // consumer tag (auto-generated)
		false, // auto-ack (FALSE: manual acknowledgment required!)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Fatalf("Consume failed: %v", err)
	}

	fmt.Printf(" [*] Waiting for messages in queue '%s'. To exit press CTRL+C\n", queueName)

	forever := make(chan struct{})

	go func() {
		for d := range msgs {
			fmt.Printf("Received delivery [%s]: %s\n", d.MessageId, string(d.Body))

			// Simulate processing logic
			err := processTask(d.Body)
			if err != nil {
				log.Printf("Task processing failed: %v. Requeueing...", err)
				// Requeue message (or send to Dead Letter Exchange)
				_ = d.Nack(false, true)
				continue
			}

			// Positively acknowledge message was processed
			_ = d.Ack(false)
		}
	}()

	<-forever
}

func processTask(data []byte) error {
	// Custom business logic here
	return nil
}
```

---

## 5. Topic Exchange Pattern (Wildcard Pub/Sub)

```go
// Setup topic exchange: "events_topic"
err := ch.ExchangeDeclare(
    "events_topic", // name
    "topic",        // type
    true,           // durable
    false,          // auto-deleted
    false,          // internal
    false,          // no-wait
    nil,            // args
)

// Binding queue with wildcard key:
// "audit.*" captures "audit.login", "audit.logout"
// "audit.#" captures "audit.user.created.success"
err = ch.QueueBind(
    q.Name,
    "audit.*",
    "events_topic",
    false,
    nil,
)
```
