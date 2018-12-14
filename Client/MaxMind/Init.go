package MaxMind

import "github.com/pinguo/pgo"

const (
    DBCountry    = 0
    DBCity       = 1
    AdapterClass = "@pgo/Client/MaxMind/Adapter"

    defaultComponentId = "maxMind"
    defaultLang        = "en"
)

func init() {
    container := pgo.App.GetContainer()
    container.Bind(&Adapter{})
    container.Bind(&Client{})
}
