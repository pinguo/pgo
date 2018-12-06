package pgo

// Object base class of context based object
type Object struct {
    context *Context
}

// GetContext get context of this object
func (o *Object) GetContext() *Context {
    return o.context
}

// SetContext set context of this object
func (o *Object) SetContext(ctx *Context) {
    o.context = ctx
}

// GetObject create new object and inject context
func (o *Object) GetObject(class interface{}, params ...interface{}) interface{} {
    params = append(params, o.GetContext())
    return CreateObject(class, params...)
}
