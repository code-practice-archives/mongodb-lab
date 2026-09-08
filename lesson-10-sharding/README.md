## Lesson 10：MongoDB Sharding

### 第一步：快速理解概念

Sharding 解决的是：

> **单个 MongoDB 节点的存储容量或吞吐不够时，把同一个 Collection 的不同 Document 拆到不同 Shard。**

例如：

```text
users Collection

            Sharding
               ↓

      Shard 1        Shard 2
     部分 users      部分 users
```

核心只记 4 个东西：

| 组件 | 作用 |
|---|---|
| `Shard` | 保存部分数据 |
| `mongos` | 请求路由 |
| Config Server | 保存分片元数据 |
| Shard Key | 决定 Document 如何分布 |

和上一课的区别：

```text
Replica Set = Copy
同一份数据复制多份

Sharding = Split
不同 Document 拆到不同节点
```

MongoDB 的分片单位是 Collection 内的数据；一个数据库可以同时包含分片和非分片 Collection。

---

# 第二步：真正搭一个 Sharding

我们只搭最小环境：

```text
               mongos
                  │
          ┌───────┴───────┐
          ↓               ↓
       shard1           shard2
    单节点 Replica    单节点 Replica

               ↑
         Config Server
         单节点 Replica
```

真实生产当然不会这么部署，我们只是为了把整个流程跑通。

### 1. 创建目录

```bash
mkdir lesson-10-sharding
cd lesson-10-sharding
```

创建：

```text
lesson-10-sharding/
└── compose.yaml
```

### 2. `compose.yaml`

```yaml
services:

  config:
    image: mongo:8.0
    container_name: mongodb-lesson-10-config
    hostname: config

    ports:
      - "27031:27017"

    command:
      - mongod
      - --configsvr
      - --replSet
      - configRS
      - --port
      - "27017"
      - --bind_ip_all

    volumes:
      - config_data:/data/db


  shard1:
    image: mongo:8.0
    container_name: mongodb-lesson-10-shard1
    hostname: shard1

    ports:
      - "27032:27017"

    command:
      - mongod
      - --shardsvr
      - --replSet
      - shard1RS
      - --port
      - "27017"
      - --bind_ip_all

    volumes:
      - shard1_data:/data/db


  shard2:
    image: mongo:8.0
    container_name: mongodb-lesson-10-shard2
    hostname: shard2

    ports:
      - "27033:27017"

    command:
      - mongod
      - --shardsvr
      - --replSet
      - shard2RS
      - --port
      - "27017"
      - --bind_ip_all

    volumes:
      - shard2_data:/data/db


  mongos:
    image: mongo:8.0
    container_name: mongodb-lesson-10-mongos

    ports:
      - "27030:27017"

    command:
      - mongos
      - --configdb
      - configRS/config:27017
      - --port
      - "27017"
      - --bind_ip_all

    depends_on:
      - config
      - shard1
      - shard2


volumes:
  config_data:
  shard1_data:
  shard2_data:
```

Config Server、Shard 都以 Replica Set 形式运行，而 `mongos` 使用 Config Server 的地址启动，这是 MongoDB 标准的 Sharded Cluster 结构。

---

### 3. 先启动三个 `mongod`

不要先启动 `mongos`：

```bash
docker compose up -d config shard1 shard2
```

检查：

```bash
docker compose ps
```

---

### 4. 初始化 Config Server

连接：

```bash
mongosh "mongodb://localhost:27031/?directConnection=true"
```

执行：

```javascript
rs.initiate({
    _id: "configRS",
    configsvr: true,
    members: [
        {
            _id: 0,
            host: "config:27017"
        }
    ]
})
```

确认：

```javascript
rs.status().members.map(m => ({
    name: m.name,
    state: m.stateStr
}))
```

应该成为：

```text
config:27017 → PRIMARY
```

---

### 5. 初始化 Shard 1

```bash
mongosh "mongodb://localhost:27032/?directConnection=true"
```

执行：

```javascript
rs.initiate({
    _id: "shard1RS",
    members: [
        {
            _id: 0,
            host: "shard1:27017"
        }
    ]
})
```

等待：

