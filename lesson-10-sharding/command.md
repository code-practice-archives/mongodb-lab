# MongoDB Lesson 10：Sharding 手册

## 核心认识

Sharding 的作用：

> **把同一个 Collection 的不同 Document 分布到不同 Shard，实现水平扩容。**

```text
Replica Set
→ 同一份数据复制多份

Sharding
→ 不同数据拆到不同 Shard
```

核心组件：

| 组件 | 作用 |
|---|---|
| Shard | 保存部分业务数据 |
| mongos | 请求路由 |
| Config Server | 保存分片元数据 |
| Shard Key | 决定 Document 如何分布 |

---

## 最小实验架构

```text
          mongos
            │
      ┌─────┴─────┐
      ↓           ↓
   shard1       shard2

      ↑
 Config Server
```

Config Server、Shard 本身都以 Replica Set 模式运行。

---

## Compose 关键配置

统一让容器内部监听 `27017`：

```yaml
config:
  command:
    - mongod
    - --configsvr
    - --replSet
    - configRS
    - --port
    - "27017"
    - --bind_ip_all

shard1:
  command:
    - mongod
    - --shardsvr
    - --replSet
    - shard1RS
    - --port
    - "27017"
    - --bind_ip_all

shard2:
  command:
    - mongod
    - --shardsvr
    - --replSet
    - shard2RS
    - --port
    - "27017"
    - --bind_ip_all
```

宿主机端口：

```text
mongos  → 27030
config  → 27031
shard1  → 27032
shard2  → 27033
```

---

## 初始化三个 Replica Set

Config Server：

```javascript
rs.initiate({
    _id: "configRS",
    configsvr: true,
    members: [
        { _id: 0, host: "config:27017" }
    ]
})
```

Shard 1：

```javascript
rs.initiate({
    _id: "shard1RS",
    members: [
        { _id: 0, host: "shard1:27017" }
    ]
})
```

Shard 2：

```javascript
rs.initiate({
    _id: "shard2RS",
    members: [
        { _id: 0, host: "shard2:27017" }
    ]
})
```

查看状态：

```javascript
rs.status()
```

---

## 添加 Shard

连接 `mongos`：

```bash
mongosh "mongodb://localhost:27030"
```

添加：

```javascript
sh.addShard(
    "shard1RS/shard1:27017"
)

sh.addShard(
    "shard2RS/shard2:27017"
)
```

查看：

```javascript
sh.status()
```

---

## 创建 Sharded Collection

选择：

```text
middleware_lab.users
```

Shard Key：

```text
userId
```

使用 Hashed Sharding：

```javascript
use middleware_lab
```

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

> 基础学习使用 `sh.shardCollection()` 即可，不需要 `sh.shardAndDistributeCollection()`。

---

## 插入并验证

```javascript
let users = []

for (let i = 1; i <= 1000; i++) {
    users.push({
        userId: i,
        name: "user-" + i
    })
}

db.users.insertMany(users)
```

查看分布：

```javascript
sh.status()
```

或：

```javascript
db.users.getShardDistribution()
```

核心观察：

```text
users Collection
      ↓
userId hashed
      ↓
Shard1 保存部分 Document
Shard2 保存另一部分 Document
```

---

## Go 接入

Sharding 后业务主要连接：

```text
mongos
```

Go：

```go
client, err := mongo.Connect(
    options.Client().
        ApplyURI("mongodb://mongos:27017"),
)
```

之后：

```go
InsertOne()
Find()
UpdateOne()
DeleteOne()
Aggregate()
```

基本不变。

应用不负责判断数据在哪个 Shard：

```text
Go
 ↓
mongos
 ↓
Shard Key
 ↓
目标 Shard
```

---

## 本课重点

```text
Sharding
→ 水平拆分数据

Shard Key
→ 决定 Document 如何分布

mongos
→ 负责路由请求

Config Server
→ 保存分片元数据
```

最重要的一句话：

> **Sharding 是把同一个 Collection 的不同 Document 横向拆到多个 Shard；Replica Set 则是在每个 Shard 内对同一份数据做副本和高可用。**