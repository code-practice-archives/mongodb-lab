# Lesson 09：MongoDB Replica Set

上一课为了使用 Transaction，我们把 MongoDB 临时启动成了单节点 Replica Set。

这一课反过来专门回答：

> **Replica Set 到底解决什么问题？Primary / Secondary 怎么同步？Primary 挂了以后为什么还能继续服务？**

这次我们真的搭 **3 个 MongoDB 节点**，然后亲手做一次数据复制和主节点故障切换。

---

## 1. 从 Standalone 的问题开始

前面大多数课程只有：

```text
Client
  ↓
MongoDB
```

这种 Standalone 有一个很明显的问题：

```text
MongoDB 挂了
    ↓
整个数据库不可用
```

而且只有一份数据。

于是最直接的想法就是：

```text
               MongoDB 1
                  │
               数据复制
              /       \
       MongoDB 2     MongoDB 3
```

这就是 Replica Set 要解决的核心问题：

> **维护多份相同的数据，并在当前主节点故障后自动选出新的主节点。**

MongoDB 官方推荐的典型三成员结构就是一个 Primary + 两个 Secondary；它能够提供数据冗余和故障切换能力。

---

# 2. Primary 和 Secondary 是什么

一个正常的三节点 Replica Set：

```text
             Primary
           /         \
    Secondary       Secondary
```

### Primary

负责接收写请求：

```text
Client
  ↓
INSERT / UPDATE / DELETE
  ↓
Primary
```

Replica Set 同一时刻最多只有一个 Primary。

### Secondary

Secondary 保存 Primary 数据的副本：

```text
Primary
   ↓
Secondary
Secondary
```

它们默认不负责写入，但可以根据 Read Preference 承担读取。Secondary 也可以在 Primary 故障后参与选举并成为新的 Primary。

---

# 3. 数据到底怎么复制？

这里需要认识我们上一课提到过的：

```text
oplog
```

假设 Primary 执行：

```javascript
db.products.insertOne({
    name: "LEGO Technic"
})
```

内部可以先粗略理解成：

```text
Primary
  │
  ├─ 修改自己的数据
  │
  └─ 写入 oplog
          ↓
      Secondary
          ↓
       读取 oplog
          ↓
       应用操作
```

MongoDB 的 oplog 位于：

```text
local.oplog.rs
```

它是一个记录数据变更操作的特殊 collection。Secondary 会异步复制并应用这些 oplog 操作，从而保持自己的数据副本。

注意这里有一个重要词：

> **异步复制。**

因此 Secondary 理论上可能比 Primary 慢一点，也就是存在：

```text
Replication Lag
```

所以以后如果让业务读 Secondary，就要意识到：

> **有可能读到稍旧的数据。** 

---

# 4. 开始搭三个节点

创建：

```text
lesson-09-replica-set/
└── compose.yaml
```

`compose.yaml`：

```yaml
services:
  mongo1:
    image: mongodb/mongodb-community-server:8.0.29-ubi9-slim
    container_name: mongodb-lesson-09-1
    hostname: mongo1

    ports:
      - "27025:27017"

    command:
      - --replSet
      - rs0

    volumes:
      - mongo1_data:/data/db

  mongo2:
    image: mongodb/mongodb-community-server:8.0.29-ubi9-slim
    container_name: mongodb-lesson-09-2
    hostname: mongo2

    ports:
      - "27026:27017"

    command:
      - --replSet
      - rs0

    volumes:
      - mongo2_data:/data/db

  mongo3:
    image: mongodb/mongodb-community-server:8.0.29-ubi9-slim
    container_name: mongodb-lesson-09-3
    hostname: mongo3

    ports:
      - "27027:27017"

    command:
      - --replSet
      - rs0

    volumes:
      - mongo3_data:/data/db

volumes:
  mongo1_data:
  mongo2_data:
  mongo3_data:
```

这里三个节点都有：

```text
--replSet rs0
```

说明：

```text
mongo1
mongo2
mongo3

都属于同一个 rs0
```

启动：

```bash
docker compose up -d
```

检查：

```bash
docker compose ps
```

---

# 5. 初始化 Replica Set

连接第一个节点：

```bash
mongosh "mongodb://localhost:27025/?directConnection=true"
```

初始化：

```javascript
rs.initiate({
    _id: "rs0",

    members: [
        {
            _id: 0,
            host: "mongo1:27017"
        },
        {
            _id: 1,
            host: "mongo2:27017"
        },
        {
            _id: 2,
            host: "mongo3:27017"
        }
    ]
})
```

