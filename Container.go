package pgo

import (
    "reflect"
    "strings"
    "sync"
)

type bindItem struct {
    pool  sync.Pool     // object pool
    info  interface{}   // binding info
    zero  reflect.Value // zero value
    cmIdx int           // construct index
    imIdx int           // init index
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

    // initialize binding
    rt := iv.Elem().Type()
    item := bindItem{zero: reflect.Zero(rt), cmIdx: -1, imIdx: -1}
    item.pool.New = func() interface{} { return reflect.New(rt) }

    // get binding info
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
        case InitMethod:
            item.imIdx = i
        }
    }

    // get class name
    pkgPath := rt.PkgPath()
    name := pkgPath + "/" + rt.Name()
    if index := strings.Index(pkgPath, ControllerWeb); index >= 0 {
        name = name[index:]
    }

    if index := strings.Index(pkgPath, ControllerCmd); index >= 0 {
        name = name[index:]
    }

    if len(name) > VendorLength && name[:VendorLength] == VendorPrefix {
        name = name[VendorLength:]
    }

    c.items[name] = &item
}

// Has check if the class exists in container
func (c *Container) Has(name string) bool {
    _, ok := c.items[name]
    return ok
}

// GetInfo get class binding info
func (c *Container) GetInfo(name string) interface{} {
    if item, ok := c.items[name]; ok {
        return item.info
    }

    panic("Container: class not found, " + name)
}

// GetType get class reflect type
func (c *Container) GetType(name string) reflect.Type {
    if item, ok := c.items[name]; ok {
        return item.zero.Type()
    }

    panic("Container: class not found, " + name)
}

// Get get new class object. name is class name, config is properties map,
// params is optional construct parameters.
func (c *Container) Get(name string, config map[string]interface{}, params ...interface{}) reflect.Value {
    item, ok := c.items[name]
    if !ok {
        panic("Container: class not found, " + name)
    }

    // get new object from pool
    rv := item.pool.Get().(reflect.Value)

    if pl := len(params); pl > 0 {
        if ctx, ok := params[pl-1].(*Context); ok {
            if c.enablePool {
                // reset properties
                rv.Elem().Set(item.zero)
                ctx.cache(name, rv)
            }

            if obj, ok := rv.Interface().(IObject); ok {
                // inject context
                obj.SetContext(ctx)
            }

            params = params[:pl-1]
        }
    }

    // call Construct([arg1, arg2, ...])
    if item.cmIdx != -1 {
        if cm := rv.Method(item.cmIdx); cm.IsValid() {
            in := make([]reflect.Value, 0)
            for _, arg := range params {
                in = append(in, reflect.ValueOf(arg))
            }

            cm.Call(in)
        }
    }

    // configure object
    Configure(rv, config)

    // call Init()
    if item.imIdx != -1 {
        if im := rv.Method(item.imIdx); im.IsValid() {
            in := make([]reflect.Value, 0)
            im.Call(in)
        }
    }

    return rv
}

// Put put back reflect value to object pool
func (c *Container) Put(name string, rv reflect.Value) {
    if item, ok := c.items[name]; ok {
        item.pool.Put(rv)
        return
    }

    panic("Container: class not found, " + name)
}
