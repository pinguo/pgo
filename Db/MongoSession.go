package Db

import (
    "gopkg.in/mgo.v2"
)

type MongoSession struct {
    session     *mgo.Session
    Server      string // mongodb://127.0.0.1:27017
    ReplicaSet  string
    MaxPoolSize string
}

func (m *MongoSession) Construct() {
    m.ReplicaSet = ""
    m.MaxPoolSize = "10000"
}

// 获取mongo session
func (m *MongoSession) Session() *mgo.Session {
    if m.session == nil {
        var err error
        m.session, err = m.initSession()
        if err != nil {
            panic(err)
        }
    }
    return m.session.New()
}

// 初始化
func (m *MongoSession) initSession() (*mgo.Session, error) {
    baseUrl := m.Server + "?connect=direct"
    if m.ReplicaSet != "" {
        baseUrl = baseUrl + "&replicaSet=" + m.ReplicaSet
    }
    if m.MaxPoolSize != "" {
        baseUrl = baseUrl + "&maxPoolSize=" + m.MaxPoolSize
    }
    session, err := mgo.Dial(baseUrl)
    session.SetMode(mgo.Monotonic, true)
    if err != nil {
        panic(err)
    }
    return session, nil
}
