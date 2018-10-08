package pgo

import (
    "encoding/json"
    "net/http"
    "regexp"
    "strings"
    "unicode"
    "unicode/utf8"

    "github.com/pinguo/pgo/Util"
)

var (
    emailRe  = regexp.MustCompile(`(?i)^[a-z0-9_-]+@[a-z0-9_-]+(\.[a-z0-9_-]+)+$`)
    mobileRe = regexp.MustCompile(`^(\+\d{2,3} )?1[35789]\d{9}$`)
    ipv4Re   = regexp.MustCompile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$`)
)

// validate bool value
func ValidateBool(data interface{}, name string, dft ...interface{}) *BoolValidator {
    value, useDft := getValidateValue(data, name, dft...)
    return &BoolValidator{name, useDft, Util.ToBool(value)}
}

// validate int value
func ValidateInt(data interface{}, name string, dft ...interface{}) *IntValidator {
    value, useDft := getValidateValue(data, name, dft...)
    return &IntValidator{name, useDft, Util.ToInt(value)}
}

// validate float value
func ValidateFloat(data interface{}, name string, dft ...interface{}) *FloatValidator {
    value, useDft := getValidateValue(data, name, dft...)
    return &FloatValidator{name, useDft, Util.ToFloat(value)}
}

// validate string value
func ValidateString(data interface{}, name string, dft ...interface{}) *StringValidator {
    value, useDft := getValidateValue(data, name, dft...)
    return &StringValidator{name, useDft, Util.ToString(value)}
}

// get validate value, four situations:
// 1. data: map, name: field, dft[0]: default
// 2. data: map, name: field, dft: empty
// 3. data: value, name: field, dft[0]: default
// 4. data: value, name: field, dft: empty
func getValidateValue(data interface{}, name string, dft ...interface{}) (interface{}, bool) {
    var value interface{}
    var useDft = false

    switch v := data.(type) {
    case map[string]interface{}:
        if mv, ok := v[name]; ok {
            value = mv
        }
    case map[string]string:
        if mv, ok := v[name]; ok {
            value = mv
        }
    case map[string][]string:
        sliceValue, sliceOk := v[name]
        if sliceOk && len(sliceValue) > 0 {
            value = sliceValue[0]
        }
    default:
        value = data
    }

    if value == nil {
        if len(dft) == 1 {
            value = dft[0]
            useDft = true
        } else {
            panic(NewException(http.StatusBadRequest, "%s is required", name))
        }
    } else if strValue, strOk := value.(string); strOk {
        strValue = strings.Trim(strValue, " \r\n\t")
        if len(strValue) > 0 {
            value = strValue
        } else if len(dft) == 1 {
            value = dft[0]
            useDft = true
        } else {
            panic(NewException(http.StatusBadRequest, "%s can't be empty", name))
        }
    }

    return value, useDft
}

// validator for bool value
type BoolValidator struct {
    Name   string
    UseDft bool
    Value  bool
}

func (b *BoolValidator) Must(v bool) *BoolValidator {
    if !b.UseDft && b.Value != v {
        panic(NewException(http.StatusBadRequest, "%s must be %v", b.Name, v))
    }
    return b
}

func (b *BoolValidator) Do() bool {
    return b.Value
}

// validator for int value
type IntValidator struct {
    Name   string
    UseDft bool
    Value  int
}

func (i *IntValidator) Min(v int) *IntValidator {
    if !i.UseDft && i.Value < v {
        panic(NewException(http.StatusBadRequest, "%s is too small", i.Name))
    }
    return i
}

func (i *IntValidator) Max(v int) *IntValidator {
    if !i.UseDft && i.Value > v {
        panic(NewException(http.StatusBadRequest, "%s is too large", i.Name))
    }
    return i
}

func (i *IntValidator) Enum(enums ...int) *IntValidator {
    found := false
    for _, v := range enums {
        if v == i.Value {
            found = true
            break
        }
    }

    if !i.UseDft && !found {
        panic(NewException(http.StatusBadRequest, "%s is invalid", i.Name))
    }
    return i
}

func (i *IntValidator) Do() int {
    return i.Value
}

// validator for float value
type FloatValidator struct {
    Name   string
    UseDft bool
    Value  float64
}

func (f *FloatValidator) Min(v float64) *FloatValidator {
    if !f.UseDft && f.Value < v {
        panic(NewException(http.StatusBadRequest, "%s is too small", f.Name))
    }
    return f
}

func (f *FloatValidator) Max(v float64) *FloatValidator {
    if !f.UseDft && f.Value > v {
        panic(NewException(http.StatusBadRequest, "%s is too large", f.Name))
    }
    return f
}

func (f *FloatValidator) Do() float64 {
    return f.Value
}

// validator for string value
type StringValidator struct {
    Name   string
    UseDft bool
    Value  string
}

func (s *StringValidator) Min(v int) *StringValidator {
    if !s.UseDft && utf8.RuneCountInString(s.Value) < v {
        panic(NewException(http.StatusBadRequest, "%s is too short", s.Name))
    }
    return s
}

func (s *StringValidator) Max(v int) *StringValidator {
    if !s.UseDft && utf8.RuneCountInString(s.Value) > v {
        panic(NewException(http.StatusBadRequest, "%s is too long", s.Name))
    }
    return s
}

func (s *StringValidator) Len(v int) *StringValidator {
    if !s.UseDft && utf8.RuneCountInString(s.Value) != v {
        panic(NewException(http.StatusBadRequest, "%s has invalid length", s.Name))
    }
    return s
}

func (s *StringValidator) Enum(enums ...string) *StringValidator {
    found := false
    for _, v := range enums {
        if v == s.Value {
            found = true
            break
        }
    }

    if !s.UseDft && !found {
        panic(NewException(http.StatusBadRequest, "%s is invalid", s.Name))
    }
    return s
}

func (s *StringValidator) RegExp(v interface{}) *StringValidator {
    var re *regexp.Regexp
    if pat, ok := v.(string); ok {
        re = regexp.MustCompile(pat)
    } else {
        re = v.(*regexp.Regexp)
    }

    if !s.UseDft && !re.MatchString(s.Value) {
        panic(NewException(http.StatusBadRequest, "%s is invalid", s.Name))
    }

    return s
}

func (s *StringValidator) Filter(f func(v, n string) string) *StringValidator {
    defer func() {
        if v := recover(); !s.UseDft && v != nil {
            panic(NewException(http.StatusBadRequest, "%s is invalid", s.Name))
        }
    }()

    if v := f(s.Value, s.Name); len(v) > 0 {
        s.Value = v
    } else if !s.UseDft {
        panic(NewException(http.StatusBadRequest, "%s is invalid", s.Name))
    }

    return s
}

func (s *StringValidator) Password() *StringValidator {
    length, number, letter, special := false, false, false, false

    if l := len(s.Value); 6 <= l && l <= 32 {
        length = true
        for i := 0; i < l; i++ {
            switch {
            case unicode.IsNumber(rune(s.Value[i])):
                number = true
            case unicode.IsLetter(rune(s.Value[i])):
                letter = true
            case unicode.IsPunct(rune(s.Value[i])):
                special = true
            case unicode.IsSymbol(rune(s.Value[i])):
                special = true
            }
        }
    }

    if !s.UseDft && (!length || !number || !letter || !special) {
        panic(NewException(http.StatusBadRequest, "%s is invalid password", s.Name))
    }

    return s
}

func (s *StringValidator) Email() *StringValidator {
    if !s.UseDft && !emailRe.MatchString(s.Value) {
        panic(NewException(http.StatusBadRequest, "%s is invalid email", s.Name))
    }

    return s
}

func (s *StringValidator) Mobile() *StringValidator {
    if !s.UseDft && !mobileRe.MatchString(s.Value) {
        panic(NewException(http.StatusBadRequest, "%s is invalid mobile", s.Name))
    }

    return s
}

func (s *StringValidator) IPv4() *StringValidator {
    if !s.UseDft && !ipv4Re.MatchString(s.Value) {
        panic(NewException(http.StatusBadRequest, "%s is invalid ipv4", s.Name))
    }

    return s
}

func (s *StringValidator) Bool() *BoolValidator {
    return &BoolValidator{s.Name, s.UseDft, Util.ToBool(s.Value)}
}

func (s *StringValidator) Int() *IntValidator {
    return &IntValidator{s.Name, s.UseDft, Util.ToInt(s.Value)}
}

func (s *StringValidator) Float() *FloatValidator {
    return &FloatValidator{s.Name, s.UseDft, Util.ToFloat(s.Value)}
}

func (s *StringValidator) Slice(sep string) *StringSliceValidator {
    validator := &StringSliceValidator{s.Name, s.UseDft, make([]string, 0)}

    if len(s.Value) > 0 {
        parts := strings.Split(s.Value, sep)
        for _, v := range parts {
            validator.Value = append(validator.Value, strings.TrimSpace(v))
        }
    }

    return validator
}

func (s *StringValidator) Json() *JsonValidator {
    validator := &JsonValidator{s.Name, s.UseDft, make(map[string]interface{})}
    decoder := json.NewDecoder(strings.NewReader(s.Value))
    if err := decoder.Decode(&validator.Value); !s.UseDft && err != nil {
        panic(NewException(http.StatusBadRequest, "%s is invalid json", s.Name))
    }

    return validator
}

func (s *StringValidator) Do() string {
    return s.Value
}

// int slice validator
type IntSliceValidator struct {
    Name  string
    Value []int
}

func (i *IntSliceValidator) Do() []int {
    return i.Value
}

// float slice validator
type FloatSliceValidator struct {
    Name  string
    Value []float64
}

func (f *FloatSliceValidator) Do() []float64 {
    return f.Value
}

// string slice validator
type StringSliceValidator struct {
    Name   string
    UseDft bool
    Value  []string
}

func (s *StringSliceValidator) Min(v int) *StringSliceValidator {
    if !s.UseDft && len(s.Value) < v {
        panic(NewException(http.StatusBadRequest, "%s has too few elements", s.Name))
    }
    return s
}

func (s *StringSliceValidator) Max(v int) *StringSliceValidator {
    if !s.UseDft && len(s.Value) > v {
        panic(NewException(http.StatusBadRequest, "%s has too many elements", s.Name))
    }
    return s
}

func (s *StringSliceValidator) Len(v int) *StringSliceValidator {
    if !s.UseDft && len(s.Value) != v {
        panic(NewException(http.StatusBadRequest, "%s has invalid length", s.Name))
    }
    return s
}

func (s *StringSliceValidator) Int() *IntSliceValidator {
    validator := &IntSliceValidator{s.Name, make([]int, 0)}
    for _, v := range s.Value {
        validator.Value = append(validator.Value, Util.ToInt(v))
    }

    return validator
}

func (s *StringSliceValidator) Float() *FloatSliceValidator {
    validator := &FloatSliceValidator{s.Name, make([]float64, 0)}
    for _, v := range s.Value {
        validator.Value = append(validator.Value, Util.ToFloat(v))
    }

    return validator
}

func (s *StringSliceValidator) Do() []string {
    return s.Value
}

// json validator
type JsonValidator struct {
    Name   string
    UseDft bool
    Value  map[string]interface{}
}

func (j *JsonValidator) Has(key string) *JsonValidator {
    if v := Util.MapGet(j.Value, key); !j.UseDft && v == nil {
        panic(NewException(http.StatusBadRequest, "%s json field missing", j.Name))
    }
    return j
}

func (j *JsonValidator) Do() map[string]interface{} {
    return j.Value
}
