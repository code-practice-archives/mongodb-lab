# Lesson 08：MongoDB Transaction 事务

## 这一课要解决什么

前面我们已经知道：

> MongoDB 对**单个 Document 的写操作具有原子性**。

例如：

```javascript
db.orders.updateOne(
    { orderNo: "ORDER-001" },
    {
        $set: {
            status: "paid",
            totalPrice: 1200
        }
    }
)
```

虽然修改了多个 Field，但它们属于同一个 Document，所以整体可以原子完成。

因此：

```text
单 Document
    ↓
通常不需要 Transaction
```

问题出现在：

```text
多个 Document
+
必须一起成功或一起失败
```

这时候才需要 Transaction。

---

# 1. 为什么需要 Transaction

准备两个账户：

```javascript
{
    _id: "A",
    balance: 1000
}

{
    _id: "B",
    balance: 500
}
```

现在 A 给 B 转账 100：

```text
A - 100
B + 100
```

如果直接执行两个 Update：

```javascript
db.accounts.updateOne(
    { _id: "A" },
    { $inc: { balance: -100 } }
)

db.accounts.updateOne(
    { _id: "B" },
    { $inc: { balance: 100 } }
)
```

可能出现：

```text
A - 100
   ↓
成功
   ↓
程序异常
   ↓
B + 100 没执行
```

最终：

```text
A = 900
B = 500
```

数据就不一致了。

我们真正想要的是：

```text
Transaction
     ↓
A - 100
     ↓
B + 100
     ↓
 ┌───┴───┐
 ↓       ↓
成功    失败
 ↓       ↓
Commit  Abort
```

---

# 2. MongoDB 怎么知道这些操作属于同一个事务？

事务不是简单地把几条相邻命令看成一组。

MongoDB 需要知道：

```text
这些请求属于哪个 Session？

又属于这个 Session 的哪一个 Transaction？
```

所以事务内部会有类似：

```text
Session
+
txnNumber
```

这样的概念。

可以理解成：

```text
Session 1001
└── Transaction 7
    ├── Update A
    ├── Update B
    └── Commit
```

所以 MongoDB Transaction 的基本模型其实是：

```text
Session
   ↓
Transaction
   ↓
多个数据库操作
   ↓
Commit / Abort
```

---

# 3. 为什么 Transaction 要求 Replica Set？

如果我们还是运行前几课那种：

```text
Standalone MongoDB
```

多 Document Transaction 是不能使用的。

MongoDB 要求运行在：

```text
Replica Set

或

Sharded Cluster
```

这里非常容易产生一个误解：

> 事务需要 Replica Set，是不是因为事务必须有多个副本？

不是。

我们这一课只启动 **一个 MongoDB 节点**，事务照样可以运行。

真正原因是：

> MongoDB 把多 Document Transaction 建立在 Replica Set / Sharded Cluster 提供的 Session、事务状态、Primary、oplog 和 Commit/Abort 等运行机制之上。

所以：

```text
Transaction
    ↓
依赖 MongoDB 的事务运行框架
    ↓
Replica Set / Sharded Cluster
```

而不是：

```text
Transaction
    ↓
必须拥有多个数据副本
```

---

# 4. 为什么一个节点也可以？

真实生产 Replica Set 通常类似：

```text
        Primary
       /       \
Secondary     Secondary
```

用于：

```text
数据冗余
故障切换
高可用
```

但我们现在只是学习事务。

所以使用：

```text
Replica Set rs0
└── Primary
```

它：

```text
✓ 可以运行 Transaction
✓ 有 Replica Set 运行模式

✗ 没有数据冗余
✗ 没有真正高可用
```

所以这一课启动 Replica Set，只是为了让 MongoDB 进入支持事务的运行模式。

---

# 5. 创建独立实验目录

```text
mongodb-lab/
├── ...
├── lesson-07-modeling/
└── lesson-08-transaction/
```

创建：

