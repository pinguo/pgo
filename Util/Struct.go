package Util

import (
    "reflect"
)

// STMergeSame merge none zero field of s2 into s1,
// s1 and s2 must have the same struct type.
// unexported field will be skipped and embedded struct
// field will be treated as a single field.
func STMergeSame(s1, s2 interface{}) {
    v1 := reflect.ValueOf(s1)
    if v1.Kind() == reflect.Ptr {
        v1 = v1.Elem()
    } else {
        panic("STMergeSame: param 1 must be pointer")
    }

    v2 := reflect.ValueOf(s2)
    if v2.Kind() == reflect.Ptr {
        v2 = v2.Elem()
    }

    if v1.Kind() != reflect.Struct || v2.Kind() != reflect.Struct {
        panic("STMergeSame: both params must be struct")
    }

    if v1.Type() != v2.Type() {
        panic("STMergeSame: params must have same type")
    }

    for i, n := 0, v2.NumField(); i < n; i++ {
        field1, field2 := v1.Field(i), v2.Field(i)
        zero := reflect.Zero(field2.Type())

        // merge none zero field, unexported field will be skipped.
        // use reflect.DeepEqual to check equality of underlying value,
        // because slice type does not support equal operator.
        if field2.CanInterface() && !reflect.DeepEqual(field2.Interface(), zero.Interface()) {
            if field1.CanSet() {
                field1.Set(field2)
            }
        }
    }
}

// STMergeField merge the same or compatible field of s2 into s1,
// zero and unexported field will be skipped and embedded struct
// field will be treated as a single field.
func STMergeField(s1, s2 interface{}) {
    v1 := reflect.ValueOf(s1)
    if v1.Kind() == reflect.Ptr {
        v1 = v1.Elem()
    } else {
        panic("STMergeField: param 1 must be pointer")
    }

    v2 := reflect.ValueOf(s2)
    if v2.Kind() == reflect.Ptr {
        v2 = v2.Elem()
    }

    if v1.Kind() != reflect.Struct || v2.Kind() != reflect.Struct {
        panic("STMergeField: both params must be struct")
    }

    for i, n := 0, v2.NumField(); i < n; i++ {
        field2 := v2.Field(i)
        zero := reflect.Zero(field2.Type())

        // skip zero or unexported field, see STMergeSame(s1, s2)
        if !field2.CanInterface() || reflect.DeepEqual(field2.Interface(), zero.Interface()) {
            continue
        }

        // get field1 by name of field2
        field1 := v1.FieldByName(v2.Type().Field(i).Name)
        if !field1.IsValid() || !field1.CanSet() {
            continue
        }

        // check if type of two field is convertible
        if !field2.Type().ConvertibleTo(field1.Type()) {
            continue
        }

        // set value of field1 to convertible value of field2
        field1.Set(field2.Convert(field1.Type()))
    }
}