这里：

```text
rs0
├── mongo1
├── mongo2
└── mongo3
```

三个容器在同一个 Compose 网络里，所以可以通过：

```text
mongo1
mongo2
mongo3
```

相互访问。

查看状态：

```javascript
rs.status()
```

为了看得简单一点，可以：

```javascript
rs.status().members.map(m => ({
    name: m.name,
    state: m.stateStr
}))
```

正常会类似：

```javascript
[
    {
        name: "mongo1:27017",
        state: "PRIMARY"
    },
    {
        name: "mongo2:27017",
        state: "SECONDARY"
    },
    {
        name: "mongo3:27017",
        state: "SECONDARY"
    }
]
```

于是：

```text
              mongo1
              PRIMARY
             /       \
       mongo2         mongo3
      SECONDARY      SECONDARY
```

成员的核心状态就是 `PRIMARY` 和 `SECONDARY`；Secondary 在需要时可以参与选举成为 Primary。

---

# 6. 实验一：Primary 写入

连接 Primary：

```bash
mongosh "mongodb://localhost:27025/?directConnection=true"
```

创建数据：

```javascript
use middleware_lab
```

```javascript
db.products.insertOne({
    name: "LEGO Technic",
    price: 599
})
```

查询：

```javascript
db.products.find()
```

Primary 上当然能看到：

```javascript
{
    name: "LEGO Technic",
    price: 599
}
```

---

# 7. 实验二：看看 Secondary 有没有数据

现在直接连接 mongo2：

```bash
mongosh "mongodb://localhost:27026/?directConnection=true&readPreference=secondary"
```

执行：

```javascript
use middleware_lab
```

然后：

```javascript
db.products.find()
```

应该同样能够看到：

```javascript
{
    name: "LEGO Technic",
    price: 599
}
```

也就是说刚才：

```text
mongo1 Primary
      ↓
insert LEGO
      ↓
写 oplog
      ↓
mongo2 / mongo3
复制并应用
      ↓
也拥有 LEGO
```

这里为了直接从 Secondary 查询，我们显式设置：

```text
readPreference=secondary
```

`mongosh` 支持通过 Read Preference 从 Secondary 读取。

---

# 8. 看看 oplog

重新连接 Primary：

```bash
mongosh "mongodb://localhost:27025/?directConnection=true"
```

执行：

```javascript
db.getSiblingDB("local")
  .oplog
  .rs
  .find()
  .sort({ $natural: -1 })
  .limit(5)
```

你会看到最近的一些复制操作记录。

现在不用分析所有字段。

只需要真正见过：

```text
local.oplog.rs
```

并建立：

```text
业务写入
   ↓
Primary
   ↓
oplog
   ↓
Secondary
```

这个模型即可。

---

# 9. 实验三：Primary 挂掉怎么办？

现在才进入 Replica Set 最有意思的地方。

当前：

```text
mongo1
PRIMARY

mongo2
SECONDARY

mongo3
SECONDARY
```

直接停掉 mongo1：

```bash
docker compose stop mongo1
```

此时只剩：

```text
mongo2
mongo3
```

因为三节点 Replica Set 仍然有：

```text
2 / 3
```

节点可用，也就是还有多数派，所以可以重新选举 Primary。MongoDB 的选举要求 Replica Set 能够形成多数派；如果无法形成多数派，就无法选出 Primary，也就不能继续正常写入。

---

# 10. 查看谁变成了新的 Primary

先连接 mongo2：

```bash
mongosh "mongodb://localhost:27026/?directConnection=true"
```

执行：

```javascript
rs.status().members.map(m => ({
    name: m.name,
    state: m.stateStr
}))
```

你可能看到：

```text
mongo1 → DOWN

mongo2 → PRIMARY

mongo3 → SECONDARY
```

也可能：

```text
mongo1 → DOWN

mongo2 → SECONDARY

mongo3 → PRIMARY
```

具体谁当选不用提前假设。

也可以执行：

```javascript
db.hello().isWritablePrimary
```

如果返回：

```text
true
```

说明当前这个节点就是 Primary。

MongoDB 在 Primary 不可用后会通过选举让一个 Secondary 成为新的 Primary。

---

# 11. 在新的 Primary 上继续写

假设 mongo2 当选：

```bash
mongosh "mongodb://localhost:27026/?directConnection=true"
```

执行：

```javascript
use middleware_lab
```

```javascript
db.products.insertOne({
    name: "LEGO City",
    price: 299
})
```

成功。

也就是说：

