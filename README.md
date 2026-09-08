
# MongoDB 学习总结

这套课程的目标不是深入 MongoDB 底层实现，而是做到：

```text
知道它是什么
→ 会基本操作
→ 能用 Go 接入
→ 理解核心架构
→ 知道什么时候适合使用
```

---

# 一、核心数据模型

MongoDB 最基础的数据层级：

```text
Database
   ↓
Collection
   ↓
Document
   ↓
Field
```

和 MySQL 可以粗略类比：

| MongoDB | MySQL |
|---|---|
| Database | Database |
| Collection | Table |
| Document | Row |
| Field | Column |
| `_id` | Primary Key |

核心认识：

- MongoDB 以 **Document** 为主要数据单位。
- Document 实际使用 BSON 存储。
- 同一个 Collection 中的 Document 可以拥有不同字段，即 Flexible Schema。
- Document 可以自然包含嵌套 Document 和 Array。

---

# 二、基础 CRUD

核心操作：

```text
Create
→ insertOne / insertMany

Read
→ find / findOne

Update
→ updateOne / updateMany

Delete
→ deleteOne / deleteMany
```

最重要的概念是：

```text
Filter
```

例如：

```javascript
{
    category: "toy",
    price: {
        $lt: 1000
    }
}
```

Filter 可以同时用于：

```text
find
update
delete
```

常见比较操作符：

| 操作符 | 含义 |
|---|---|
| `$gt` | 大于 |
| `$gte` | 大于等于 |
| `$lt` | 小于 |
| `$lte` | 小于等于 |
| `$ne` | 不等于 |
| `$in` | 包含于指定集合 |

更新常见：

```text
$set
→ 设置值

$inc
→ 数值增加

$push
→ 向数组追加元素
```

---

# 三、查询与 Cursor

Go Driver 中：

```text
FindOne
→ 返回一条

Find
→ 返回 Cursor
```

Cursor 不是所有数据本身，而是：

> 查询结果的读取游标。

基本流程：

```text
Find
 ↓
Cursor
 ↓
All / Next
```

两种读取方式：

```text
cursor.All()
→ 一次读取全部剩余结果
→ 适合结果较小

cursor.Next()
→ 逐条处理
→ 更适合大量数据
```

常见查询选项：

| API | 类似 SQL |
|---|---|
| `SetSort()` | `ORDER BY` |
| `SetSkip()` | `OFFSET` |
| `SetLimit()` | `LIMIT` |
| `SetProjection()` | 指定 SELECT 字段 |

---

# 四、BSON 在 Go 中的表达

最常见：

```go
bson.M
```

适合：

```text
Filter
普通 Document
Update
```

例如：

```go
bson.M{
    "category": "toy",
}
```

需要明确顺序时常使用：

```go
bson.D
```

例如排序：

```go
bson.D{
    {Key: "price", Value: -1},
    {Key: "name", Value: 1},
}
```

Go Struct 可以直接映射 MongoDB Document：

```go
type Product struct {
    Name  string `bson:"name"`
    Price int    `bson:"price"`
}
```

---

# 五、索引

索引解决的核心问题：

> 减少查询时无效的 Document 扫描。

没有合适索引：

```text
COLLSCAN
```

有合适索引：

```text
IXSCAN
```

主要学习了三种：

```text
单字段索引
联合索引
唯一索引
```

创建：

```javascript
db.products.createIndex({
    sku: 1
})
```

联合：

```javascript
db.products.createIndex({
    category: 1,
    price: 1
})
```

唯一：

```javascript
db.products.createIndex(
    { sku: 1 },
    { unique: true }
)
```

重要排查命令：

```javascript
.explain("executionStats")
```

重点关注：

```text
COLLSCAN / IXSCAN

totalDocsExamined

totalKeysExamined

nReturned
```

索引的基本取舍：

```text
提高查询效率
+
增加存储和写入维护成本
```

---

# 六、Aggregation 聚合

MongoDB 聚合采用：

```text
Aggregation Pipeline
```

Document 会依次经过多个 Stage：

```text
Documents
   ↓
$match
   ↓
$group
   ↓
$sort
   ↓
$project
```

和 SQL 可以粗略对应：

| MongoDB | SQL |
|---|---|
| `$match` | `WHERE` |
| `$group` | `GROUP BY` |
| `$sort` | `ORDER BY` |
| `$project` | `SELECT` |
| `$limit` | `LIMIT` |
| `$sum` | `SUM / COUNT` |
| `$avg` | `AVG` |

Go 中：

```go
collection.Aggregate(ctx, pipeline)
```

同样返回：

```text
Cursor
```

核心不是记大量操作符，而是理解：