```bash
mkdir lesson-08-transaction
cd lesson-08-transaction
```

最终：

```text
lesson-08-transaction/
├── compose.yaml
├── go.mod
├── go.sum
└── main.go
```

---

# 6. 启动单节点 Replica Set

`compose.yaml`：

```yaml
services:
  mongodb:
    image: mongodb/mongodb-community-server:8.0.29-ubi9-slim
    container_name: mongodb-lesson-08
    hostname: mongodb

    ports:
      - "27024:27017"

    command:
      - --replSet
      - rs0

    volumes:
      - mongodb_lesson_08_data:/data/db

volumes:
  mongodb_lesson_08_data:
```

和前面的课程相比，关键新增：

```yaml
command:
  - --replSet
  - rs0
```

意思是：

> 让 MongoDB 以 `rs0` Replica Set 成员的身份运行。

启动：

```bash
docker compose up -d
```

查看：

```bash
docker compose ps
```

有问题时：

```bash
docker compose logs mongodb
```

---

# 7. 初始化 Replica Set

启动参数只是告诉 MongoDB：

> 我要运行 Replica Set。

还需要真正初始化成员。

连接：

```bash
mongosh "mongodb://localhost:27024"
```

执行：

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

看到：

```text
stateStr: "PRIMARY"
```

说明初始化成功。

现在环境：

```text
Replica Set rs0
       │
       └── mongodb
              │
              └── PRIMARY
```

---

# 8. 准备事务实验数据

进入：

```javascript
use middleware_lab
```

清空旧数据：

```javascript
db.accounts.deleteMany({})
```

插入：

```javascript
db.accounts.insertMany([
    {
        _id: "A",
        balance: 1000
    },
    {
        _id: "B",
        balance: 500
    }
])
```

查询：

```javascript
db.accounts.find()
```

当前：

```text
A = 1000
B = 500
```

---

# 9. mongosh：手动执行第一个事务

这一步是这一课最重要的实践。

先创建 Session：

```javascript
session = db.getMongo().startSession()
```

从这个 Session 获取 Database：

```javascript
txDB = session.getDatabase("middleware_lab")
```

现在：

```text
db
```

是普通 Database。

而：

```text
txDB
```

是在这个 Session 中使用的 Database。

---

## 开始事务

```javascript
session.startTransaction()
```

现在事务已经开始。

---

## A 扣 100

```javascript
txDB.accounts.updateOne(
    { _id: "A" },
    {
        $inc: {
            balance: -100
        }
    }
)
```

---

## B 加 100

```javascript
txDB.accounts.updateOne(
    { _id: "B" },
    {
        $inc: {
            balance: 100
        }
    }
)
```

---

## 提交事务

```javascript
session.commitTransaction()
```

普通查询：

```javascript
db.accounts.find()
```

结果：

```text
A = 900
B = 600
```

整个过程：

```text
startSession
      ↓
startTransaction
      ↓
A - 100
      ↓
B + 100
      ↓
commitTransaction
```

---

# 10. 手动体验 Abort / Rollback

现在再做一次实验。

创建一个新的事务：

```javascript
session.startTransaction()
```

先执行：

```javascript
txDB.accounts.updateOne(
    { _id: "A" },
    {
        $inc: {
            balance: -100
        }
    }
)
```

但这次我们假设程序发现问题了。

不要 Commit。

执行：

```javascript
session.abortTransaction()
```

查询：

```javascript
db.accounts.find()
```

你会发现仍然：

```text
A = 900
B = 600
```

并没有变成：

```text
A = 800
B = 600
```

因为：

```text
A - 100
   ↓
Abort
   ↓
这次修改被回滚
```

---

# 11. 到这里应该真正理解事务生命周期

现在我们已经亲手执行过：

```text
Session
   ↓
startTransaction
   ↓
数据库操作
   ↓
数据库操作
   ↓
commitTransaction
```

以及：

