package Util

import (
    "fmt"
    "io"

    "github.com/go-yaml/yaml"
)

// YamlMarshal wrapper for yaml.Marshal.
func YamlMarshal(in interface{}) ([]byte, error) {
    return yaml.Marshal(in)
}

// YamlUnmarshal wrapper for yaml.Unmarshal or yaml.UnmarshalStrict,
// if type of out is map[string]interface{}, *map[string]interface{},
// the inner map[interface{}]interface{} will fix to map[string]interface{}
// recursively. if type of out is *interface{}, the underlying type of
// out will change to *map[string]interface{}.
func YamlUnmarshal(in []byte, out interface{}, strict ...bool) error {
    var err error
    if len(strict) > 0 && strict[0] {
        err = yaml.UnmarshalStrict(in, out)
    } else {
        err = yaml.Unmarshal(in, out)
    }

    if err == nil {
        yamlFixOut(out)
    }

    return err
}

// YamlEncode wrapper for yaml.Encoder.
func YamlEncode(w io.Writer, in interface{}) error {
    enc := yaml.NewEncoder(w)
    defer enc.Close()
    return enc.Encode(in)
}

// YamlDecode wrapper for yaml.Decoder, strict is for Decoder.SetStrict().
// if type of out is map[string]interface{}, *map[string]interface{},
// the inner map[interface{}]interface{} will fix to map[string]interface{}
// recursively. if type of out is *interface{}, the underlying type of
// out will change to *map[string]interface{}.
func YamlDecode(r io.Reader, out interface{}, strict ...bool) error {
    dec := yaml.NewDecoder(r)
    if len(strict) > 0 && strict[0] {
        dec.SetStrict(true)
    }

    err := dec.Decode(out)
    if err == nil {
        yamlFixOut(out)
    }

    return err
}

func yamlFixOut(out interface{}) {
    switch v := out.(type) {
    case *map[string]interface{}:
        for key, val := range *v {
            (*v)[key] = yamlCleanValue(val)
        }

    case map[string]interface{}:
        for key, val := range v {
            v[key] = yamlCleanValue(val)
        }

    case *interface{}:
        if vv, ok := (*v).(map[interface{}]interface{}); ok {
            *v = yamlCleanMap(vv)
        }
    }
}

func yamlCleanValue(v interface{}) interface{} {
    switch vv := v.(type) {
    case map[interface{}]interface{}:
        return yamlCleanMap(vv)

    case []interface{}:
        return yamlCleanArray(vv)

    default:
        return v
    }
}

func yamlCleanMap(in map[interface{}]interface{}) map[string]interface{} {
    result := make(map[string]interface{}, len(in))
    for k, v := range in {
        result[fmt.Sprintf("%v", k)] = yamlCleanValue(v)
    }
    return result
}

func yamlCleanArray(in []interface{}) []interface{} {
    result := make([]interface{}, len(in))
    for k, v := range in {
        result[k] = yamlCleanValue(v)
    }
    return result
}
