package UnitTesting

import (
    "github.com/pinguo/pgo"
)

var TestObj *pgo.Object

func init() {
    TestObj = GetTestObj()
}

func GetTestObj() *pgo.Object {
    TestObj := &pgo.Object{}
    ctx := &pgo.Context{}
    TestObj.SetContext(ctx)
    return TestObj
}
