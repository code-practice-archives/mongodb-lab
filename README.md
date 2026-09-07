非常合适。对你现在的目标来说，MongoDB 最好的学习方式不是“系统学一遍数据库理论”，而是：

**MySQL 已知经验 → MongoDB 对照理解 → 命令实际操作 → Go 代码操作 → 最后补一点面试需要的原理。**

MongoDB 官方 Go Driver 本身也基本按照连接、CRUD、Aggregation、Index、Transaction 等路径组织，所以我们可以顺着这条线学习。

### MongoDB 入门学习清单

| 阶段 | 学什么 | 实践 |
|---|---|---|
| ① 跑起来 | MongoDB 是什么、Database / Collection / Document | Docker 启动 MongoDB，进入 `mongosh` |
| ② 数据模型 | BSON、Document，与 MySQL 表/行的区别 | 插入几条 JSON 风格数据 |
| ③ CRUD | insert / find / update / delete | 用 `mongosh` 完成增删改查 |
| ④ 查询能力 | 条件查询、排序、分页、数组、嵌套字段 | 写十几个简单查询 |
| ⑤ Go 接入 | MongoDB 官方 Go Driver | Go 完成连接 + CRUD |
| ⑥ 索引 | 单字段、联合、唯一索引，为什么需要索引 | 创建索引 + `explain()` 对比 |
| ⑦ 聚合 | `$match`、`$group`、`$sort`、`$project` | 类比 MySQL `GROUP BY` 写统计 |
| ⑧ 文档设计 | 嵌入 Embedded vs 引用 Reference | 设计一个商品/订单 Document |
| ⑨ 事务 | 单 Document 原子性、多 Document Transaction | Go 写一个简单事务 |
| ⑩ 部署原理 | Replica Set、Primary/Secondary、读写基本流程 | Docker 起一个简单 Replica Set |
| ⑪ 扩展概念 | Sharding 是什么、什么时候需要 | **只理解，不深入搭建** |
| ⑫ 选型 | MongoDB vs MySQL vs Redis | 能回答“什么时候该用 MongoDB” |

其中 CRUD、Aggregation、Indexes、Transactions 都是当前 MongoDB Go Driver 官方文档里的核心能力。

我建议我们贯穿整个学习过程做一个特别小的 **商品目录服务**：

```text
Product
├── id
├── name
├── category
├── price
├── attributes
│   ├── color
│   ├── size
│   └── ...
└── tags []
```

它非常适合学习 MongoDB，因为不同商品的 `attributes` 可以完全不同：

```json
{
  "name": "MacBook Pro",
  "category": "computer",
  "attributes": {
    "cpu": "M5",
    "memory": "32GB"
  }
}
```

另一件商品可以直接是：

```json
{
  "name": "LEGO Technic",
  "category": "toy",
  "attributes": {
    "pieces": 1500,
    "age": "10+"
  }
}
```

这样你会非常直观地体会：

> **为什么有了 MySQL，还会有人选择 MongoDB。**

而不是背一句“MongoDB 是文档型数据库”。

我们暂时**不深入** WiredTiger 存储引擎实现、B-Tree 细节、Replica Set 选举算法、Sharding Balancer 实现、复杂一致性理论。这些等以后遇到面试问题再补。

另外实践环境也很简单。官方目前提供 Community Docker 镜像，可以直接：

```bash
docker pull mongodb/mongodb-community-server:latest

docker run \
  --name mongodb \
  -p 27017:27017 \
  -d mongodb/mongodb-community-server:latest
```

然后本机就是：

```text
mongodb://localhost:27017
```

官方也推荐用这种方式快速建立本地 MongoDB 环境。

截至 2026 年 9 月，MongoDB 当前稳定系列已经到 **8.3**，官方 Go Driver 当前文档对应 v2 系列；不过我们整个入门过程尽量写**不依赖特定小版本**的代码。

我建议现在直接从 **第 ① 课：MongoDB 到底是什么 + Docker 跑起来 + 插入第一条数据** 开始，我带你边操作边和 MySQL 对比。