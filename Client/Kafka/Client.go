package Kafka

import (
    "github.com/Shopify/sarama"
)

// Kafka client component, require kafka>=0.9, configuration:
// kafka:
//     class:  "@pgo/Client/Kafka/Client"
type Client struct {
}

func (c *Client) Construct() {
    _ = sarama.NewConfig()
}
