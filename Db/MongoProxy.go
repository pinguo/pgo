package Db

import (
	"gopkg.in/mgo.v2"
	"time"
	"github.com/pinguo/pgo"
	"strings"
)

type MongoProxy struct {
	 connectInfo [3]string
	 pgo.Object
	 profilename string

}

//设置链接信息
// m.info[0] componentId 组件id
// m.info[1] dbname 数据库名
// m.info[2] connection 集合名
func (m *MongoProxy) SetConnectInfo(info [3]string){
	m.connectInfo = info

	m.profilename = strings.Join(m.connectInfo[:], ".")
}

// 查询单个数据
// query 查询条件
//fields 查询的字段
//result 返回的数据  指针
func (m *MongoProxy) FindOne( query interface{},fields interface{}, result interface{}) (err error) {
	profileKey := m.profilename + ".FindOne"
	m.GetContext().ProfileStart(profileKey)
	defer m.GetContext().ProfileStop(profileKey)
    session := pgo.App.Get(m.connectInfo[0]).(IMongoSession).Session()
	defer session.Close()
	c := session.DB(m.connectInfo[1]).C(m.connectInfo[2])
	queryObj := c.Find(query)
	if fields!=nil{
		queryObj.Select(fields)
	}
	err = queryObj.One(result)
	return err
}

// 插入数据
// addData 增加数据
func (m *MongoProxy) Insert( addData interface{}) (error) {

	profileKey := m.profilename + ".Insert"

	m.GetContext().ProfileStart(profileKey)
	defer m.GetContext().ProfileStop(profileKey)
    session := pgo.App.Get(m.connectInfo[0]).(IMongoSession).Session()
	defer session.Close()
	c := session.DB(m.connectInfo[1]).C(m.connectInfo[2])
	return c.Insert(addData)
}

// 批量插入数据
// addData 增加数据数组
func (m *MongoProxy) BatchAdd( addData []interface{}) (error) {
	profileKey := m.profilename + ".BatchAdd"
	m.GetContext().ProfileStart(profileKey)
	defer m.GetContext().ProfileStop(profileKey)
    session := pgo.App.Get(m.connectInfo[0]).(IMongoSession).Session()
	defer session.Close()
	c := session.DB(m.connectInfo[1]).C(m.connectInfo[2])
	return c.Insert(addData...)
}

// 获取数据集
// query 查询条件
// fields 查询字段
// sort 排序
// limit 限制返回的数据文档数
// skip 开始返回的offset
// timeout 查询超时时间，default 2s
//result 返回的数据 指针
func (m *MongoProxy) Query( query, fields interface{}, sort []string, limit int, skip int, timeout time.Duration, result []interface{}) ( err error) {
	profileKey := m.profilename + ".Query"
	m.GetContext().ProfileStart(profileKey)
	defer m.GetContext().ProfileStop(profileKey)
    session := pgo.App.Get(m.connectInfo[0]).(IMongoSession).Session()
	defer session.Close()
	c := session.DB(m.connectInfo[1]).C(m.connectInfo[2])
	var queryObj *mgo.Query
	queryObj = c.Find(query)
	if fields != nil {
		queryObj = queryObj.Select(fields)
	}
	if sort != nil {
		queryObj = queryObj.Sort(sort...)
	}
	if limit != 0 {
		queryObj = queryObj.Limit(limit)
	}
	if skip != 0 {
		queryObj = queryObj.Skip(skip)
	}
	if timeout == 0 {
		timeout = 2000 * time.Millisecond
	} else {
		timeout = timeout * time.Millisecond
	}
	queryObj = queryObj.SetMaxTime(timeout)
	err = queryObj.All(result)
	return err
}

// 对当前Collection所在Database上执行command
// module 模块名
// dbname 数据库名
// cmd 运行命令
func (m *MongoProxy) Run(module, dbname string, cmd interface{}) (result interface{}, err error) {
	profileKey := module + "." + dbname  + ".Run"
	m.GetContext().ProfileStart(profileKey)
	defer m.GetContext().ProfileStop(profileKey)
    session := pgo.App.Get(module).(IMongoSession).Session()
	defer session.Close()
	db := session.DB(dbname)
	err = db.Run(cmd, &result)
	return result, err
}

// 获取count
// selector 限制条件
func (m *MongoProxy) Count( selector interface{}) (int, error) {
	profileKey := m.profilename + ".Count"
	m.GetContext().ProfileStart(profileKey)
	defer m.GetContext().ProfileStop(profileKey)
    session := pgo.App.Get(m.connectInfo[0]).(IMongoSession).Session()
	defer session.Close()
	c := session.DB(m.connectInfo[1]).C(m.connectInfo[2])
	return c.Find(selector).Count()
}

// 删除单条数据
// selector 限制条件
func (m *MongoProxy) Delete( selector interface{}) (error) {
	profileKey := m.profilename + ".Delete"
	m.GetContext().ProfileStart(profileKey)
	defer m.GetContext().ProfileStop(profileKey)
    session := pgo.App.Get(m.connectInfo[0]).(IMongoSession).Session()
	defer session.Close()
	c := session.DB(m.connectInfo[1]).C(m.connectInfo[2])
	return c.Remove(selector)
}

