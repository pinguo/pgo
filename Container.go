package pgo

import (
    "reflect"
)

type bindItem struct {
    rt    reflect.Type // binding reflect type
    info  interface{}  // extra binding info
    cmIdx int          // Construct method index
    imIdx int          // Init method index
}

type OnReflectNew func(reflect.Value)

// Container the container component
type Container struct {
    items map[string]*bindItem
}

func (c *Container) Construct() {
    c.items = make(map[string]*bindItem)
}

// Bind bind reflect.Type to class name,
// param i must be a pointer.
func (c *Container) Bind(i interface{}) {
    iv := reflect.ValueOf(i)
    if iv.Kind() != reflect.Ptr {
        panic("Container: invalid type, need pointer")
    }

    // get reflect type and class name
    rt := iv.Elem().Type()
    name := rt.PkgPath() + "/" + rt.Name()
    if len(name) > VendorLength && name[:VendorLength] == VendorPrefix {
        name = name[VendorLength:]
    }

    item := bindItem{rt, nil, -1, -1}

    // get extra bind info
    if bind, ok := i.(IBind); ok {
        item.info = bind.GetBindInfo(i)
    }

    // get index of Construct and Init method
    it := iv.Type()
    nm := it.NumMethod()
    for i := 0; i < nm; i++ {
        switch it.Method(i).Name {
        case ConstructMethod:
            item.cmIdx = i
        case InitMethod:
            item.imIdx = i
        }
    }

    c.items[name] = &item
}

// Has check if the class has bound
func (c *Container) Has(name string) bool {
    _, ok := c.items[name]
    return ok
}

// Get get new object of the class,
// name is class name string,
// config is properties map,
// params is optional construct parameters.
func (c *Container) Get(name string, config map[string]interface{}, params ...interface{}) interface{} {
    if v, _ := c.GetValue(name, config, params...); v.IsValid() {
        return v.Interface()
    }
    return nil
}

// GetValue get reflect.Value and binding info of the class.
// name is class name string,
// config is properties map,
// params is optional construct parameters.
func (c *Container) GetValue(name string, config map[string]interface{}, params ...interface{}) (reflect.Value, interface{}) {
    item, ok := c.items[name]
    if !ok {
        panic("Container: class not found, " + name)
    }

    // construct new object
    rv := reflect.New(item.rt)

    // call hook to change object before construct
    if pl := len(params); pl > 0 {
        if hook, ok := params[pl-1].(OnReflectNew); ok {
            hook(rv)
            params = params[:pl-1]
        }
    }

    // call Construct method
    if item.cmIdx != -1 {
        if cm := rv.Method(item.cmIdx); cm.IsValid() {
            in := make([]reflect.Value, 0)
            for _, arg := range params {
                in = append(in, reflect.ValueOf(arg))
            }

            cm.Call(in)
        }
    }

    // configure this object
    Configure(rv, config)

    // call Init method
    if item.imIdx != -1 {
        if im := rv.Method(item.imIdx); im.IsValid() {
            in := make([]reflect.Value, 0)
            im.Call(in)
        }
    }

    return rv, item.info
}