> 每个 Stage 都是在加工上一个 Stage 的结果。

---

# 七、Document 数据建模

MongoDB 关联数据主要有两种组织方式：

```text
Embedded
Reference
```

## Embedded

把相关数据直接放进一个 Document：

```text
Product
├── name
└── attributes
      ├── cpu
      └── memory
```

适合：

```text
经常一起读取
生命周期一致
数据规模可控
```

## Reference

拆到不同 Collection，通过 ID 建立关联：

```text
Product.categoryId
        ↓
Category._id
```

适合：

```text
数据独立存在
多个 Document 共享
数据可能持续增长
```

MongoDB 建模的重要思维不是：

```text
应该拆几张表？
```

而是：

```text
数据平时怎么读？
是否一起修改？
生命周期是否一致？
数据会不会无限增长？
```

---

# 八、Transaction 事务

MongoDB 单 Document 写操作本身具有原子性。

因此：

```text
单 Document
→ 通常不需要 Transaction
```

当：

```text
多个 Document
+
必须一起成功或一起失败
```

才需要：

```text
Transaction
```

事务生命周期：

```text
Session
   ↓
startTransaction
   ↓
多个操作
   ↓
Commit / Abort
```

mongosh 核心命令：

```javascript
session = db.getMongo().startSession()

txDB = session.getDatabase("middleware_lab")

session.startTransaction()

session.commitTransaction()

session.abortTransaction()

session.endSession()
```

Go Driver：

```go
client.StartSession()

session.WithTransaction(...)

session.EndSession(...)
```

`WithTransaction()` 本质是把：

```text
start
→ 执行
→ commit / abort
```

封装起来。

---

# 九、Replica Set

Replica Set 解决：

```text
数据冗余
高可用
Primary 故障切换
```

基本架构：

```text
        Primary
       /       \
Secondary     Secondary
```

核心角色：

| 概念 | 作用 |
|---|---|
| Primary | 接收写入 |
| Secondary | 保存数据副本 |
| oplog | 记录并复制数据变化 |
| Election | Primary 故障后的重新选举 |

复制过程：

```text
Write
 ↓
Primary
 ↓
oplog
 ↓
Secondary
```

Primary 故障：

```text
Primary DOWN
     ↓
Election
     ↓
New Primary
```

Go Driver 连接 Replica Set：

```text
mongodb://mongo1,mongo2,mongo3/?replicaSet=rs0
```

Driver 会自动发现当前 Primary。

---

# 十、为什么事务需要 Replica Set

MongoDB 多 Document Transaction 不支持 Standalone，需要：

```text
Replica Set
或
Sharded Cluster
```

不是因为：

```text
事务必须有多个数据副本
```

而是因为 MongoDB 的事务能力建立在这些运行模式提供的：

```text
Session
txnNumber
Primary
oplog
事务状态
Commit / Abort
```

等机制上。

因此学习事务时：

```text
单节点 Replica Set
```

就已经可以运行 Transaction。

只是没有真正的高可用。

---

# 十一、Sharding

Sharding 解决：

```text
单节点容量不够
单节点吞吐不够
        ↓
水平扩容
```

核心区别：

```text
Replica Set
→ Copy
→ 同一份数据复制多份

Sharding
→ Split
→ 不同 Document 拆到不同 Shard
```

一个 Sharded Collection：

```text
users Collection

       Shard Key
          ↓

Shard 1        Shard 2
部分 users     部分 users
```

核心组件：

| 组件 | 作用 |
|---|---|
| Shard | 保存部分业务数据 |
| mongos | 请求路由 |
| Config Server | 保存分片元数据 |
| Shard Key | 决定数据如何分布 |

基本架构：

```text
Application
     ↓
   mongos
     ↓
 ┌───┴────┐
 ↓        ↓
Shard1   Shard2

Config Server
→ 保存分片元数据
```

---

# 十二、Shard Key

Shard Key 决定：

```text
Document 存到哪个 Shard
+
查询能否精准路由
```

例如：

```text
Shard Key = userId
```

同一个：

```text
users Collection
```

里面不同 Document 会根据 `userId` 分布到不同 Shard。

主要认识了：

```text
Range Sharding
Hashed Sharding
```

Hashed：

```text
userId
 ↓
Hash
 ↓
目标 Shard
```

创建：

```javascript
db.users.createIndex({
    userId: "hashed"
})
```

```javascript
sh.shardCollection(
    "middleware_lab.users",
    {
        userId: "hashed"
    }
)
```

查看：

```javascript
sh.status()
```

或者：

```javascript
db.users.getShardDistribution()
```

---

# 十三、mongos 与 Go

