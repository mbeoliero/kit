# kit

[![Go Reference](https://pkg.go.dev/badge/github.com/mbeoliero/kit.svg)](https://pkg.go.dev/github.com/mbeoliero/kit)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mbeoliero/kit)](https://go.dev/)
[![License](https://img.shields.io/github/license/mbeoliero/kit)](./LICENSE)

`kit` 是一组 Go 服务开发常用工具包，覆盖查询条件构建、数据库连接初始化、通用 Repository、Redis 结构封装、HTTP 客户端、日志和基础类型转换。

## Getting Started

安装依赖：

```bash
go get github.com/mbeoliero/kit
```

当前模块声明的 Go 版本为 `1.25.5`。使用前请确认本地 Go 版本满足项目要求。

## Demo

### 构建 MongoDB 查询条件

```go
package main

import (
	"fmt"

	"github.com/mbeoliero/kit/builder"
)

func main() {
	filter := builder.NewMongoQueryBuilder().
		Eq("status", "active").
		Gte("score", 60).
		Like("name", "tom", builder.MatchContains).
		Build()

	fmt.Printf("%#v\n", filter)
}
```

### 构建 GORM 查询条件

```go
package main

import (
	"context"

	"github.com/mbeoliero/kit/builder"
	"gorm.io/gorm"
)

type User struct {
	ID     uint
	Name   string
	Status string
}

func listActiveUsers(ctx context.Context, db *gorm.DB) ([]User, error) {
	cond := builder.NewGormQueryBuilder().
		Eq("status", "active").
		Like("name", "tom", builder.MatchContains).
		Build()

	var users []User
	err := db.WithContext(ctx).Where(cond).Find(&users).Error
	return users, err
}
```

### 使用泛型 Repository

```go
package main

import (
	"context"

	"github.com/mbeoliero/kit/builder"
	"github.com/mbeoliero/kit/repox"
	"gorm.io/gorm"
)

type User struct {
	ID     uint
	Name   string
	Status string
}

func findUser(ctx context.Context, db *gorm.DB, id uint) (*User, error) {
	repo := repox.NewGormRepo[User](db)

	return repo.FindOne(
		ctx,
		builder.Id(id),
		repox.Find().SetReturnFields("id", "name", "status"),
	)
}
```

### 使用 Redis 工具

```go
package main

import (
	"context"
	"time"

	"github.com/mbeoliero/kit/redisx"
	"github.com/redis/go-redis/v9"
)

func cacheValue(ctx context.Context, cli redis.UniversalClient) error {
	if err := redisx.SetByClient(ctx, cli, "user:1:name", "alice", time.Hour); err != nil {
		return err
	}

	name, err := redisx.GetByClient[string](ctx, cli, "user:1:name")
	if err != nil {
		return err
	}

	_ = name
	return nil
}
```

## Features

### `builder`

查询条件构建器，支持 MongoDB 和 GORM 两种输出：

- `Eq` / `Ne` / `Gt` / `Gte` / `Lt` / `Lte`
- `In` / `Nin`
- `Like`，支持 `MatchContains`、`MatchStartsWith`、`MatchEndsWith`
- `Id`
- `And` / `Or`

默认构建器是 MongoDB，可通过 `builder.SetQueryBuilder(builder.BuilderTypeGorm)` 切换全局默认构建器。

### `connector`

基础设施连接初始化：

- MySQL/GORM：单实例和读写分离模式，连接池配置，OpenTelemetry tracing 注入，SQL 日志包装。
- MongoDB：连接初始化，TLS 选项，OpenTelemetry command monitor。
- Redis：单实例和 Cluster 初始化，TLS、连接池、Tracing 和命令日志 hook。

### `repox`

面向 GORM 和 MongoDB 的泛型 Repository 抽象：

- `Create` / `CreateMany`
- `FindOne` / `Find` / `Count`
- `Update` / `UpdateOne` / `UpdateMany` / `Incr` / `UpsertOne`
- `DeleteOne` / `DeleteMany`
- `Transaction`
- `Native` 获取底层数据库对象

查询选项通过 `repox.Find()` 链式配置返回字段、分页和排序。

### `redisx`

Redis 常用数据结构和操作封装：

- KV：`Get`、`Set`、`Del`
- Counter：`Incr`、`Decr`
- HashMap：泛型字段和值读写、批量读写、自增、过期时间维护
- ZQueue：基于 Sorted Set 的队列封装，支持按分数范围查询、分页、弹出和删除

### `httpx`

基于 `resty.dev/v3` 的轻量 HTTP 客户端：

- JSON `POST`
- query 参数 `GET`
- Header 和 Bearer token 设置
- 请求与响应日志

### `log`

面向 CloudWeGo Kitex/Hertz 的日志封装：

- 默认 zerolog 实现，可切换 logrus
- 统一的 `Fatal`、`Error`、`Warn`、`Notice`、`Info`、`Debug`、`Trace` 方法
- Context 日志字段追加
- lumberjack 日志滚动文件输出
- 生产环境错误日志 Prometheus counter
- Kitex 和 Hertz logger 接入

### `utils`

基础工具包：

- `utils/typex`：字符串与泛型类型互转
- `utils/jsonx`：基于 sonic 的 JSON 字符串序列化
- `utils/filex`：当前文件路径获取
- `utils/generic`：泛型 `Once`

## Development

运行测试：

```bash
go test ./...
```

只验证某个包：

```bash
go test ./builder
go test ./redisx
```

查看包文档：

```bash
go doc ./builder
go doc ./repox
```

## Contributing

提交改动前请至少运行受影响包的测试。新增导出 API 时，同步补充 godoc 注释和必要的测试。

## License

本项目基于 [MIT License](./LICENSE) 发布。
