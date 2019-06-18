package pgo

import "time"

type IBind interface {
    GetBindInfo(v interface{}) interface{}
}

type IObject interface {
    SetContext(ctx *Context)
    GetContext() *Context
    GetObject(class interface{}, params ...interface{}) interface{}
}

type IController interface {
    BeforeAction(action string)
    AfterAction(action string)
    HandlePanic(v interface{})
}

type IPlugin interface {
    HandleRequest(ctx *Context)
}

type IEvent interface {
    HandleEvent(event string, ctx *Context, args ...interface{})
}

type IFormatter interface {
    Format(item *LogItem) string
}

type ITarget interface {
    Process(item *LogItem)
    Flush(final bool)
}

type IConfigParser interface {
    Parse(path string) map[string]interface{}
}

type ICache interface {
    Get(key string) *Value
    MGet(keys []string) map[string]*Value
    Set(key string, value interface{}, expire ...time.Duration) bool
    MSet(items map[string]interface{}, expire ...time.Duration) bool
    Add(key string, value interface{}, expire ...time.Duration) bool
    MAdd(items map[string]interface{}, expire ...time.Duration) bool
    Del(key string) bool
    MDel(keys []string) bool
    Exists(key string) bool
    Incr(key string, delta int) int
}
