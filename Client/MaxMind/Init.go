package MaxMind

import "github.com/pinguo/pgo"

const (
    DBCountry = 0
    DBCity    = 1

    defaultLang        = "en"
    defaultComponentId = "maxMind"
)

func init() {
    container := pgo.App.GetContainer()
    container.Bind(&Adapter{})
    container.Bind(&Client{})
}