// 删除单条数据
// selector 限制条件
func (m *MongoProxy) DeleteId( id interface{}) (error) {
	profileKey := m.profilename + ".DeleteId"
	m.GetContext().ProfileStart(profileKey)
	defer m.GetContext().ProfileStop(profileKey)
    session := pgo.App.Get(m.connectInfo[0]).(IMongoSession).Session()
	defer session.Close()
	c := session.DB(m.connectInfo[1]).C(m.connectInfo[2])
	return c.RemoveId(id)
}

// 删除多条数据
// selector 限制条件
func (m *MongoProxy) DeleteAll( selector interface{}) (*mgo.ChangeInfo, error) {
	profileKey := m.profilename + ".DeleteAll"
	m.GetContext().ProfileStart(profileKey)
	defer m.GetContext().ProfileStop(profileKey)
    session := pgo.App.Get(m.connectInfo[0]).(IMongoSession).Session()
	defer session.Close()
	c := session.DB(m.connectInfo[1]).C(m.connectInfo[2])
	return c.RemoveAll(selector)
}

// 修改单条数据
// selector 限制条件
func (m *MongoProxy) Modify( selector interface{}, update interface{}) (error) {
	profileKey := m.profilename + ".Modify"
	m.GetContext().ProfileStart(profileKey)
	defer m.GetContext().ProfileStop(profileKey)
    session := pgo.App.Get(m.connectInfo[0]).(IMongoSession).Session()
	defer session.Close()
	c := session.DB(m.connectInfo[1]).C(m.connectInfo[2])
	updateDoc := make(map[string]interface{})
	updateDoc["$set"] = update
	return c.Update(selector, updateDoc)
}

// 根据id修改单条数据
// selector 限制条件
func (m *MongoProxy) ModifyById( id interface{}, update interface{}) (error) {
	profileKey := m.profilename + ".ModifyById"
	m.GetContext().ProfileStart(profileKey)
	defer m.GetContext().ProfileStop(profileKey)
    session := pgo.App.Get(m.connectInfo[0]).(IMongoSession).Session()
	defer session.Close()
	c := session.DB(m.connectInfo[1]).C(m.connectInfo[2])
	updateDoc := make(map[string]interface{})
	updateDoc["$set"] = update
	return c.UpdateId(id, updateDoc)
}

// 修改单条数据
// selector 限制条件
func (m *MongoProxy) ModifyUpsert( selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	profileKey := m.profilename + ".ModifyUpsert"
	m.GetContext().ProfileStart(profileKey)
	defer m.GetContext().ProfileStop(profileKey)
    session := pgo.App.Get(m.connectInfo[0]).(IMongoSession).Session()
	defer session.Close()
	c := session.DB(m.connectInfo[1]).C(m.connectInfo[2])
	updateDoc := make(map[string]interface{})
	updateDoc["$set"] = update
	return c.Upsert(selector, updateDoc)
}

// 修改多条数据
// selector 限制条件
func (m *MongoProxy) ModifyAll( selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	profileKey := m.profilename + ".ModifyAll"
	m.GetContext().ProfileStart(profileKey)
	defer m.GetContext().ProfileStop(profileKey)
    session := pgo.App.Get(m.connectInfo[0]).(IMongoSession).Session()
	defer session.Close()
	c := session.DB(m.connectInfo[1]).C(m.connectInfo[2])
	updateDoc := make(map[string]interface{})
	updateDoc["$set"] = update
	return c.UpdateAll(selector, updateDoc)
}

// 修改单条数据
// selector 限制条件
func (m *MongoProxy) UpdateDoc( selector interface{}, update interface{}) (error) {
	profileKey := m.profilename + ".UpdateDoc"
	m.GetContext().ProfileStart(profileKey)
	defer m.GetContext().ProfileStop(profileKey)
    session := pgo.App.Get(m.connectInfo[0]).(IMongoSession).Session()
	defer session.Close()
	c := session.DB(m.connectInfo[1]).C(m.connectInfo[2])
	return c.Update(selector, update)
}

// 修改单条数据
// selector 限制条件
func (m *MongoProxy) UpdateDocUpsert( selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	profileKey := m.profilename + ".UpdateDocUpsert"
	m.GetContext().ProfileStart(profileKey)
	defer m.GetContext().ProfileStop(profileKey)
    session := pgo.App.Get(m.connectInfo[0]).(IMongoSession).Session()
	defer session.Close()
	c := session.DB(m.connectInfo[1]).C(m.connectInfo[2])
	return  c.Upsert(selector, update)
}

// 修改多条数据
// selector 限制条件
func (m *MongoProxy) UpdateDocAll( selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	profileKey := m.profilename + ".UpdateDocAll"
	m.GetContext().ProfileStart(profileKey)
	defer m.GetContext().ProfileStop(profileKey)
    session := pgo.App.Get(m.connectInfo[0]).(IMongoSession).Session()
	defer session.Close()
	c := session.DB(m.connectInfo[1]).C(m.connectInfo[2])
	return  c.UpdateAll(selector, update)
}
