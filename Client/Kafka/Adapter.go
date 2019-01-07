package Kafka

import (
    "github.com/pinguo/pgo"
)

// Adapter of Kafka Client, add context support.
// usage: kafka := this.GetObject(Kafka.AdapterClass).(*Kafka.Adapter)
type Adapter struct {
    pgo.Object
    client *Client
}

func (a *Adapter) Construct(componentId ...string) {
    id := defaultComponentId
    if len(componentId) > 0 {
        id = componentId[0]
    }

    a.client = pgo.App.Get(id).(*Client)
}

func (a *Adapter) GetClient() *Client {
    return a.client
}
