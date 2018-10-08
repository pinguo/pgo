package Util

import (
	"fmt"
	"strings"
)

// clear to empty
func MapClear(m map[string]interface{}) {
	for k := range m {
		delete(m, k)
	}
}

// merge map recursively
func MapMerge(a, b map[string]interface{}) {
	for k := range b {
		va, oa := a[k].(map[string]interface{})
		vb, ob := b[k].(map[string]interface{})

		if oa && ob {
			MapMerge(va, vb)
		} else {
			a[k] = b[k]
		}
	}
}

// get value by dot separated key, empty key for m itself
func MapGet(m map[string]interface{}, key string) interface{} {
	var data interface{} = m
	ks := strings.Split(key, ".")

	for _, k := range ks {
		// skip empty key segment
		if k = strings.TrimSpace(k); len(k) == 0 {
			continue
		}

		if v, ok := data.(map[string]interface{}); ok {
			if data, ok = v[k]; ok {
				continue
			}
		}

		// not found
		return nil
	}

	return data
}

// set value by dot separated key, empty key for root, nil val for clear
func MapSet(m map[string]interface{}, key string, val interface{}) {
	data := m
	last := ""

	ks := strings.Split(key, ".")
	for _, k := range ks {
		// skip empty key segment
		if k = strings.TrimSpace(k); len(k) == 0 {
			continue
		}

		if len(last) > 0 {
			if _, ok := data[last].(map[string]interface{}); !ok {
				data[last] = make(map[string]interface{})
			}

			data = data[last].(map[string]interface{})
		}

		last = k
	}

	if len(last) > 0 {
		if nil == val {
			delete(data, last)
		} else {
			data[last] = val
		}
	} else {
		MapClear(m)
		if v, ok := val.(map[string]interface{}); ok {
			MapMerge(m, v)
		} else if nil != val {
			panic(fmt.Sprintf("MapSet: invalid type: %T", val))
		}
	}
}
