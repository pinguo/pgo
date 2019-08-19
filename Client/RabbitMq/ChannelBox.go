package RabbitMq

import (
    "time"

    "github.com/pinguo/pgo"
    "github.com/streadway/amqp"
)

func newChannelBox(connBox *ConnBox, pool *Pool) *ChannelBox {
    connBox.newChannelLock.Lock()
    defer connBox.newChannelLock.Unlock()
    channel, err := connBox.connection.Channel()

    if err != nil {
        panic("Rabbit newChannelBox err:" + err.Error())
    }
    connBox.useConnCount++
    return &ChannelBox{connBoxId: connBox.id, pool: pool, channel: channel, connStartTime: connBox.startTime, lastActive: time.Now()}
}

type ChannelBox struct {
    connBoxId     string
    pool          *Pool
    channel       *amqp.Channel
    connStartTime time.Time
    lastActive    time.Time
}

func (c *ChannelBox) Close(force bool) {
    if force || c.connStartTime != c.pool.getConnBox(c.connBoxId).startTime {
        c.channelClose()
        return
    }

    if !c.pool.putFreeChannel(c) {
        c.channelClose()
    } else {
        c.lastActive = time.Now()
    }
}

func (c *ChannelBox) channelClose() {
    connBox := c.pool.getConnBox(c.connBoxId)
    connBox.useConnCount--
    if connBox.isClosed() == false{
        err := c.channel.Close()
        if err != nil {
            pgo.GLogger().Warn("Rabbit ChannelBox.channelClose.channel.Close() err:" + err.Error())
        }
    }
}

func (c *ChannelBox) GetChannel() *amqp.Channel {
    return c.channel
}
