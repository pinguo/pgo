package Mongo

import (
    "github.com/globalsign/mgo"
    "github.com/globalsign/mgo/bson"
    "github.com/pinguo/pgo"
)

// Adapter of Mongo Client, add context support.
// usage: mongo := this.GetObject(Mongo.AdapterClass, db, coll).(*Mongo.Adapter)
type Adapter struct {
    pgo.Object
    client *Client
    db     string
    coll   string
}

func (a *Adapter) Construct(db, coll string, componentId ...string) {
    id := defaultComponentId
    if len(componentId) > 0 {
        id = componentId[0]
    }

    a.client = pgo.App.Get(id).(*Client)
    a.db = db
    a.coll = coll
}

func (a *Adapter) GetClient() *Client {
    return a.client
}

// FindOne retrieve the first document that match the query,
// query can be a map or bson compatible struct, such as bson.M or properly typed map,
// nil query is equivalent to empty query such as bson.M{}.
// result is pointer to interface{}, map, bson.M or bson compatible struct, if interface{} type
// is provided, the output result is a bson.M.
// options provided optional query option listed as follows:
// fields: bson.M, set output fields, eg. bson.M{"_id":0, "name":1},
// sort: string or []string, set sort order, eg. "key1" or []string{"key1", "-key2"},
// skip: int, set skip number, eg. 100,
// limit: int, set result limit, eg. 1,
// hint: string or []string, set index hint, eg. []string{"key1", "key2"}
//
// for example:
//      var v1 interface{} // type of output v1 is bson.M
//      m.FindOne(bson.M{"_id":"k1"}, &v1)
//
//      var v2 struct {
//          Id    string `bson:"_id"`
//          Name  string `bson:"name"`
//          Value string `bson:"value"`
//      }
//      m.FindOne(bson.M{"_id": "k1"}, &v2)
func (a *Adapter) FindOne(query interface{}, result interface{}, options ...bson.M) error {
    profile := "Mongo.FindOne"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)

    session := a.client.GetSession()
    defer session.Close()

    q := session.DB(a.db).C(a.coll).Find(query)
    a.applyQueryOptions(q, options)

    e := q.One(result)
    if e != nil && e != mgo.ErrNotFound {
        a.GetContext().Error(profile + " error, " + e.Error())
    }

    return e
}

// FindAll retrieve all documents that match the query,
// param result must be a slice(interface{}, map, bson.M or bson compatible struct)
// other params see FindOne()
func (a *Adapter) FindAll(query interface{}, result interface{}, options ...bson.M) error {
    profile := "Mongo.FindAll"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)

    session := a.client.GetSession()
    defer session.Close()

    q := session.DB(a.db).C(a.coll).Find(query)
    a.applyQueryOptions(q, options)

    e := q.All(result)
    if e != nil && e != mgo.ErrNotFound {
        a.GetContext().Error(profile + " error, " + e.Error())
    }

    return e
}

// FindAndModify execute findAndModify command, which allows atomically update or remove one document,
// param change specify the change operation, eg. mgo.Change{Update:bson.M{"$inc": bson.M{"n":1}}, ReturnNew:true},
// other params see FindOne()
func (a *Adapter) FindAndModify(query interface{}, change mgo.Change, result interface{}, options ...bson.M) error {
    profile := "Mongo.FindAndModify"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)

    session := a.client.GetSession()
    defer session.Close()

    q := session.DB(a.db).C(a.coll).Find(query)
    a.applyQueryOptions(q, options)

    _, e := q.Apply(change, result)
    if e != nil && e != mgo.ErrNotFound {
        a.GetContext().Error(profile + " error, " + e.Error())
    }

    return e
}

