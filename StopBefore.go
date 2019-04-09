package pgo

import (
    "reflect"
)

type StopBefore struct {
    performQueue []*Item
}

type Item struct {
    Action reflect.Value
    Params []reflect.Value
}

// 增加停止前执行的对象和方法
func (s *StopBefore) Add(obj interface{}, method string, dfParams ...[]interface{}) {

    if len(s.performQueue) > 10 {
        panic("The length of the performQueue cannot be greater than 10")
    }
    var params = make([]reflect.Value, 0)
    if len(dfParams) > 0 {
        for _, v := range dfParams[0] {
            params = append(params, reflect.ValueOf(v))
        }
    }
    action := reflect.ValueOf(obj).MethodByName(method)

    if action.IsValid() == false {
        panic("err obj or method")
    }
    s.performQueue = append(s.performQueue, &Item{Action: action, Params: params})
}

// 执行
func (s *StopBefore) Exec() {
    for _, item := range s.performQueue {
        item.Action.Call(item.Params)
    }
}
