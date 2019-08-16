package RabbitMq

import (
    "github.com/pinguo/pgo"
    "github.com/pinguo/pgo/Util"
    "github.com/streadway/amqp"
)

type Adapter struct {
    pgo.Object
    client       *Client
    panicRecover bool
}

func (a *Adapter) Construct(componentId ...string) {
    id := defaultComponentId
    if len(componentId) > 0 {
        id = componentId[0]
    }

    a.client = pgo.App.Get(id).(*Client)
    a.panicRecover = true
}

func (a *Adapter) SetPanicRecover(v bool) {
    a.panicRecover = v
}

func (a *Adapter) GetClient() *Client {
    return a.client
}

func (a *Adapter) handlePanic() {
    if a.panicRecover {
      if v := recover(); v != nil {
          a.GetContext().Error(Util.ToString(v))
      }
    }
}

func (a *Adapter) ExchangeDeclare() {
    profile := "rabbit.ExchangeDeclare"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()
    a.client.setExchangeDeclare()
}

func (a *Adapter) Publish(opCode string, data interface{}, dftOpUid ...string) bool {
    profile := "rabbit.Publish"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    opUid := ""
    if len(dftOpUid) > 0 {
        opUid = dftOpUid[0]
    }

    return a.client.publish(&PublishData{OpCode: opCode, Data: data, OpUid: opUid}, a.GetContext().GetLogId())
}

func (a *Adapter) GetConsumeChannelBox(queueName string, opCodes []string) *ChannelBox {
    profile := "rabbit.GetConsumeChannelBox"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.getConsumeChannelBox(queueName, opCodes)
}

// 消费，返回chan 可以不停取数据
// queueName 队列名字
// opCodes 绑定队列的code
// limit 每次接收多少条
// autoAck 是否自动答复 如果为false 需要手动调用Delivery.ack(false)
// noWait 是否一直等待
// 是否独占队列
func (a *Adapter) Consume(queueName string, opCodes []string, limit int, autoAck, noWait, exclusive bool) <-chan amqp.Delivery {
    profile := "rabbit.Consume"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.consume(&ConsumeData{QueueName: queueName, OpCodes: opCodes, Limit: limit, AutoAck: autoAck, NoWait: noWait, Exclusive: exclusive})
}

func (a *Adapter) DecodeBody(d amqp.Delivery, ret interface{}) error {
    return a.client.decodeBody(d,ret)
}

func (a *Adapter) DecodeHeaders(d amqp.Delivery) *RabbitHeaders {
    return a.client.decodeHeaders(d)
}