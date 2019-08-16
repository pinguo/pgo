package RabbitMq

import (
    "bytes"
    "encoding/gob"
    "time"

    "github.com/streadway/amqp"
)

// RabbitMq client component,
// support Publisher-Consumer  configuration:
// rabbitMq:
//     class: "@pgo/Client/RabbitMq/Client"
//     tlsRootCAs:""
//     tlsCert: ""
//     tlsCertKey: ""
//     user: "guest"
//     pass: "guest"
//     exchangeName: ""
//     exchangeType: ""
//     maxChannelNum: 2000
//     maxIdleChannel: "200"
//     maxIdleChannelTime:"10s"
//     probeInterval: "0s"
//     maxWaitTime: "200ms"
//     serverName: "pgo-xxx"
//     servers:
//         - "127.0.0.1:6379"
//         - "127.0.0.1:6380"
type Client struct {
    Pool
}

func (c *Client) decodeBody(d amqp.Delivery, ret interface{}) error {
    var network *bytes.Buffer
    network = bytes.NewBuffer(d.Body)
    dec := gob.NewDecoder(network)
    err := dec.Decode(ret)

    return err
}

func (c *Client) decodeHeaders(d amqp.Delivery) *RabbitHeaders {
    ret := &RabbitHeaders{
        Exchange:  d.Exchange,
        RouteKey:  d.RoutingKey,
        Timestamp: d.Timestamp,
        MessageId: d.MessageId,
    }

    for k, iV := range d.Headers {
        v, _ := iV.(string)
        switch k {
        case "logId":
            ret.LogId = v
        case "service":
            ret.Service = v
        case "opUid":
            ret.OpUid = v
        }
    }

    return ret
}

func (c *Client) setExchangeDeclare() {
    ch := c.getFreeChannel()
    defer ch.Close(false)
    c.exchangeDeclare(ch)
}

func (c *Client) publish(parameter *PublishData, logId string) bool{
    if parameter.OpCode == "" || parameter.Data == nil {
        panic("Rabbit OpCode and LogId cannot be empty")
    }

    ch := c.getFreeChannel()
    defer ch.Close(false)

    // 增加速度，在消费端定义交换机 或者单独定义交换机
    // c.exchangeDeclare(ch)

    var goBytes bytes.Buffer
    myGob := gob.NewEncoder(&goBytes)
    err := myGob.Encode(parameter.Data)
    c.failOnError(err, "Encode err")

    err = ch.channel.Publish(
        c.getExchangeName(),             // exchange
        c.getRouteKey(parameter.OpCode), // routing key
        false,                           // mandatory
        false,                           // immediate
        amqp.Publishing{
            ContentType: "text/plain",
            Body:        goBytes.Bytes(),
            Headers:     amqp.Table{"logId": logId, "service": c.ServiceName, "opUid": parameter.OpUid},
            Timestamp:   time.Now(),
        })
    c.failOnError(err, "Failed to publish a message")
    return true
}

// 定义交换机
func (c *Client) exchangeDeclare(ch *ChannelBox) bool {
    err := ch.channel.ExchangeDeclare(
        c.getExchangeName(), // name
        c.exchangeType,      // type
        true,                // durable
        false,               // auto-deleted
        false,               // internal
        false,               // no-wait
        nil,                 // arguments
    )
    c.failOnError(err, "Failed to declare an exchange")

    return true
}

// 定义交换机
func (c *Client) bindQueue(ch *ChannelBox, queueName string, opCodes []string) bool {
    for _, opCode := range opCodes {
        err := ch.channel.QueueBind(
            queueName,             // queue name
            c.getRouteKey(opCode), // routing key
            c.getExchangeName(),   // exchange
            false,
            nil)

        c.failOnError(err, "Failed to bind a queue")
    }

    return true
}

func (c *Client) queueDeclare(ch *ChannelBox, queueName string) amqp.Queue {
    q, err := ch.channel.QueueDeclare(
        queueName, // name
        true,      // durable
        false,     // delete when usused
        false,     // exclusive
        false,     // no-wait
        nil,       // arguments
    )

    if err != nil {
        c.failOnError(err, "Failed to declare a queue")
    }

    return q
}

func (c *Client) getConsumeChannelBox(queueName string, opCodes []string) *ChannelBox {
    ch := c.getFreeChannel()
    // 定义交换机
    c.exchangeDeclare(ch)
    // 定义queue
    c.queueDeclare(ch, queueName)
    // 绑定queue
    c.bindQueue(ch, queueName, opCodes)

    return ch

}

func (c *Client) consume(parameter *ConsumeData) <-chan amqp.Delivery {
    ch := c.getConsumeChannelBox(parameter.QueueName, parameter.OpCodes)
    // defer ch.Close(false)
    err := ch.channel.Qos(parameter.Limit, 0, false)
    c.failOnError(err, "set Qos err")

    messages, err := ch.channel.Consume(
        parameter.QueueName, // queue
        "",                  // consumer
        parameter.AutoAck,   // auto ack
        parameter.Exclusive, // exclusive
        false,               // no local
        parameter.NoWait,    // no wait
        nil,                 // args
    )
    c.failOnError(err, "get msg err")

    return messages
}

func (c *Client) failOnError(err error, msg string) {
    if err != nil {
        panic("Rabbit:" + msg + ",err:" + err.Error())
    }
}