```text
Session
   ↓
startTransaction
   ↓
数据库操作
   ↓
发现问题
   ↓
abortTransaction
```

所以：

```text
Commit
→ 事务中的修改生效

Abort
→ 事务中的修改全部放弃
```

---

# 12. 结束 Session

实验结束可以：

```javascript
session.endSession()
```

---

# 13. 现在再看 Go Driver

到这里 Go 的事务代码就很好理解了。

先初始化项目：

```bash
go mod init mongodb-lab/lesson-08-transaction

go get go.mongodb.org/mongo-driver/v2/mongo
```

连接：

```go
client, err := mongo.Connect(
    options.Client().ApplyURI(
        "mongodb://localhost:27024/?replicaSet=rs0&directConnection=true",
    ),
)
```

---

# 14. Go 中的 Session

mongosh：

```javascript
session = db.getMongo().startSession()
```

Go：

```go
session, err := client.StartSession()
```

结束：

```go
defer session.EndSession(ctx)
```

---

# 15. Go 中的 Transaction

mongosh 手动写：

```text
startTransaction
     ↓
执行操作
     ↓
commitTransaction

或

abortTransaction
```

Go Driver 提供了更方便的：

```go
session.WithTransaction(...)
```

基本结构：

```go
_, err = session.WithTransaction(
    ctx,
    func(txCtx context.Context) (any, error) {

        // MongoDB 操作

        return nil, nil
    },
)
```

可以简单理解成：

```text
WithTransaction
       ↓
Start Transaction
       ↓
执行 callback
       ↓
callback 返回 error？
    /            \
  没有            有
   ↓              ↓
Commit           Abort
```

所以：

> `WithTransaction()` 并不是另一套事务机制，只是 Driver 把刚才在 mongosh 手动做的流程封装起来。

---

# 16. 完整 Go Demo

`main.go`：

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Account struct {
	ID      string `bson:"_id"`
	Balance int    `bson:"balance"`
}

