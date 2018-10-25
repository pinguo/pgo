# PGO
PGO应用框架即"Pinguo GO application framework"，是Camera360广告服务端团队研发的一款简单、高性能、组件化的GO应用框架。受益于GO语言高性能与原生协程支持，业务从php+yii2升级到PGO后，线上实际表现单机处理能力提高5-10倍，从实际使用中看其开发效率亦不输于PHP。

## 环境要求
- GO 1.10+
- Make 3.8+
- Linux/MacOS
- Glide 0.13+ (建议)
- GoLand 2018 (建议)

## 项目目录
规范：
- 项目的GOPATH为项目根目录，不要依赖系统的GOPATH。
- 除GO标准库外，所有外部依赖代码放到"src/vendor"下。
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
└── src/                # 项目源码目录
    ├── Command/        # 命令行控制器目录
    ├── Controller/     # HTTP控制器目录
    ├── Lib/            # 项目基础库目录
    ├── Main/           # 项目入口目录
    ├── Model/          # 模型目录(数据交互)
    ├── Service/        # 服务目录(业务逻辑)
    ├── Struct/         # 结构目录(数据定义)
    ├── Test/           # 测试目录(单测/性能)
    ├── glide.lock      # 项目依赖锁文件
    ├── glide.yaml      # 项目依赖配置文件
    └── vendor/         # 第三方依赖目录
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


## 使用示例
