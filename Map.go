package pgo

import (
    "encoding/json"
)

// Map 兼容当前实现
type Map map[string]interface{}

// ToStringMap 转换成 StringMap 类型
func (m Map) ToStringMap() StringMap {
    return StringMap(m)
}

// StringMap 以 string 为 key 的map
type StringMap map[string]interface{}

// NewStringMap 创建 StringMap，分配容量100
func NewStringMap() StringMap {
    return make(StringMap, 100)
}

// Add 添加一个元素，当 key 已存在时返回 false
func (sm StringMap) Add(key string, v interface{}) bool {
    if _, ok := sm[key]; ok {
        return false
    }
    sm[key] = v
    return true
}

// Set 添加一个元素，若 key 已存在则覆盖
func (sm StringMap) Set(key string, v interface{}) {
    sm[key] = v
}

// Has key 是否存在
func (sm StringMap) Has(key string) bool {
    _, ok := sm[key]
    return ok
}

// Remove 移出元素
func (sm StringMap) Remove(keys ...string) {
    for _, k := range keys {
        delete(sm, k)
    }
}

// Len map长度
func (sm StringMap) Len() int {
    return len(sm)
}

// Keys 获取map全部key
func (sm StringMap) Keys() []string {
    keys := make([]string, 0, len(sm))
    for k := range sm {
        keys = append(keys, k)
    }
    return keys
}

// Clear 清空map
func (sm StringMap) Clear() {
    sm = nil
}

// ToJSON 转换成 json
func (sm StringMap) ToJSON() ([]byte, error) {
    return json.Marshal(sm)
}