// FindDistinct retrieve distinct values for the param key,
// param result must be a slice,
// other params see FindOne()
func (a *Adapter) FindDistinct(query interface{}, key string, result interface{}, options ...bson.M) error {
    profile := "Mongo.FindDistinct"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)

    session := a.client.GetSession()
    defer session.Close()

    q := session.DB(a.db).C(a.coll).Find(query)
    a.applyQueryOptions(q, options)

    e := q.Distinct(key, result)
    if e != nil && e != mgo.ErrNotFound {
        a.GetContext().Error(profile + " error, " + e.Error())
    }

    return e
}

// InsertOne insert one document into collection,
// param doc can be a map, bson.M, bson compatible struct,
// for example:
//      a.InsertOne(bson.M{"field1":"value1", "field2":"value2"})
//
//      doc := struct {
//          Field1 string `bson:"field1"`
//          Field2 string `bson:"field2"`
//      } {"value1", "value2"}
//      a.InsertOne(doc)
func (a *Adapter) InsertOne(doc interface{}) error {
    profile := "Mongo.InsertOne"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)

    session := a.client.GetSession()
    defer session.Close()

    e := session.DB(a.db).C(a.coll).Insert(doc)
    if e != nil {
        a.GetContext().Error(profile + " error, " + e.Error())
    }

    return e
}

// InsertAll insert all documents provided by params docs into collection,
// for example:
//      docs := []interface{}{
//          bson.M{"_id":1, "name":"v1"},
//          bson.M{"_id":2, "name":"v2"},
//      }
//      a.InsertAll(docs)
func (a *Adapter) InsertAll(docs []interface{}) error {
    profile := "Mongo.InsertAll"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)

    session := a.client.GetSession()
    defer session.Close()

    e := session.DB(a.db).C(a.coll).Insert(docs...)
    if e != nil {
        a.GetContext().Error(profile + " error, " + e.Error())
    }

    return e
}

// UpdateOne update one document that match the query,
// mgo.ErrNotFound is returned if a document not found,
// a value of *LastError is returned if other error occurred.
func (a *Adapter) UpdateOne(query interface{}, update interface{}) error {
    profile := "Mongo.UpdateOne"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)

    session := a.client.GetSession()
    defer session.Close()

    e := session.DB(a.db).C(a.coll).Update(query, update)
    if e != nil && e != mgo.ErrNotFound {
        a.GetContext().Error(profile + " error, " + e.Error())
    }

    return e
}

// UpdateAll update all documents that match the query,
// see UpdateOne()
func (a *Adapter) UpdateAll(query interface{}, update interface{}) error {
    profile := "Mongo.UpdateAll"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)

    session := a.client.GetSession()
    defer session.Close()

    _, e := session.DB(a.db).C(a.coll).UpdateAll(query, update)
    if e != nil && e != mgo.ErrNotFound {
        a.GetContext().Error(profile + " error, " + e.Error())
    }

    return e
}

// UpdateOrInsert update a existing document that match the query,
// or insert a new document base on the update document if no document match,
// an error of *LastError is returned if error is detected.
func (a *Adapter) UpdateOrInsert(query interface{}, update interface{}) error {
    profile := "Mongo.UpdateOrInsert"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)

    session := a.client.GetSession()
    defer session.Close()

    _, e := session.DB(a.db).C(a.coll).Upsert(query, update)
    if e != nil {
        a.GetContext().Error(profile + " error, " + e.Error())
    }

    return e
}

// DeleteOne delete one document that match the query.
func (a *Adapter) DeleteOne(query interface{}) error {
    profile := "Mongo.DeleteOne"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)

    session := a.client.GetSession()
    defer session.Close()

    e := session.DB(a.db).C(a.coll).Remove(query)
    if e != nil && e != mgo.ErrNotFound {
        a.GetContext().Error(profile + " error, " + e.Error())
    }

    return e
}

// DeleteAll delete all documents that match the query.
func (a *Adapter) DeleteAll(query interface{}) error {
    profile := "Mongo.DeleteAll"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)

    session := a.client.GetSession()
    defer session.Close()

    _, e := session.DB(a.db).C(a.coll).RemoveAll(query)
    if e != nil && e != mgo.ErrNotFound {
        a.GetContext().Error(profile + " error, " + e.Error())
    }

    return e
}

