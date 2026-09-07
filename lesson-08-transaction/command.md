# MongoDB Lesson 08：Transaction 手册

## 核心认识

MongoDB：

```text
单 Document
→ 写操作本身原子
→ 通常不需要 Transaction
```

当多个 Document 必须一起成功或一起失败时：

```text
多个 Document
→ Transaction
→ Commit / Abort
```

MongoDB 多 Document Transaction 要求运行在：

```text
Replica Set
或
Sharded Cluster
```

并不是因为事务必须有多个副本，而是因为 MongoDB 的事务机制建立在这些运行模式提供的 Session、事务状态、Primary、oplog 等基础设施之上。

本地学习可以只运行：

```text
Replica Set rs0
└── 1 个 Primary
```

---

## Replica Set 最小配置

Compose：

```yaml
command:
  - --replSet
  - rs0
```

初始化：

```javascript
rs.initiate({
    _id: "rs0",
    members: [
        {
            _id: 0,
            host: "mongodb:27017"
        }
    ]
})
```

查看：

```javascript
rs.status()
```

确认节点为：

```text
PRIMARY
```

---

## mongosh 手动事务

创建 Session：

```javascript
session = db.getMongo().startSession()
```

获取 Session 中的 Database：

```javascript
txDB = session.getDatabase("middleware_lab")
```

开始事务：

```javascript
session.startTransaction()
```

执行操作：

```javascript
txDB.accounts.updateOne(
    { _id: "A" },
    { $inc: { balance: -100 } }
)

txDB.accounts.updateOne(
    { _id: "B" },
    { $inc: { balance: 100 } }
)
```

提交：

```javascript
session.commitTransaction()
```

回滚：

```javascript
session.abortTransaction()
```

结束 Session：

```javascript
session.endSession()
```

事务生命周期：

```text
Session
  ↓
startTransaction
  ↓
多个 CRUD
  ↓
commitTransaction
或
abortTransaction
```

---

## Go Transaction

核心 API：

| API | 作用 |
|---|---|
| `client.StartSession()` | 创建 Session |
| `session.WithTransaction()` | 执行并管理事务 |
| `session.EndSession()` | 结束 Session |

基本结构：

```go
session, err := client.StartSession()
if err != nil {
    return err
}
defer session.EndSession(ctx)

_, err = session.WithTransaction(
    ctx,
    func(txCtx context.Context) (any, error) {

        // 多个 MongoDB 操作

        return nil, nil
    },
)
```

`WithTransaction()` 可以理解为帮我们封装：

```text
startTransaction
      ↓
执行 callback
      ↓
成功 → Commit
失败 → Abort
```

---

## mongosh 与 Go 对照

| mongosh | Go Driver |
|---|---|
| `startSession()` | `StartSession()` |
| `startTransaction()` | `WithTransaction()` 内部处理 |
| CRUD | CRUD |
| `commitTransaction()` | 成功时自动处理 |
| `abortTransaction()` | 失败时自动处理 |
| `endSession()` | `EndSession()` |

---

## 本课重点

```text
单 Document
→ 利用原子操作

多个 Document 必须保持一致
→ Transaction

Transaction
→ Session
→ Start
→ 多个操作
→ Commit / Abort

MongoDB Transaction
→ 需要 Replica Set / Sharded Cluster
```

最重要的一句话：

> **先利用合理的 Document 建模和单 Document 原子性；确实存在跨 Document 一致性需求时，再使用 Transaction。**