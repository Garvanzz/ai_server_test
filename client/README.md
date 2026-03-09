# 模拟客户端

用于联调/压测：先登录 **login_server** 获取 token，再连接 **main_server** TCP 发送 `C2SLogin`，之后按间隔随机请求部分游戏接口。

## 流程

1. **登录服**（HTTP）：可选注册 → `POST /login`，拿到 `token`、`uid`、`serverId`
2. **游戏服**（TCP）：连接配置的 main 地址 → 发 `C2SLogin{Token}` → 收到 `S2CLogin` 后按间隔随机发无参/默认参 C2S（背包、任务、邮件、排行榜、商店等）

## 用法

在项目根目录执行：

```bash
# 默认：1 个客户端，先注册再登录，每 2s 随机打一次接口
go run ./client/

# 指定登录服与游戏服
go run ./client/ -login http://127.0.0.1:9033 -main 127.0.0.1:8082

# 多客户端（账号 prefix_1, prefix_2, ...）
go run ./client/ -n 3 -prefix test_user -interval 3s

# 单账号密码（不自动注册）
go run ./client/ -account myuser -password mypass -register=false

# 只跑 30 秒
go run ./client/ -duration 30s
```

## 参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-login` | `http://127.0.0.1:9033` | 登录服 HTTP 地址 |
| `-main` | `127.0.0.1:8082` | 游戏服 TCP 地址 |
| `-server` | `1` | 区服 ID |
| `-account` | 空 | 单客户端时使用的账号 |
| `-password` | 空 | 单客户端时使用的密码 |
| `-prefix` | `test_user` | 多客户端账号前缀（账号为 prefix_1, prefix_2...） |
| `-n` | `1` | 并发客户端数量 |
| `-register` | `true` | 是否先调用注册再登录 |
| `-interval` | `2s` | 随机请求间隔 |
| `-duration` | `0` | 运行时长，0 表示一直运行直到 Ctrl+C |

## 随机请求的接口（无参/默认参）

- 背包、任务、玩家属性、邮件列表、排行榜、关卡、商店
- 签到、日奖励、英雄/阵容/图鉴/技能初始化
- 挂机宝箱、好友列表、月卡/在线奖励/功能开放初始化等

Ctrl+C 优雅退出，等待所有客户端断开后进程退出。
