<p align="center">
    <img src="./logo.png#gh-light-mode-only" alt="Goyave Logo" width="550"/>
    <img src="./logo_dark.png#gh-dark-mode-only" alt="Goyave Logo" width="550"/>
</p>

<p align="center">
    <a href="https://github.com/ixugo/goddd/releases"><img src="https://img.shields.io/github/v/release/ixugo/goweb?include_prereleases" alt="Version"/></a>
    <a href="https://github.com/ixugo/goddd/blob/master/LICENSE.txt"><img src="https://img.shields.io/dub/l/vibe-d.svg" alt="License"/></a>
	<a href="https://gin-gonic.com"><img width=30px  src="https://avatars.githubusercontent.com/u/7894478?s=48&v=4" alt="GIN"/></a>
    <a href="https://gorm.io"><img width=70px src="https://gorm.io/gorm.svg" alt="GORM"/></a>

</p>

# 企业 REST API 模板工具

用于自动生成 CRUD 代码

## 设计参考

[Google API Design Guide](https://google-cloud.gitbook.io/api-design-guide)

## 安装

在终端执行

```bash
go install github.com/ixugo/godddx@latest
go install mvdan.cc/gofumpt@latest
go install golang.org/x/tools/cmd/goimports@latest
```

## 流程

1. clone goweb 模板，或初始化项目 go mod init project
2. 创建 model.go 文件，写入结构体
   ```go
    // 包名即为生成的模块目录名
    package user

    type User struct {
        // 按照 gorm 的建议，应当包含  CreatedAt, UpdatedAt
        // goddx 生成的列表查询也会依赖 CreatedAt 查询排序
        ID int
        CreatedAt orm.Time
        UpdatedAt orm.Time

	    Name string // 昵称
	    Age  int64  //  年龄
    }
   ```
3. 执行 `godddx -f ./model.go` 即可生成代码
4. 在项目中调用 registerUser 函数，将生成的代码注册到 gin 路由上。

## 功能

- [x] 生成 5 项常用 CRUD (增删改查,分页搜索)
- [x] 生成 5 项常用 CRUD 缓存
- [ ] 生成 5 项常用 CRUD 的测试函数
- [ ] 生成 5 项常用 CRUD 的接口文档
- [ ] 支持分页查询中，前端传递排序方式
- [ ] 支持分页查询中，前端传递条件
- [ ] 生成 5 项常用的 redis 缓存代码

## 问题

> 为什么不读数据库生成代码?

平时在表中用 json 类型较多，读数据库没办法生成 json 结构体。

> 模型中定义了函数 CacheKey 做什么用的?

生成的缓存代码，必须知道键，才能到值，如果键有很多个，则删除的时候会很麻烦。
CacheKey 方法用来确定这个模型的唯一标识，默认应该是 ID，但如果不是通过 id 频繁查询，可以自行修改成其它键
例如 goddd 项目中的 token 实现，就是以 hash 为键。
