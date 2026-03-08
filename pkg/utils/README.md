# pkg/utils

与**业务无关**的通用工具库，供全项目（core、main_server 等）使用。

## 约定

- **工具集中在此处**：时间、随机、切片、类型转换、ID、敏感词、加密等纯逻辑都放在 `pkg/utils`（或子包）。
- **core/common**：仅保留业务侧兼容入口，内部委托本包；新代码应直接引用 `xfx/pkg/utils`。

---

## 文件与使用情况

| 文件/子包 | 说明 | 项目内使用情况 |
|-----------|------|----------------|
| **clock.go** | 游戏逻辑时间源 | `Now()` 返回 真实时间 + 偏移；偏移由 GM 后台 POST `/gm/time/set_offset` 设置。业务取“当前游戏时间”请用 `utils.Now()`。 |
| **time.go** | 时间戳、自然日/周/月判断 | 内部已统一用 `Now()`。**常用**：`CheckIsSameDayBySec`、`DaysDiff`、`GetTodayEndMinUnix`、`GetTodayEndUnixInHour`、`GetTodayUnixInHour`。其余保留作扩展。 |
| **random.go** | 随机数、加权随机、概率 | **常用**：`RandInt`、`WeightedRandom`、`WeightIndex`、`MicsSlice`、`SelectByOdds`。`Random(length)`、`UpdateRand` 当前未用，保留。 |
| **randString.go** | 随机字符串（多种策略） | **当前未使用**。若只需简单随机串，可只保留 `RandomAlphanumeric` / `RandomNumeric`，其余按需删除。 |
| **slice.go** | 切片包含、去重、删除 | **常用**（多通过 core/common 调用）：`ContainsInt32`、`ContainsString`、`ContainsAllInt32`、`RemoveFirstInt32`、`HasDuplicateInt32`。 |
| **conv.go** | 类型转换 | **常用**：`ParseInt64`（core/common 的 `StringToInt64` 内部使用）。 |
| **reply.go** | Redis/redigo 风格 reply→基础类型 | **当前未使用**。保留供后续 Redis 客户端或类似场景使用。 |
| **math.go** | 数值边界（Min/Max/Clamp） | 新增，通用小工具。 |
| **id/** | 雪花 ID、短 ID 编解码 | **使用**：`id.Init`（main 启动）。`General()`、`Itoa`/`Atoi` 供业务按需使用。 |
| **sensitive/** | 敏感词过滤 | **使用**：main_server/player。 |
| **crypto/** | AES 等加密 | 当前仅注释引用，保留备用。 |

---

## 建议

1. **可删或精简**：若确定不需要随机字符串，可删 `randString.go` 或只保留 1～2 个函数；`reply.go` 若长期不用可删，需要 Redis 再补回。
2. **可补充**：已增加 `math.go`（Min/Max/Clamp）；若业务常做“取两数小/大”、“限制在区间”可直接用。其他可按需加（如 MustParseInt、截断字符串等）。