Sharding 后，业务不直接连接某个 Shard：

```text
Go
 ↓
mongos
 ↓
Shard Key
 ↓
目标 Shard
```

Go 连接：

```go
mongo.Connect(
    options.Client().
        ApplyURI("mongodb://mongos:27017"),
)
```

CRUD 基本不变：

```text
InsertOne
Find
UpdateOne
DeleteOne
Aggregate
```

业务代码不需要自己判断：

```text
这条数据属于 Shard1 还是 Shard2
```

这是：

```text
mongos
```

的职责。

---

# 十四、整体架构认知

把 Replica Set 和 Sharding 合起来：

```text
                       Application
                            ↓
                          mongos
                            ↓

           ┌────────────────┴────────────────┐

        Shard 1                           Shard 2
           │                                 │

      Replica Set                       Replica Set

         Primary                           Primary
        /      \                           /      \
 Secondary  Secondary               Secondary  Secondary
```

可以理解成两层：

```text
Sharding
→ 横向拆数据
→ 扩容量、扩吞吐

Replica Set
→ 每个 Shard 内复制数据
→ 高可用、容灾
```

---

# 十五、常用 mongosh 命令汇总

## Database / Collection

```javascript
db

use middleware_lab

show dbs

show collections
```

## CRUD

```javascript
db.products.insertOne(...)

db.products.insertMany(...)

db.products.find(...)

db.products.updateOne(...)

db.products.updateMany(...)

db.products.deleteOne(...)

db.products.deleteMany(...)
```

## Index

```javascript
db.products.createIndex(...)

db.products.getIndexes()

db.products.dropIndex(...)

db.products.find(...).explain("executionStats")
```

## Aggregation

```javascript
db.products.aggregate([
    ...
])
```

## Replica Set

```javascript
rs.initiate(...)

rs.status()

db.hello()
```

## Transaction

```javascript
startSession()

startTransaction()

commitTransaction()

abortTransaction()
```

## Sharding

```javascript
sh.addShard(...)

sh.shardCollection(...)

sh.status()

db.collection.getShardDistribution()
```

---

# 十六、Go Driver 核心 API

连接：

```go
mongo.Connect(...)
```

Database：

```go
client.Database(...)
```

Collection：

```go
db.Collection(...)
```

CRUD：

```go
InsertOne
InsertMany

Find
FindOne

UpdateOne
UpdateMany

DeleteOne
DeleteMany
```

查询：

```go
SetSort
SetSkip
SetLimit
SetProjection
```

索引：

```go
collection.Indexes().CreateOne(...)
```

聚合：

```go
collection.Aggregate(...)
```

事务：

```go
StartSession
WithTransaction
EndSession
```

---

# 十七、MongoDB 与 MySQL / Redis 的选型

## 更倾向 MongoDB

```text
数据天然像 Document

Schema 变化比较频繁

嵌套数据较多

不同类型数据字段差异较大

经常整体读取一个对象
```

例如：

```text
商品属性
内容系统
配置数据
用户画像
复杂 JSON 数据
```

## 更倾向 MySQL

```text
结构稳定

强事务需求明显

关系和 JOIN 很多

数据约束非常重要
```

例如：

```text
订单核心数据
支付
账户
财务
```

## 更倾向 Redis

```text
Key-Value

极低延迟

缓存

计数器

Session

临时状态
```

---

# 十八、整个学习过程最终建立的认知

这套课程可以压缩成一条完整主线：

```text
Document
   ↓
CRUD
   ↓
Filter / Cursor
   ↓
Index
   ↓
Aggregation
   ↓
Document Modeling
   ↓
Transaction
   ↓
Replica Set
   ↓
Sharding
```

分别解决：

```text
Document
→ 怎么存数据

CRUD
→ 怎么操作数据

Filter / Cursor
→ 怎么查询数据

Index
→ 怎么让查询更快

Aggregation
→ 怎么做统计和数据加工

Modeling
→ 数据应该怎么组织

Transaction
→ 多 Document 怎么保证一致性

Replica Set
→ 怎么做副本和高可用

Sharding
→ 怎么水平扩容
```

---

# 最终需要带走的几个关键词

```text
Document
Flexible Schema
BSON
Filter
Cursor
Index
Aggregation Pipeline
Embedded
Reference
Transaction
Session
Replica Set
Primary / Secondary
oplog
Election
Sharding
mongos
Config Server
Shard Key
```

如果这些关键词你都能用自己的话解释清楚，并且知道对应的基本命令和 Go API，那么对于“工作里没真正用过 MongoDB，但需要补齐中间件能力并应对后端面试”这个目标，已经具备了一套比较完整的基础认知和实践经验。