# go-homework

Go 语言课程课后作业集合（homework01 ~ homework05）。

从基础语法、并发编程，一路做到 GORM 持久化、Gin Web 服务与以太坊链上交互。

## 目录结构

```
go-homework/
├── homework01/     Go 基础算法题（8 题）+ 单元测试
├── homework02/     Go 基础语法与并发练习（10 题）
├── homework03/     GORM + SQLite 博客系统（模型 / 关联查询 / 钩子函数）
├── homework04/     Gin + GORM + JWT 博客 API 项目（分层架构）
└── homework05/     Sepolia 链上交互 + Counter 合约（go-ethereum / abigen）
```

---

## homework01 — 基础算法题

`homework.go` 中实现了 8 道常见算法题，配套 `homework_test.go` 单元测试。

| # | 题目 | 函数 |
|---|------|------|
| 1 | 只出现一次的数字 | `SingleNumber` |
| 2 | 回文数 | `IsPalindrome` |
| 3 | 有效的括号 | `IsValid` |
| 4 | 最长公共前缀 | `LongestCommonPrefix` |
| 5 | 加一 | `PlusOne` |
| 6 | 删除有序数组中的重复项 | `RemoveDuplicates` |
| 7 | 合并区间 | `Merge` |
| 8 | 两数之和 | `TwoSum` |

运行测试：

```bash
cd homework01
go test -v ./...
```

## homework02 — 基础语法与并发

10 个独立的小练习，每个文件都是一个完整的 `main` 程序，需逐个运行：

```bash
cd homework02
go run exercise01.go     # 例如第 1 题
```

| # | 题目 |
|---|------|
| 01 | 指针参数：在函数内将指针指向的值 +10 |
| 02 | 切片指针：将切片中每个元素乘 2 |
| 03 | 启动两个协程，分别打印 1~10 的奇数与偶数 |
| 04 | 任务调度器：并发执行一组任务并统计各自耗时 |
| 05 | `Shape` 接口与 `Rectangle` / `Circle` 实现 |
| 06 | 结构体组合：`Employee` 组合 `Person`，实现 `PrintInfo()` |
| 07 | 用通道实现两个协程通信（生产者 1~10） |
| 08 | 带缓冲通道的生产者—消费者模型（100 个整数） |
| 09 | 用 `sync.Mutex` 保护共享计数器（10 协程 × 1000 次） |
| 10 | 用 `sync/atomic` 实现无锁计数器（10 协程 × 1000 次） |

> 注意：本目录没有 `go.mod`，各文件是彼此独立的单文件程序，请用 `go run 文件名.go` 逐个运行（同一目录下多个 `main` 函数不能一起编译）。

## homework03 — GORM 与 SQLite

以博客系统（`User` / `Post` / `Comment`）为例，练习 GORM 的三块内容：

- **题目 1 · 模型定义**：一对多关系建模（User→Post、Post→Comment），并建表
- **题目 2 · 关联查询**：查询某用户的全部文章及评论；查询评论数最多的文章
- **题目 3 · 钩子函数**：`Post` 创建时更新用户的文章计数；`Comment` 删除后若无评论则把文章状态置为「无评论」

```bash
cd homework03
go run exercise01.go
```

`homework03.db` 由程序运行时自动创建并迁移，不入库。

## homework04 — Gin 博客 API

一个分层架构的博客后端，包含用户注册登录、文章与评论的增删改查。

```
homework04/
├── main.go                  程序入口（优雅关闭）
├── internal/
│   ├── config/              配置加载（环境变量 + 默认值）
│   ├── database/            数据库初始化、迁移与演示数据
│   ├── handlers/            HTTP 处理器（auth / post / comment / dto）
│   ├── logging/             结构化日志
│   ├── middleware/          日志与 JWT 认证中间件
│   ├── models/              GORM 数据模型
│   ├── router/              路由注册
│   └── utils/               JWT、统一响应、错误、参数工具
├── gin的洋葱模型.md          中间件执行顺序笔记
├── 认证全流程.md             注册/登录/JWT 校验全流程图解
└── 数据库ER图.md             数据表关系说明
```

```bash
cd homework04
go run main.go
```

可用环境变量配置：`PORT`、`DB_PATH`、`JWT_SECRET`、`JWT_TTL`。

> `JWT_SECRET` 默认值为 `dev-secret`，仅供本地开发使用，部署时请务必通过环境变量覆盖。
> 演示账号由 `internal/database` 在初始化时写入：用户名 `demo` / 密码 `demo123456`。

## homework05 — Sepolia 区块链交互

使用 `go-ethereum` 连接 Sepolia 测试网：查询区块、发送 ETH 转账、部署 `Counter` 合约，并通过 `abigen` 生成的绑定代码读写合约。

```
homework05/
├── cmd/account/              查询账户余额、nonce、账户类型
├── cmd/block/                查询指定区块（哈希、时间戳、交易等）
├── cmd/transfer/             构造、签名并发送 ETH 转账
├── cmd/deploy/               部署 Counter，地址写入 deployments/
├── cmd/counter/              读取 count、increment、set、查询事件
├── contracts/Counter.sol     计数器合约源码（同目录含 ABI / 字节码）
├── bindings/counter/         abigen 生成的 Go 绑定
└── .env                      RPC 地址与私钥（私钥为占位符）
```

```bash
cd homework05
go run ./cmd/account
go run ./cmd/block -number 12020312
go run ./cmd/transfer -to 0x... -amount 0.001
go run ./cmd/deploy
go run ./cmd/counter get
go run ./cmd/counter increment
go run ./cmd/counter set 42
go run ./cmd/counter events
```

> `.env` 中只提交私钥占位符 `0xYOUR_PRIVATE_KEY_HERE`；运行写操作前请在本地替换成自己的 Sepolia 测试账户私钥，不要把真实私钥提交到 git。

部署记录保存在 `homework05/deployments/counter-11155111.json`。

---

## 环境

- Go 1.27.1
- homework03 / homework04 依赖 GORM、Gin、`github.com/glebarez/sqlite`（纯 Go 的 SQLite 驱动，无需 CGO）
- homework05 依赖 `github.com/ethereum/go-ethereum`；合约用 solc 编译，Go 绑定由 abigen 生成
