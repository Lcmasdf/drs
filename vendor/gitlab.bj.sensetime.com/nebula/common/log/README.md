# log

This repository is folked from [zap](https://github.com/uber-go/zap).

## level

```go
// DebugLevel logs are typically voluminous, and are usually disabled in
// production.
DebugLevel = core.DebugLevel
// InfoLevel is the default logging priority.
InfoLevel = core.InfoLevel
// WarnLevel logs are more important than Info, but don't need individual
// human review.
WarnLevel = core.WarnLevel
// ErrorLevel logs are high-priority. If an application is running smoothly,
// it shouldn't generate any error-level logs.
ErrorLevel = core.ErrorLevel
// DPanicLevel logs are particularly important errors which will include more specific stack message.
// In development the logger panics after writing the message.
DPanicLevel = core.DPanicLevel
// PanicLevel logs a message, then panics.
PanicLevel = core.PanicLevel
// FatalLevel logs a message, then calls os.Exit(1).
FatalLevel = core.FatalLevel
```

## easy to use.

```go
defer log.Sync()

log.Init(log.Config{
  Level:         "debug",
  DisableCaller: false,
})

log.With(log.String("one", "bar"), log.Int64("age", 21)).Infof("name is %v", "kim")
```

## 业务日志格式

```js
{
  level: "ERROR",
  time: "2020-06-17T19:32:27.015+0800",
  caller: "main/main.go:29@main.main",
  msg: "E10021: invalid params, ak is required",
  requestid: "", // requestid 默认从 ctx 中读出来，可以手动指定该字段的值。
  ctx: { // 写入 context 中的 value
    ak: "",
    storeid: "",
    deviceid: "",
    bindgroupid: ""
  },
  meta: {}, // 业务的复杂的辅助分析信息，一个 json 类型，但是在 es 中会设置为 text 类型。
  ... // 用户可以添加自定义的字段
}
```

## submodules

### trace

实现了 context 接口，同时方便封装 requestid，超时等业务信息

### log.Meta 方法

接收 object，key 固定为 meta

```go
type user struct {
  Name      string
  Email     string
  CreatedAt time.Time
}

func (u *user) MarshalLogObject(enc logcore.ObjectEncoder) error {
  enc.AddString("name", u.Name)
  enc.AddString("email", u.Email)
  enc.AddInt64("created_at", u.CreatedAt.UnixNano())
  return nil
}

log.With(log.Meta(&user{
  Name: "kim",
  Email: "zhanggguojin1@sensetime.com",
})).Info("this is message with meta")
```

## 跨服务传递 trace

包括 requestid, 全链路的超时

```go
getUserByID := func (ctx context.Context, userID int64) (user *User, err error) {
  user, err := userClient.GetUserByID(ctx, &user.Request{
    Header: &user.RequestHeader{
      Trace: logtrace.FromContext(ctx),
    },
    UserID: userID,
  })
  return
}

*****************

user service:

func GetUserByID(cctx context.Context, req *user.Request) (res *user.Response, err error) {
  // ctx 中包含 requestid 等 trace 信息，还包括过期时间
  ctx := logtrace.NewContext(ctx, req.Header.Trace)
  return
}
```

全链路的超时设置

```go
入口服务，设置整条链路的超时时间：

ctx := logtrace.NewContext(context.BackGround(), Trace{
  logtrace.KeyRequestid: "reqid",
  logtrace.KeyTimeout: "1000000000", // nanosecond
})
```
