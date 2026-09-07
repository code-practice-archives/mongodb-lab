# MongoDB Lesson 06：Aggregation 手册

## 核心认识

MongoDB 聚合使用 **Aggregation Pipeline**：

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

前一个 Stage 的输出，会作为下一个 Stage 的输入。

---

## 常用 Stage

| Stage | 作用 | 类似 MySQL |
|---|---|---|
| `$match` | 过滤数据 | `WHERE` |
| `$group` | 分组统计 | `GROUP BY` |
| `$sort` | 排序 | `ORDER BY` |
| `$project` | 整理返回字段 | `SELECT` |
| `$limit` | 限制返回数量 | `LIMIT` |

常用聚合操作：

| 操作符 | 作用 |
|---|---|
| `$sum` | 求和 / 计数 |
| `$avg` | 平均值 |

---

## 基础聚合

按分类统计数量：

```javascript
db.products.aggregate([
    {
        $group: {
            _id: "$category",
            count: {
                $sum: 1
            }
        }
    }
])
```

其中：

```text
_id: "$category"
```

表示按 `category` 分组。

---

## 常见 Pipeline

```javascript
db.products.aggregate([
    {
        $match: {
            stock: {
                $gt: 10
            }
        }
    },
    {
        $group: {
            _id: "$category",
            count: {
                $sum: 1
            },
            avgPrice: {
                $avg: "$price"
            }
        }
    },
    {
        $sort: {
            avgPrice: -1
        }
    },
    {
        $project: {
            _id: 0,
            category: "$_id",
            count: 1,
            avgPrice: 1
        }
    }
])
```

可以理解成：

```text
筛选
→ 分组统计
→ 排序
→ 整理结果
```

---

## Go 中执行聚合

Pipeline：

```go
pipeline := mongo.Pipeline{
    {
        {"$match", bson.D{
            {"stock", bson.D{
                {"$gt", 10},
            }},
        }},
    },
    {
        {"$group", bson.D{
            {"_id", "$category"},
            {"count", bson.D{
                {"$sum", 1},
            }},
            {"avgPrice", bson.D{
                {"$avg", "$price"},
            }},
        }},
    },
}
```

执行：

```go
cursor, err := collection.Aggregate(
    ctx,
    pipeline,
)
```

读取结果：

```go
var results []CategoryStats

err = cursor.All(
    ctx,
    &results,
)
```

---

## 本课重点

```text
Aggregation
   ↓
由多个 Stage 组成 Pipeline
   ↓
每个 Stage 负责一步数据处理
```

现阶段重点掌握：

```text
$match
$group
$sort
$project
$sum
$avg
```

以及 Go 中：

```go
collection.Aggregate(ctx, pipeline)
```

Aggregation 最重要的是 **Pipeline 思维**，而不是记大量操作符。