```text
原 Primary 挂掉
       ↓
Secondary 选举
       ↓
产生新 Primary
       ↓
业务继续写入
```

这就是 Replica Set 的核心高可用能力。

---

# 12. 原来的 Primary 恢复会怎样？

重新启动：

```bash
docker compose start mongo1
```

mongo1 会重新加入 Replica Set。

它发现自己期间少了一些操作：

```text
LEGO City
```

于是会通过 Replica Set 的同步机制追赶最新数据。MongoDB 对新加入或落后的成员会通过初始同步或持续 oplog replication 来保持数据一致。

之后查看：

```javascript
rs.status()
```

它会重新进入正常成员状态。

---

# 13. 为什么通常是三个节点？

这是 Replica Set 很值得理解的一点。

假设只有两个节点：

```text
mongo1
mongo2
```

mongo1 挂了：

```text
只剩 mongo2
```

它到底能不能确认：

```text
mongo1 真挂了？
```

还是：

```text
只是我和 mongo1 之间网络断了？
```

所以分布式系统通常需要**多数派**。

三节点：

```text
mongo1
mongo2
mongo3
```

只要：

```text
2 / 3
```

还能互相通信，就有多数派。

因此一个节点挂掉：

```text
2 / 3
→ 可以继续选 Primary
```

如果两个节点挂掉：

```text
1 / 3
→ 没有多数派
→ 无法维持 Primary
→ 无法继续正常写入
```

MongoDB 官方也将三个数据节点的 `Primary + Secondary + Secondary` 作为推荐的基础 Replica Set 结构。

---

# 14. 读请求一定走 Primary 吗？

默认：

```text
Write
 ↓
Primary

Read
 ↓
Primary
```

MongoDB Driver 默认读取 Primary。

但也可以配置：

```text
Read Preference
```

例如：

```text
primary
secondary
primaryPreferred
secondaryPreferred
nearest
```

比如：

```text
secondary
```

意味着：

```text
读请求
  ↓
Secondary
```

这样可以把部分读压力分散到 Secondary。

但代价是：

```text
Primary 刚写完
     ↓
Secondary 还没同步完
     ↓
立即去 Secondary 查询
     ↓
可能看到旧数据
```

也就是：

> **读扩展能力和数据新鲜度之间存在取舍。** 

这部分我们现在知道即可，不深入各种 Read Preference 配置。

---

# 15. 工程代码里的连接是什么样

如果你的 Go 服务和 MongoDB 都运行在同一个 Docker Compose 网络中，连接 URI 通常可以写成：

```text
mongodb://mongo1:27017,mongo2:27017,mongo3:27017/?replicaSet=rs0
```

Go：

```go
client, err := mongo.Connect(
    options.Client().ApplyURI(
        "mongodb://mongo1:27017,mongo2:27017,mongo3:27017/?replicaSet=rs0",
    ),
)
```

注意这里不是：

```text
只连接 Primary
```

而是告诉 Driver：

```text
这是 rs0

这些是可发现的成员
```

Driver 会根据 Replica Set 状态寻找当前 Primary。

所以 Primary 切换之后，正常的 MongoDB Driver 会重新发现新的 Primary，而不是要求业务代码手动改 IP。

这就是生产环境中 Replica Set 对应用真正有价值的地方：

```text
应用
 ↓
MongoDB Driver
 ↓
Replica Set
 ↓
自动发现当前 Primary
```

---

# 16. 这一课真正要建立的模型

先从问题开始：

```text
Standalone
   ↓
单点故障
   ↓
Replica Set
```

Replica Set：

```text
           Primary
          /       \
   Secondary     Secondary
```

数据同步：

```text
Write
 ↓
Primary
 ↓
oplog
 ↓
Secondary
```

故障切换：

```text
Primary DOWN
     ↓
Secondary Election
     ↓
New Primary
```

最终记住四个关键词就够了：

| 概念 | 作用 |
|---|---|
| Primary | 接收写入 |
| Secondary | 保存数据副本 |
| oplog | 记录并复制数据变化 |
| Election | Primary 故障后选出新的 Primary |

这就是 MongoDB Replica Set 的核心。

这一课我们暂时**不深入**选举算法细节、Write Concern、Priority、Hidden Member、Delayed Member、Arbiter 等配置；这些知道存在即可。后面如果继续按照原清单，下一课就会进入 **Sharding（分片）**：Replica Set 解决的是“一个数据集怎么做副本和高可用”，Sharding 则解决“单个数据集太大，一个 Replica Set 都装不下或扛不住怎么办”。