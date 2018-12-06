package pgo

import (
    "reflect"
    "sync"
)

type bindItem struct {
    pool  sync.Pool     // object pool
    info  interface{}   // extra bind info
    zero  reflect.Value // zero value
    cmIdx int           // Construct method index
    dmIdx int           // Destruct method index
    imIdx int           // Init method index
}

// Container the container component, configuration:
// container:
//     enablePool: true
type Container struct {
    enablePool bool
    items      map[string]*bindItem
}

func (c *Container) Construct() {
    c.enablePool = true
    c.items = make(map[string]*bindItem)
}

// SetEnablePool set context based object pool enable or not, default is enabled.
func (c *Container) SetEnablePool(enable bool) {
    c.enablePool = enable
}

// Bind bind template object to class,
// param i must be a pointer of struct.
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

    item := bindItem{zero: reflect.Zero(rt), cmIdx: -1, dmIdx: -1, imIdx: -1}
    item.pool.New = func() interface{} { return reflect.New(rt) }

    // get extra bind info
    if bind, ok := i.(IBind); ok {
        item.info = bind.GetBindInfo(i)
    }

    // get method index
    it := iv.Type()
    nm := it.NumMethod()
    for i := 0; i < nm; i++ {
        switch it.Method(i).Name {
        case ConstructMethod:
            item.cmIdx = i
        case DestructMethod:
            item.dmIdx = i
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

// GetInfo get bind info of class
func (c *Container) GetInfo(name string) interface{} {
    if item, ok := c.items[name]; ok {
        return item.info
    }

    panic("Container: class not found, " + name)
}

// GetType get reflect type of class
func (c *Container) GetType(name string) reflect.Type {
    if item, ok := c.items[name]; ok {
        return item.zero.Type()
    }

    panic("Container: class not found, " + name)
}

// Get get new object of the class,
// name is class name string,
// config is properties map,
// params is optional construct parameters.
func (c *Container) Get(name string, config map[string]interface{}, params ...interface{}) interface{} {
    return c.GetValue(name, config, params...).Interface()
}

// GetValue get reflect.Value.
// name is class name string,
// config is properties map,
// params is optional construct parameters.
func (c *Container) GetValue(name string, config map[string]interface{}, params ...interface{}) reflect.Value {
    item, ok := c.items[name]
    if !ok {
        panic("Container: class not found, " + name)
    }

    // get new object from pool
    rv := item.pool.Get().(reflect.Value)

    // cache object and inject context
    if pl := len(params); pl > 0 {
        if ctx, ok := params[pl-1].(*Context); ok {
            if c.enablePool {
                ctx.addObject(name, rv)
            }

            if obj, ok := rv.Interface().(IObject); ok {
                obj.SetContext(ctx)
            }
        }
    }

    // call Construct method
    if item.cmIdx != -1 {
        if cm := rv.Method(item.cmIdx); cm.IsValid() {
            num, in := cm.Type().NumIn(), make([]reflect.Value, 0)
            for i := 0; i < num; i++ {
                in = append(in, reflect.ValueOf(params[i]))
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

    return rv
}

// PutValue put back reflect value to object pool
func (c *Container) PutValue(name string, rv reflect.Value) {
    item, ok := c.items[name]
    if !ok {
        panic("Container: class not found, " + name)
    }

    // call Destruct method
    if item.dmIdx != -1 {
        if dm := rv.Method(item.dmIdx); dm.IsValid() {
            in := make([]reflect.Value, 0)
            dm.Call(in)
        }
    }

    // reset value to zero
    rv.Elem().Set(item.zero)

    // put back to pool
    item.pool.Put(rv)
}