```javascript
rs.status().members[0].stateStr
```

返回：

```text
PRIMARY
```

---

### 6. 初始化 Shard 2

```bash
mongosh "mongodb://localhost:27033/?directConnection=true"
```

执行：

```javascript
rs.initiate({
    _id: "shard2RS",
    members: [
        {
            _id: 0,
            host: "shard2:27017"
        }
    ]
})
```

确认：

```javascript
rs.status().members[0].stateStr
```

返回：

```text
PRIMARY
```

到这里：

```text
configRS
└── config

shard1RS
└── shard1

shard2RS
└── shard2
```

都准备好了。

---

### 7. 现在启动 mongos

```bash
docker compose up -d mongos
```

连接：

```bash
mongosh "mongodb://localhost:27030"
```

注意：

> 从现在开始，业务操作主要连接 `mongos`，而不是直接连接 shard。

---

### 8. 把两个 Shard 加入集群

```javascript
sh.addShard(
    "shard1RS/shard1:27017"
)
```

```javascript
sh.addShard(
    "shard2RS/shard2:27017"
)
```

查看：

```javascript
sh.status()
```

应该能看到：

```text
shard1RS
shard2RS
```

`sh.addShard()` 就是把 Replica Set 注册成 Sharded Cluster 的一个 Shard。

---

### 9. 创建真正的 Sharded Collection

我们使用：

```text
middleware_lab.users
```

并选择：

```text
userId
```

作为 Shard Key。

```javascript
use middleware_lab
```

先建立 hashed index：

```javascript
db.users.createIndex({
    userId: "hashed"
})
```

然后：

```javascript
sh.shardAndDistributeCollection(
    "middleware_lab.users",
    {
        userId: "hashed"
    },
    false,
    {
        numInitialChunks: 4
    }
)
```

这里表达的就是：

```text
users Collection
      ↓
userId 做 Hash
      ↓
划分不同数据范围
      ↓
Shard1 / Shard2
```

MongoDB 8.0 的 `sh.shardAndDistributeCollection()` 会对 Collection 分片并立即进行数据分布；Shard Key 必须有对应索引。

---

### 10. 插入数据

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

注意我们仍然只是：

```javascript
db.users.insertMany(...)
```

没有：

```text
insert shard1
insert shard2
```

因为过程是：

```text
Client
   ↓
mongos
   ↓
userId Shard Key
   ↓
决定目标 Shard
```

---

### 11. 看数据有没有真正分片

先：

```javascript
sh.status()
```

也可以：

```javascript
db.users.getShardDistribution()
```

你要观察的不是每个数字，而是：

```text
Shard1
→ 有一部分 users

Shard2
→ 有另一部分 users
```

到这里，这一课最重要的实验其实就完成了：

```text
同一个 users Collection

        ↓ Shard Key

Shard1             Shard2
一部分 Document    另一部分 Document
```

---

# 第三步：Go 有什么变化？

这一步反而非常简单。

以前连接普通 MongoDB：

```go
mongodb://localhost:27017
```

Replica Set：

```go
mongodb://mongo1,mongo2,mongo3/?replicaSet=rs0
```

Sharding：

```go
mongodb://mongos:27017
```

例如：

```go
client, err := mongo.Connect(
    options.Client().
        ApplyURI("mongodb://mongos:27017"),
)
```

后面的：

```go
collection.InsertOne(...)
collection.Find(...)
collection.UpdateOne(...)
collection.DeleteOne(...)
```

**基本不变。**

应用不应该写：

```go
if userID < 10000 {
    // shard1
} else {
    // shard2
}
```

而是：

```text
Go
 ↓
mongos
 ↓
根据 Shard Key 路由
 ↓
正确的 Shard
```

客户端应通过 `mongos` 访问 Sharded Cluster，而不是自行连接某个 Shard 做正常业务读写。

---

这样这一课就只保留三个大步骤：

```text
① 概念
Sharding = 把同一 Collection 的 Document 横向拆分

② MongoDB 实践
Compose
→ 初始化 3 个 Replica Set
→ mongos
→ addShard
→ Shard Key
→ 插入并验证

③ Go
主要变化就是：
连接 mongos
```
