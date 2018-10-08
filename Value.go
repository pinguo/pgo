package pgo

import (
    "encoding/json"
    "errors"
    "fmt"
    "reflect"
    "strconv"

    "github.com/pinguo/pgo/Util"
)

func NewValue(data interface{}) *Value {
    return &Value{data}
}

func Encode(data interface{}) []byte {
    v := Value{data}
    return v.Encode()
}

func Decode(data interface{}, ptr interface{}) {
    v := Value{data}
    v.Decode(ptr)
}

type Value struct {
    data interface{}
}

func (v *Value) Valid() bool {
    return v.data != nil
}

func (v *Value) TryEncode() (output []byte, err error) {
    defer func() {
        if v := recover(); v != nil {
            output, err = nil, errors.New(Util.ToString(v))
        }
    }()
    return v.Encode(), nil
}

func (v *Value) TryDecode(ptr interface{}) (err error) {
    defer func() {
        if v := recover(); v != nil {
            err = errors.New(Util.ToString(v))
        }
    }()
    v.Decode(ptr)
    return nil
}

func (v *Value) Encode() []byte {
    var output []byte
    switch d := v.data.(type) {
    case []byte:
        output = d
    case string:
        output = []byte(d)
    case bool:
        output = strconv.AppendBool(output, d)
    case float32, float64:
        f64 := reflect.ValueOf(v.data).Float()
        output = strconv.AppendFloat(output, f64, 'g', -1, 64)
    case int, int8, int16, int32, int64:
        i64 := reflect.ValueOf(v.data).Int()
        output = strconv.AppendInt(output, i64, 10)
    case uint, uint8, uint16, uint32, uint64:
        u64 := reflect.ValueOf(v.data).Uint()
        output = strconv.AppendUint(output, u64, 10)
    default:
        if j, e := json.Marshal(v.data); e == nil {
            output = j
        } else {
            panic("Value.Encode: " + e.Error())
        }
    }
    return output
}

func (v *Value) Decode(ptr interface{}) {
    switch p := ptr.(type) {
    case *[]byte:
        *p = v.Bytes()
    case *string:
        *p = v.String()
    case *bool:
        *p = Util.ToBool(v.data)
    case *float32, *float64:
        fv := Util.ToFloat(v.data)
        rv := reflect.ValueOf(ptr).Elem()
        rv.Set(reflect.ValueOf(fv).Convert(rv.Type()))
    case *int, *int8, *int16, *int32, *int64:
        iv := Util.ToInt(v.data)
        rv := reflect.ValueOf(ptr).Elem()
        rv.Set(reflect.ValueOf(iv).Convert(rv.Type()))
    case *uint, *uint8, *uint16, *uint32, *uint64:
        iv := Util.ToInt(v.data)
        rv := reflect.ValueOf(ptr).Elem()
        rv.Set(reflect.ValueOf(iv).Convert(rv.Type()))
    default:
        if e := json.Unmarshal(v.Bytes(), ptr); e != nil {
            rv := reflect.ValueOf(ptr)
            if rv.Kind() != reflect.Ptr || rv.IsNil() {
                panic("Value.Decode: require a valid pointer")
            }

            if rv = rv.Elem(); rv.Kind() == reflect.Interface {
                rv.Set(reflect.ValueOf(v.data))
            } else {
                panic("Value.Decode: " + e.Error())
            }
        }
    }
}

func (v *Value) Data() interface{} {
    return v.data
}

func (v *Value) Bool() bool {
    return Util.ToBool(v.data)
}

func (v *Value) Int() int {
    return Util.ToInt(v.data)
}

func (v *Value) Float() float64 {
    return Util.ToFloat(v.data)
}

func (v *Value) String() string {
    switch d := v.data.(type) {
    case []byte:
        return string(d)
    case string:
        return d
    default:
        if j, e := json.Marshal(v.data); e == nil {
            return string(j)
        }
        return fmt.Sprintf("%+v", v.data)
    }
}

func (v *Value) Bytes() []byte {
    switch d := v.data.(type) {
    case []byte:
        return d
    case string:
        return []byte(d)
    default:
        if j, e := json.Marshal(v.data); e == nil {
            return j
        }
        return []byte(fmt.Sprintf("%+v", v.data))
    }
}

func (v *Value) MarshalJSON() ([]byte, error) {
    return json.Marshal(v.String())
}

func (v *Value) UnmarshalJSON(b []byte) error {
    var s string
    if e := json.Unmarshal(b, &s); e != nil {
        return e
    }
    v.data = s
    return nil
}
