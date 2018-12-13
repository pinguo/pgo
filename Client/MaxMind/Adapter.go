package MaxMind

import (
    "github.com/pinguo/pgo"
)

// Adapter of MaxMind Client, add context support.
// usage: mmd := this.GetObject("@pgo/Client/MaxMind/Adapter").(*Adapter)
type Adapter struct {
    pgo.Object
    client *Client
}

func (a *Adapter) Construct(componentId ...string) {
    id := defaultComponentId
    if len(componentId) > 0 {
        id = componentId[0]
    }

    a.client = pgo.App.Get(id).(*Client)
}

func (a *Adapter) GetClient() *Client {
    return a.client
}

// get geo info by ip, optional args:
// db int: preferred max mind db
// lang string: preferred i18n language
func (a *Adapter) GeoByIp(ip string, args ...interface{}) *Geo {
    profile := "GeoByIp:" + ip
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)

    return a.client.GeoByIp(ip, args...)
}