func main() {
	ctx := context.Background()

	client, err := mongo.Connect(
		options.Client().ApplyURI(
			"mongodb://localhost:27024/?replicaSet=rs0&directConnection=true",
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	accounts := client.
		Database("middleware_lab").
		Collection("accounts")

	// 每次运行重新准备数据
	_, err = accounts.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = accounts.InsertMany(
		ctx,
		[]any{
			Account{
				ID:      "A",
				Balance: 1000,
			},
			Account{
				ID:      "B",
				Balance: 500,
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Initial ===")
	printAccounts(ctx, accounts)

	// 故意失败
	fmt.Println("\n=== Failed Transaction ===")

	err = transferWithError(
		ctx,
		client,
		accounts,
		"A",
		"B",
		100,
	)

	fmt.Println("error:", err)

	fmt.Println("\n=== After Failed Transaction ===")
	printAccounts(ctx, accounts)

	// 正常事务
	fmt.Println("\n=== Successful Transaction ===")

	err = transfer(
		ctx,
		client,
		accounts,
		"A",
		"B",
		100,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n=== After Successful Transaction ===")
	printAccounts(ctx, accounts)
}
```

---

# 17. 正常事务

```go
func transfer(
	ctx context.Context,
	client *mongo.Client,
	accounts *mongo.Collection,
	from string,
	to string,
	amount int,
) error {

	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(
		ctx,
		func(txCtx context.Context) (any, error) {

			// A 扣钱
			result, err := accounts.UpdateOne(
				txCtx,
				bson.M{
					"_id":     from,
					"balance": bson.M{"$gte": amount},
				},
				bson.M{
					"$inc": bson.M{
						"balance": -amount,
					},
				},
			)

			if err != nil {
				return nil, err
			}

			if result.MatchedCount == 0 {
				return nil, errors.New(
					"账户不存在或余额不足",
				)
			}

			// B 加钱
			result, err = accounts.UpdateOne(
				txCtx,
				bson.M{
					"_id": to,
				},
				bson.M{
					"$inc": bson.M{
						"balance": amount,
					},
				},
			)

			if err != nil {
				return nil, err
			}

			if result.MatchedCount == 0 {
				return nil, errors.New(
					"目标账户不存在",
				)
			}

			return nil, nil
		},
	)

	return err
}
```

---

# 18. 故意失败验证回滚

```go
func transferWithError(
	ctx context.Context,
	client *mongo.Client,
	accounts *mongo.Collection,
	from string,
	to string,
	amount int,
) error {

	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(
		ctx,
		func(txCtx context.Context) (any, error) {

			// A - 100
			_, err := accounts.UpdateOne(
				txCtx,
				bson.M{
					"_id": from,
				},
				bson.M{
					"$inc": bson.M{
						"balance": -amount,
					},
				},
			)

			if err != nil {
				return nil, err
			}

			// 模拟中途异常
			return nil, errors.New(
				"模拟程序异常",
			)
		},
	)

	return err
}
```

---

# 19. 打印数据

```go
func printAccounts(
	ctx context.Context,
	accounts *mongo.Collection,
) {

	cursor, err := accounts.Find(
		ctx,
		bson.M{},
		options.Find().SetSort(
			bson.D{
				{Key: "_id", Value: 1},
			},
		),
	)

	if err != nil {
		log.Fatal(err)
	}

	defer cursor.Close(ctx)

	var result []Account

	if err := cursor.All(ctx, &result); err != nil {
		log.Fatal(err)
	}

	for _, account := range result {
		fmt.Printf(
			"%s = %d\n",
			account.ID,
			account.Balance,
		)
	}
}
```

---

# 20. 运行结果

```bash
go run .
```

失败事务：

```text
=== Initial ===
A = 1000
B = 500

=== Failed Transaction ===
error: 模拟程序异常

=== After Failed Transaction ===
A = 1000
B = 500
```

说明：

```text
A - 100
虽然执行了
    ↓
事务最终 Abort
    ↓
修改没有生效
```

正常事务：

```text
=== Successful Transaction ===

=== After Successful Transaction ===
A = 900
B = 600
```

说明：

```text
A - 100
B + 100
    ↓
Commit
    ↓
一起生效
```

---

# 21. mongosh 与 Go Driver 对照

| mongosh | Go Driver |
|---|---|
| `startSession()` | `client.StartSession()` |
| `startTransaction()` | `WithTransaction()` 内部处理 |
| 数据库操作 | 数据库操作 |
| `commitTransaction()` | `WithTransaction()` 成功时处理 |
| `abortTransaction()` | `WithTransaction()` 失败时处理 |
| `endSession()` | `session.EndSession()` |

所以真正的关系是：

```text
MongoDB Transaction
        ↓
mongosh 可以手动操作
        ↓
Go Driver 只是进一步封装
```

---

# 22. 本课最终认识

首先：

```text
单 Document
    ↓
MongoDB 保证原子性
    ↓
通常不需要 Transaction
```

当：

```text
多个 Document
+
必须保持一致
```

才引入：

```text
Transaction
```

事务生命周期：

```text
Session
   ↓
Start Transaction
   ↓
多个操作
   ↓
 ┌─────┴─────┐
 ↓           ↓
Commit      Abort
```

MongoDB 的多 Document Transaction 又要求：

```text
Replica Set
或
Sharded Cluster
```

因为 MongoDB 将事务建立在这些运行模式提供的 Session、事务状态、Primary、oplog 和事务管理机制之上。

学习环境：

```text
Replica Set rs0
└── 一个 Primary
```

已经足够体验 Transaction。

最后再记住：

> **MongoDB 是先利用 Document 建模和单 Document 原子性；确实存在跨 Document 一致性需求时，再使用 Transaction。**

而 Go 的：

```go
session.WithTransaction(...)
```

只是把我们已经在 `mongosh` 亲手执行过的：

```text
start
→ 执行
→ commit / abort
```

封装了起来。