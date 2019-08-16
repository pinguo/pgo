package RabbitMq

import (
    "time"

    "github.com/pinguo/pgo"
)

const (
    AdapterClass = "@pgo/Client/RabbitMq/Adapter"
    defaultComponentId = "rabbitMq"
    dftExchangeType = "direct"
    dftExchangeName = "direct_pgo_dft"
    dftMaxChannelNum      = 2000
    dftMaxIdleChannel     = 200
    dftMaxIdleChannelTime = 60 * time.Second
    dftMaxWaitTime        = 200 * time.Microsecond
    dftProbeInterval      = 0
    dftProtocol = "amqp"
    defaultTimeout = 1 * time.Second
    errSetProp     = "rabbitMq: failed to set %s, %s"
)

func init() {
    container := pgo.App.GetContainer()

    container.Bind(&Client{})
    container.Bind(&Adapter{})
}

type RabbitHeaders struct {
    LogId string
    Exchange string
    RouteKey string
    Service string
    OpUid string
    Timestamp time.Time
    MessageId string
}

// rabbit 发布结构
type PublishData struct {
    OpCode  string // 操作code 和queue绑定相关
    OpUid   string // 操作用户id 可以为空
    Data    interface{} // 发送数据
}

type ConsumeData struct {
    QueueName string
    OpCodes   []string
    AutoAck   bool
    NoWait    bool
    Exclusive bool
    Limit     int
}
