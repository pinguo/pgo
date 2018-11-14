# PGO
PGO应用框架即"Pinguo GO application framework"，是Camera360广告服务端团队研发的一款简单、高性能、组件化的GO应用框架。受益于GO语言高性能与原生协程，业务从php+yii2升级到PGO后，线上表现单机处理能力提高10倍。PGO参考了php-yii2/php-msf/go-gin等框架的设计思想，可快速地开发出高性能的web应用程序。


参考文档：[pgo-docs](https://github.com/pinguo/pgo-docs)

应用示例：[pgo-demo](https://github.com/pinguo/pgo-demo)

## 目录
- [环境要求](#环境要求)
- [项目目录](#项目目录)
- [依赖管理](#依赖管理)
- [基准测试](#基准测试)
- [快速开始](#快速开始)
- [使用示例](#使用示例)
    - [配置(Config)](#配置config)
    - [控制器(Controller)](#控制器controller)
    - [上下文(Context)](#上下文context)
    - [容器(Container)](#容器container)
    - [组件(Component)](#组件component)
    - [其它](#其它)

## 环境要求
- GO 1.10+
- Make 3.8+
- Linux/MacOS
- Glide 0.13+ (建议)
- GoLand 2018 (建议)

## 项目目录
规范：
- 一个项目为一个独立的目录，不使用GO全局工作空间。
- 项目的GOPATH为项目根目录，不要依赖系统的GOPATH。
- 除GO标准库外，所有外部依赖代码放入"src/vendor"。
- 项目源码文件与目录使用大写驼峰(CamelCase)形式。

```
<project>
├── bin/                # 编译程序目录
├── conf/               # 配置文件目录
│   ├── production/     # 环境配置目录
│   │   ├── app.json
│   │   └── params.json
│   ├── testing/
│   ├── app.json        # 项目配置文件
│   └── params.json     # 自定义配置文件
├── makefile            # 编译打包
├── runtime/            # 运行时目录
├── public/             # 静态资源目录
├── view/               # 视图模板目录
└── src/                # 项目源码目录
    ├── Command/        # 命令行控制器目录
    ├── Controller/     # HTTP控制器目录
    ├── Lib/            # 项目基础库目录
    ├── Main/           # 项目入口目录
    ├── Model/          # 模型目录(数据交互)
    ├── Service/        # 服务目录(业务逻辑)
    ├── Struct/         # 结构目录(数据定义)
    ├── Test/           # 测试目录(单测/性能)
    ├── vendor/         # 第三方依赖目录
    ├── glide.lock      # 项目依赖锁文件
    └── glide.yaml      # 项目依赖配置文件
```

## 依赖管理
建议使用glide做为依赖管理工具(类似php的composer)，不使用go官方的dep工具

安装(mac)：`brew install glide`

使用(调用目录为项目的src目录)：
```
glide init              # 初始化项目
glide get <pkg>         # 下载pkg并添加依赖
    --all-dependencies  # 下载pkg的所有依赖
glide get <pkg>#v1.2    # 下载指定版本的pkg
glide install           # 根据lock文件下载依赖
glide update            # 更新依赖包
```

## 基准测试
TODO

## 快速开始
1. 创建项目目录(以下三种方法均可)
    - 参见《项目目录》手动创建
    - 从[pgo-demo](https://github.com/pinguo/pgo-demo)克隆目录结构
    - 复制makefile至项目根目录并执行`make init`
2. 修改配置文件(conf/app.json)
    ```json
    {
        "name": "pgo-demo",
        "GOMAXPROCS": 2,
        "runtimePath": "@app/runtime",
        "publicPath": "@app/public",
        "viewPath": "@app/view",
        "server": {
            "httpAddr": "0.0.0.0:8000",
            "readTimeout": "30s",
            "writeTimeout": "30s"
        },
        "components": {
            "log": {
                "levels": "ALL",
                "targets": {
                    "info": {
                        "class": "@pgo/FileTarget",
                        "levels": "DEBUG,INFO,NOTICE",
                        "filePath": "@runtime/info.log",
                    },
                    "error": {
                        "class": "@pgo/FileTarget",
                        "levels": "WARN,ERROR,FATAL",
                        "filePath": "@runtime/error.log",
                    },
                    "console": {
                        "class": "@pgo/ConsoleTarget",
                        "levels": "ALL"
                    }
                }
            }
        }
    }
    ```
3. 安装PGO(以下两种方法均可)
    - 在项目根目录执行`export GOPATH=$(pwd) && cd src && glide get github.com/pinguo/pgo`
    - 复制makefile至项目根目录并执行`make pgo`
4. 创建控制器(src/Controller/WelcomeController.go)
    ```go
    package Controller

    import (
        "github.com/pinguo/pgo"
        "net/http"
        "time"
    )

    type WelcomeController struct {
        pgo.Controller
    }

    func (w *WelcomeController) ActionIndex() {
        data := pgo.Map{"text": "welcome to pgo-demo", "now": time.Now()}
        w.OutputJson(data, http.StatusOK)
    }
    ```
5. 注册控制器(src/Controller/Init.go)
    ```go
    package Controller

    import "github.com/pinguo/pgo"

    func init() {
        container := pgo.App.GetContainer()

        container.Bind(&WelcomeController{})
    }
    ```
6. 创建程序入口(src/Main/main.go)
    ```go
    package main

    import (
        _ "Controller" // 导入控制器

        "github.com/pinguo/pgo"
    )

    func main() {
        pgo.Run() // 运行程序
    }
    ```
7. 编译运行
    ```sh
    make start
    curl http://127.0.0.1:8000/welcome
    ```

## 使用示例
### 配置(Config)
- 项目配置文件`conf/app.json`, 可任意添加自定义配置文件如params.json
- 目前仅支持json配置文件，后续会支持yaml配置文件
- 所有配置文件均是一个json对象
- 支持任意环境目录，环境目录中的同名字段会覆盖基础配置中的字段
- 通过bin/binName --env production指定程序环境目录`production`
- 配置都有默认值，配置文件中的值会覆盖默认值(默认值参见组件说明)
- 配置文件支持环境变量，格式`${envName||default}`，当envName不存在时使用default
- 配置文件中路径及类名支持别名字符串，PGO定义的别名如下：
    - `@app` 项目根目录绝对路径
    - `@runtime` 项目运行时目录绝对路径
    - `@view` 项目视图模板目录绝对路径
    - `@pgo` PGO框架import路径

示例：
```go
cfg := pgo.App.GetConfig() // 获取配置对象
name := cfg.GetString("app.name", "demo") // 获取String，不存在返回"demo"
procs := cfg.GetInt("app.GOMAXPROCS", 2) // 获取Integer, 不存在返回2
price := cfg.GetFloat("params.goods.price", 0) // 获取Float, 不存在返回0
enable := cfg.GetBool("params.detect.enable", false) // 获取Bool, 不存在返回false

// 除基本类型外，通过Get方法获取原始配置数据，需要进行类型转换
plugins, ok := cfg.Get("app.servers.plugins").([]interface{}) // 获取数组
log, ok := cfg.Get("app.conponents.log").(map[string]interface{}) // 获取对象
```

### 控制器(Controller)
- 支持HTTP(Controller)和命令行(Command)控制器
- 支持`URL`路由和`正则`路由(详见Router组件)
- 支持URL动作(Action)和RESTFULL动作(Action)
- 支持参数验证(详见ValidateXxx方法)
- 支持BeforeAction/AfterAction/FinishAction钩子
- 支持HandlePanic钩子，捕获未处理异常
- 支持过滤器(Filter)，TODO
- 提供OutputXxx方法，方便输出各种类型数据

示例：
```go
type WelcomeController struct {
    pgo.Controller
}

// 可选构造函数(框架自动调用)
func (w *WelcomeController) Construct() {}

// 可选初始化函数(框架自动调用)
func (w *WelcomeController) Init() {}

// 默认动作为index, /path/to/welcome调用此动作
func (w *WelcomeController) ActionIndex() {
    data := pgo.Map{"text": "welcome to pgo-demo", "now": time.Now()}
    w.OutputJson(data, http.StatusOK)
}

// URL路由动作，根据url自动映射控制器及方法，不需要配置.
// url的最后一段为动作名称，不存在则为index,
// url的其余部分为控制器名称，不存在则为index,
// 例如：/path/to/welcome/say-hello，控制器类名为
// Path/To/WelcomeController 动作方法名为ActionSayHello
func (w *WelcomeController) ActionSayHello() {
    ctx := w.GetContext() // 获取PGO请求上下文件

    // 验证参数，提供参数名和默认值，当不提供默认值时，表明该参数为必选参数。
    // 详细验证方法参见Validate.go
    name := ctx.ValidateParam("name").Min(5).Max(50).Do() // 验证GET/POST参数(string)，为空或验证失败时panic
    age := ctx.ValidateQuery("age", 20).Int().Min(1).Max(100).Do() // 只验证GET参数(int)，为空或失败时返回20
    ip := ctx.ValidatePost("ip", "").IPv4().Do() // 只验证POST参数(string), 为空或失败时返回空字符串

    // 打印日志
    ctx.Info("request from welcome, name:%s, age:%d, ip:%s", name, age, ip)
    ctx.PushLog("clientIp", ctx.GetClientIp()) // 生成clientIp=xxxxx在pushlog中

    // 调用业务逻辑，一个请求生命周期内的对象都要通过GetObject()获取，
    // 这样可自动查找注册的类，并注入请求上下文(Context)到对象中。
    svc := w.GetObject("Service/Welcome").(*Service.Welcome)

    // 添加耗时到profile日志中
    ctx.ProfileStart("Welcome.SayHello")
    svc.SayHello(name, age, ip)
    ctx.ProfileStop("Welcome.SayHello")

    data := pgo.Map{
        "name": name,
        "age": age,
        "ip": ip,
    }

    // 输出json数据
    w.OutputJson(data, http.StatusOK)
}

// 正则路由动作，需要配置Router组件(components.router.rules)
// 规则中捕获的参数通过动作函数参数传递，没有则为空字符串.
// eg. "^/reg/eg/(\\w+)/(\\w+)$ => /welcome/regexp-example"
func (w *WelcomeController) ActionRegexpExample(p1, p2 string) {
    data := pgo.Map{"p1": p1, "p2": p2}
    w.OutputJson(data, http.StatusOK)
}

// RESTFULL动作，url中没有指定动作名，使用请求方法作为动作的名称(需要大写)
// 例如：GET方法请求ActionGET(), POST方法请求ActionPOST()
func (w *WelcomeController) ActionGET() {}
```

### 上下文(Context)
- 上下文存在于一个请求的生命周期中
- 包含一个请求的上下文信息(输入、输出、自定义数据)
- 继承pgo.Object的类通过pgo.Object.GetObject()会自动注入当前上下文

示例：
```go
ctx.GetParam("p1", "")  // 获取GET/POST参数，默认空串
ctx.GetQuery("p2", "")  // 获取GET参数，默认空串
ctx.GetPost("p3", "")   // 获取POST参数，默认空串
ctx.GetHeader("h1", "") // 获取Header，默认空串
ctx.GetCookie("c1", "") // 获取Cookie，默认空串
ctx.GetPath()           // 获取请求路径
ctx.SetUserData("u1", "v1") // 设置自定义数据
ctx.GetUserData("u1", "")   // 获取自定义数据
ctx.GetClientIp()           // 获取客户端IP
ctx.ValidateParam("p1", "").Do()    // 获取并验证GET/POST参数(有默认值)
ctx.ValidateQuery("p2").Do()        // 获取并验证GET参数(必选参数)
ctx.ValidatePost("p3").Do()         // 获取并验证POST参数(必选参数)
ctx.End(status, data)               // 输出数据到response

ctx.Debug/Info/Notice/Warn/Error/Fatal()    // 日志输出函数(带日志跟踪ID)
ctx.PushLog("key", "val")                   // 记录pushlog
ctx.Counting("key", 1, 1)                   // 记录命中记数
ctx.ProfileStart/ProfileStop/ProfileAdd()   // 记录耗时数据
```

### 容器(Container)
- 容器用于类的注册与创建
- 通过容器创建的对象会自动按序调用以下函数(如果有)：
    - 构造函数(Construct)
    - 属性设置(SetXxx)
    - 初始函数(Init)
- 构造函数支持任意个参数
- 属性可以通过Set方法或导出字段设置

示例：
```go
type People struct {
    // 继承自pgo.Object可增加上下文支持，
    // 由于组件是全局对象，没有请求上下文，
    // 所以组件是不能继承自pgo.Object的。
    pgo.Object

    name    string
    age     int
    sex     string
}

// 可选构造函数，用于设置初始值
func (p *People) Construct() {
    p.name = "unknown"
    p.age  = 0
    p.sex  = "unknown"
}

// 可选初始函数，对象创建完成回调
func (p *People) Init() {
    fmt.Printf("people created, name:%s age:%d sex:%s\n", p.name, p.age, p.sex)
}

// 可选设置函数，根据配置自动调用
func (p *People) SetName(name string) {
    p.name = name
}

func (p *People) SetAge(age int) {
    p.age = age
}

func (p *People) SetSex(sex string) {
    p.sex = sex
}

// init方法通常放在包的Init.go文件中
func init() {
    container := pgo.App.GetContainer() // 获取容器对象
    container.Bind(&People{}) // 注册类模板对象
}

// 获取People的新对象
func (t *TestController) ActionTest() {
    // 方法1: 通过类字符串获取
    p1 := t.GetObject("People").(*People)

    // 方法2: 如果构造函数定义有参数，可以传递构造参数
    p2 := t.GetObject("People", arg1, arg2).(*People)

    // 方法3: 指定配置属性(通常用于从配置文件生成组件)
    conf := map[string]interface{} {
        "class" "People",   // 配置必须包含class字段
        "name": "zhang san",
        "age": 30,
        "sex": "male",
    }
    p3 := t.GetObject(conf).(*People)
}
```

### 组件(Component)
- 组件为全局单例对象
- 组件第一次使用时自动构造、配置和初始化
- 组件是全局对象没有上下文
- 组件通常使用配置文件进行配置

示例：
```go
// 日志组件配置示例，app.components.log
// "log": { //组件ID, class固定为"@pgo/Dispatcher"
//     "levels": "ALL",
//     "traceLevels": "DEBUG"
//     "chanLen": 1000,
//     "flushInterval": "60s",
//     "targets": {
//         "info": {
//             "class": "@pgo/FileTarget",
//             "levels": "DEBUG,INFO,NOTICE",
//             "filePath": "@runtime/info.log",
//             "maxLogFile": 10
//         },
//         "error": {
//             "class": "@pgo/FileTarget",
//             "levels": "WARN,ERROR,FATAL",
//             "filePath": "@runtime/error.log",
//             "maxLogFile": 10
//         }
//     }
// }

// 获取日志组件(核心组件通过框架提供的方法获取)
log := pgo.App.GetLog()

// redis组件配置示例，app.components.redis
// "redis": {
//     "class": "@pgo/Client/Redis/Client",
//     "prefix": "pgo_",
//     "password": "",
//     "db": 0,
//     "maxIdleConn": 10,
//     "maxIdleTime": "60s",
//     "netTimeout": "1s",
//     "probInterval": "0s",
//     "servers": [
//         "127.0.0.1:6379",
//         "127.0.0.1:6380"
//     ]
// }

// 获取Redis组件(非核心组件需要进行类型转换)
redis := pgo.App.Get("redis").(*Redis.Client)
```

### 其它
参见[pgo-docs](https://github.com/pinguo/pgo-docs)
