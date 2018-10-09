package Util

import (
    "encoding/json"
    "fmt"
    "reflect"
    "strconv"
    "strings"
)

func ToBool(v interface{}) bool {
    switch val := v.(type) {
    case bool:
        return val
    case float32, float64:
        // direct type conversion may cause data loss, use reflection instead
        return reflect.ValueOf(v).Float() != 0
    case int, int8, int16, int32, int64:
        return reflect.ValueOf(v).Int() != 0
    case uint, uint8, uint16, uint32, uint64:
        return reflect.ValueOf(v).Uint() != 0
    case string:
        return str2bool(val)
    case []byte:
        return str2bool(string(val))
    default:
        rv := reflect.ValueOf(v)
        if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
            rv = rv.Elem()
        }

        // none empty array/slice/map convert to true, otherwise false
        if rv.Kind() == reflect.Array ||
            rv.Kind() == reflect.Slice ||
            rv.Kind() == reflect.Map {
            return rv.Len() != 0
        }

        // valid value convert to true, otherwise false
        return rv.IsValid()
    }
}

func ToInt(v interface{}) int {
    switch val := v.(type) {
    case bool:
        if val {
            return 1
        } else {
            return 0
        }
    case float32, float64:
        // direct type conversion may cause data loss, use reflection instead
        return int(reflect.ValueOf(v).Float())
    case int, int8, int16, int32, int64:
        return int(reflect.ValueOf(v).Int())
    case uint, uint8, uint16, uint32, uint64:
        return int(reflect.ValueOf(v).Uint())
    case string:
        return str2int(val)
    case []byte:
        return str2int(string(val))
    case nil:
        return 0
    default:
        panic(fmt.Sprintf("ToInt: invalid type: %T", v))
    }
}

func ToFloat(v interface{}) float64 {
    switch val := v.(type) {
    case bool:
        if val {
            return 1
        } else {
            return 0
        }
    case float32, float64:
        // direct type conversion may cause data loss, use reflection instead
        return reflect.ValueOf(v).Float()
    case int, int8, int16, int32, int64:
        return float64(reflect.ValueOf(v).Int())
    case uint, uint8, uint16, uint32, uint64:
        return float64(reflect.ValueOf(v).Uint())
    case string:
        return str2float(val)
    case []byte:
        return str2float(string(val))
    case nil:
        return 0
    default:
        panic(fmt.Sprintf("ToFloat: invalid type: %T", v))
    }
}

func ToString(v interface{}) string {
    switch val := v.(type) {
    case bool:
        return strconv.FormatBool(val)
    case int, int8, int16, int32, int64:
        return strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
    case uint, uint8, uint16, uint32, uint64:
        return strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
    case float32, float64:
        return strconv.FormatFloat(reflect.ValueOf(v).Float(), 'g', -1, 64)
    case []byte:
        return string(val)
    case string:
        return val
    case error:
        return val.Error()
    case fmt.Stringer:
        return val.String()
    default:
        // convert to json encoded string
        if j, e := json.Marshal(v); e == nil {
            return string(j)
        }

        // convert to default print string
        return fmt.Sprintf("%+v", v)
    }
}

func str2bool(s string) bool {
    s = strings.TrimSpace(s)
    if b, e := strconv.ParseBool(s); e == nil {
        return b
    }
    return len(s) != 0
}

func str2int(s string) int {
    s = strings.TrimSpace(s)
    if i64, e := strconv.ParseInt(s, 0, 0); e == nil {
        // convert int string(decimal, hexadecimal, octal)
        return int(i64)
    } else if f64, e := strconv.ParseFloat(s, 64); e == nil {
        // convert float string
        return int(f64)
    } else {
        return 0
    }
}

func str2float(s string) float64 {
    s = strings.TrimSpace(s)
    if f64, e := strconv.ParseFloat(s, 64); e == nil {
        // convert float string
        return f64
    } else if i64, e := strconv.ParseInt(s, 0, 0); e == nil {
        // convert int string(decimal, hexadecimal, octal)
        return float64(i64)
    } else {
        return 0
    }
}

func ToBytes(arg interface{}) (string, error) {
    switch arg := arg.(type) {
    case int:
        return strconv.Itoa(arg), nil
    case int64:
        return strconv.Itoa(int(arg)), nil
    case float64:
        return strconv.FormatFloat(arg, 'g', -1, 64), nil
    case string:
        return arg, nil
    case []byte:
        return string(arg), nil
    default:
        return "", fmt.Errorf("unknown type %T", arg)
    }
}
