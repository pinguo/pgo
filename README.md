# PGO
PGO应用框架即"Pinguo GO application framework"，是Camera360广告服务端团队研发的一款简单、高性能、组件化的GO应用框架。受益于GO语言高性能与原生协程，业务从php+yii2升级到PGO后，线上表现单机处理能力提高10倍。PGO吸收了php-yii2/php-msf/go-gin等框架的设计思想，平衡了开发效率和运行性能，使用PGO可快速地开发出高性能的web应用程序。

参考文档：[pgo-docs](https://github.com/pinguo/pgo-docs)

应用示例：[pgo-demo](https://github.com/pinguo/pgo-demo)

## 基准测试
主要测试PGO框架与php-yii2，php-msf，go-gin的性能差异。

说明:
- 测试机为4核8G虚拟机
- php版本为7.1.24, 开启opcache
- go版本为1.11.2, GOMAXPROCS=4
- swoole版本1.9.21, worker_num=4, reactor_num=2
- 输出均为字符串{"code": 200, "message": "success","data": "hello world"}
- 命令: ab -n 1000000 -c 100 -k 'http://target-ip:8000/welcome'

分类 | QPS | 平均响应时间(ms) |CPU
---- | ---- | ---- | -----
php-yii2 | 2715 | 36.601 | 72%
php-msf | 20053 | 4.575 | 73%
go-gin | 41798 | 2.339 | 55%
go-pgo | 33902 | 2.842 | 64%

结论:
- pgo相比yii2性能提升10+倍, 对低于php7的版本性能还要翻倍。
- pgo相比msf性能提升70%, 相较于msf的yield模拟的协程，pgo协程理解和使用更简单。
- pgo相比gin性能降低19%, 但pgo内置多种常用组件，工程化做得更好，使用方式类似yii2和msf。

## 环境要求
- GO 1.10+
- Make 3.8+
- Linux/MacOS/Cygwin
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
│   │   ├── app.yaml
│   │   └── params.yaml
│   ├── testing/
│   ├── app.yaml        # 项目配置文件
│   └── params.yaml     # 自定义配置文件
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
    ├── Test/           # 测试目录
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

## 快速开始
1. 拷贝makefile
    
    非IDE环境(命令行)下，推荐使用make做为编译打包的控制工具，从[pgo](https://github.com/pinguo/pgo)或[pgo-demo](https://github.com/pinguo/pgo-demo)将makefile复制到项目目录下。
    ```sh
    make start      # 编译并运行当前工程
    make stop       # 停止当前工程的进程
    make build      # 仅编译当前工程
    make update     # 更新glide依赖(递归更新)
    make install    # 安装glide.lock文件锁定的依赖包
    make pgo        # 安装pgo框架到当前工程
    make init       # 初始化工程目录
    make help       # 输出帮助信息
    ```

2. 创建项目目录(以下三种方法均可)
    - 执行`make init`创建目录
    - 参见《项目目录》手动创建
    - 从[pgo-demo](https://github.com/pinguo/pgo-demo)克隆目录结构

3. 修改配置文件(conf/app.yaml)
    ```yaml
    name: "pgo-demo"
    GOMAXPROCS: 2
    runtimePath: "@app/runtime"
    publicPath: "@app/public"
    viewPath: "@app/view"
    server:
        httpAddr: "0.0.0.0:8000"
        readTimeout: "30s"
        writeTimeout: "30s"
    components:
        log:
            levels: "ALL"
            targets:
                info:
                    class: "@pgo/FileTarget"
                    levels: "DEBUG,INFO,NOTICE"
                    filePath: "@runtime/info.log"
                error:
                    class: "@pgo/FileTarget"
                    levels: "WARN,ERROR,FATAL"
                    filePath: "@runtime/error.log"
                console: 
                    class: "@pgo/ConsoleTarget"
                    levels: "ALL"
    ```

4. 安装PGO(以下两种方法均可)
    - 在项目根目录执行`make pgo`安装PGO
    - 在项目根目录执行`export GOPATH=$GOPATH:$(pwd) && cd src && glide get github.com/pinguo/pgo && glide update`
5. 创建Service(src/Service/Welcome.go)
    ```go
    package Service

    import (
        "fmt"

        "github.com/pinguo/pgo"
    )

    type Welcome struct {
        pgo.Object
    }

    // 框架自动调用的构造函数(可选)
    func (w *Welcome) Construct() {
        fmt.Printf("call in Service/Welcome.Construct\n")
    }

    // 框架自动调用的初始函数(可选)
    func (w *Welcome) Init() {
        fmt.Printf("call in Service/Welcome.Init\n")
    }

    func (w *Welcome) SayHello(name string, age int, sex string) {
        fmt.Printf("call in  Service/Welcome.SayHello, name:%s age:%d sex:%s\n", name, age, sex)
    }
    ```
6. 注册Service(src/Service/Init.go)
    
    ```go
    package Service

    import "github.com/pinguo/pgo"

    func init() {
        container := pgo.App.GetContainer()

        // 注册类
        container.Bind(&Welcome{})

        // 除控制器目录外，其它包的init函数中应该只注册该包的类，
        // 而不应该包含子包。
    }

    ```
7. 创建控制器(src/Controller/WelcomeController.go)
    ```go
    package Controller

    import (
        "Service"
        "net/http"
     
        "github.com/pinguo/pgo"
    )

    type WelcomeController struct {
        pgo.Controller
    }

    // 默认动作为index, 通过/welcome或/welcome/index调用
    func (w *WelcomeController) ActionIndex() {
        w.OutputJson("hello world", http.StatusOK)
    }
    
    // URL路由动作，根据url自动映射控制器及方法，不需要配置.
    // url的最后一段为动作名称，不存在则为index,
    // url的其余部分为控制器名称，不存在则为index,
    // 例如：/welcome/say-hello，控制器类名为
    // Controller/WelcomeController 动作方法名为ActionSayHello
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
    func (w *WelcomeController) ActionGET() {
        w.GetContext().End(http.StatusOK, []byte("call restfull GET"))
    }
    ```
8. 注册控制器(src/Controller/Init.go)
    ```go
    package Controller

    import "github.com/pinguo/pgo"

    func init() {
        container := pgo.App.GetContainer()
        container.Bind(&WelcomeController{})
    }
    ```
9. 创建程序入口(src/Main/main.go)
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
10. 编译运行
    ```sh
    make update
    make start
    curl http://127.0.0.1:8000/welcome
    ```

### 其它
参见[pgo-docs](https://github.com/pinguo/pgo-docs)
