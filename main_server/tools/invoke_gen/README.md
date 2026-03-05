# invoke_gen

根据 `main_server/logic` 下各 module 的 `Register` 调用，自动生成 `main_server/invoke` 的强类型 client 代码。

## 用法

在**项目根目录**执行：

```bash
go run ./main_server/tools/invoke_gen
```

会扫描 `main_server/logic` 下的直接子包（如 `activity`、`common`、`login`），要求：

- 包内存在 `GetType() string` 且返回 `define.ModuleXxx`
- 包内在 `OnInit` 等位置调用了 `Register("方法名", 方法值)`

然后为每个这样的包生成对应的 `main_server/invoke/<包名>.go`，包含：

- `XxxModClient` 结构体与 `XxxClient(invoker)` 构造函数
- 每个注册方法的强类型包装，内部通过 `Invoker.Invoke(module, "方法名", args...)` 调用，并用 `As[T]` 做安全类型断言，避免参数/返回值类型不一致时 panic

## 新增 logic 子包时

1. 在新包中实现 module（含 `GetType()` 和 `Register(...)`）。
2. 在项目根目录执行：`go run ./main_server/tools/invoke_gen`。
3. 会在 `main_server/invoke/` 下生成新的 `<包名>.go`，无需手写 client。

## 依赖

- `golang.org/x/tools/go/packages`（用于加载与解析 logic 包）