// Count return the count of documents match the query.
func (a *Adapter) Count(query interface{}, options ...bson.M) (int, error) {
    profile := "Mongo.Count"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)

    session := a.client.GetSession()
    defer session.Close()

    q := session.DB(a.db).C(a.coll).Find(query)
    a.applyQueryOptions(q, options)

    n, e := q.Count()
    if e != nil {
        a.GetContext().Error(profile + " error, " + e.Error())
    }

    return n, e
}

// PipeOne execute aggregation queries and get the first item from result set.
// param pipeline must be a slice, such as []bson.M,
// param result is a pointer to interface{}, map, bson.M or bson compatible struct.
// for example:
//      pipeline := []bson.M{
//          bson.M{"$match": bson.M{"status":"A"}},
//          bson.M{"$group": bson.M{"_id":"$field1", "total":"$field2"}},
//      }
//      a.PipeOne(pipeline, &result)
func (a *Adapter) PipeOne(pipeline interface{}, result interface{}) error {
    profile := "Mongo.PipeOne"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)

    session := a.client.GetSession()
    defer session.Close()

    e := session.DB(a.db).C(a.coll).Pipe(pipeline).One(result)
    if e != nil && e != mgo.ErrNotFound {
        a.GetContext().Error(profile + " error, " + e.Error())
    }

    return e
}

// PipeAll execute aggregation queries and get all item from result set.
// param result must be slice(interface{}, map, bson.M or bson compatible struct).
// see PipeOne().
func (a *Adapter) PipeAll(pipeline interface{}, result interface{}) error {
    profile := "Mongo.PipeAll"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)

    session := a.client.GetSession()
    defer session.Close()

    e := session.DB(a.db).C(a.coll).Pipe(pipeline).All(result)
    if e != nil && e != mgo.ErrNotFound {
        a.GetContext().Error(profile + " error, " + e.Error())
    }

    return e
}

// MapReduce execute map/reduce job that match the query.
// param result is a slice(interface{}, map, bson.M, bson compatible struct),
// param query and options see FindOne().
// for example:
//      job := &mgo.MapReduce{
//          Map: "function() { emit(this.n, 1) }",
//          Reduce: "function(key, values) { return Array.sum(values) }",
//      }
//      result := []bson.M{}
//      a.MapReduce(query, job, &result)
func (a *Adapter) MapReduce(query interface{}, job *mgo.MapReduce, result interface{}, options ...bson.M) error {
    profile := "Mongo.MapReduce"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)

    session := a.client.GetSession()
    defer session.Close()

    q := session.DB(a.db).C(a.coll).Find(query)
    a.applyQueryOptions(q, options)

    _, e := q.MapReduce(job, result)
    if e != nil && e != mgo.ErrNotFound {
        a.GetContext().Error(profile + " error, " + e.Error())
    }

    return e
}

func (a *Adapter) applyQueryOptions(q *mgo.Query, options []bson.M) {
    if len(options) == 0 {
        return
    }

    for key, opt := range options[0] {
        switch key {
        case "fields":
            if fields, ok := opt.(bson.M); ok {
                q.Select(fields)
            }

        case "sort":
            if arr, ok := opt.([]string); ok {
                q.Sort(arr...)
            } else if str, ok := opt.(string); ok {
                q.Sort(str)
            }

        case "skip":
            if skip, ok := opt.(int); ok {
                q.Skip(skip)
            }

        case "limit":
            if limit, ok := opt.(int); ok {
                q.Limit(limit)
            }

        case "hint":
            if arr, ok := opt.([]string); ok {
                q.Hint(arr...)
            } else if str, ok := opt.(string); ok {
                q.Hint(str)
            }

        default:
            panic(errInvalidOpt + key)
        }
    }
}
