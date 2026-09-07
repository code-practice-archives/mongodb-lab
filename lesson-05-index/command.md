可以，继续保持前几课一样的风格：**短、归纳、可速查**。

# MongoDB Lesson 05：索引手册

## 核心认识

索引的核心作用：

> **减少无效 Document 扫描，提高查询效率。**

没有合适索引时，常见：

```text
COLLSCAN
```

使用索引后，常见：

```text
IXSCAN
```

---

## 常用索引命令

| 命令 | 作用 |
|---|---|
| `createIndex()` | 创建索引 |
| `getIndexes()` | 查看 Collection 索引 |
| `dropIndex()` | 删除指定索引 |
| `explain("executionStats")` | 查看查询执行情况 |

### 单字段索引

```javascript
db.products.createIndex({
    sku: 1
})
```

```text
1  → 升序
-1 → 降序
```

### 联合索引

```javascript
db.products.createIndex({
    category: 1,
    price: 1
})
```

### 唯一索引

```javascript
db.products.createIndex(
    { sku: 1 },
    { unique: true }
)
```

---

## explain

```javascript
db.products.find({
    sku: "SKU-09000"
}).explain("executionStats")
```

现阶段重点关注：

| 指标 | 含义 |
|---|---|
| `COLLSCAN` | 扫描 Collection |
| `IXSCAN` | 使用索引扫描 |
| `totalDocsExamined` | 实际检查了多少 Document |
| `totalKeysExamined` | 扫描了多少索引项 |
| `nReturned` | 最终返回多少条数据 |

理想情况通常是：

```text
扫描数量
≈
实际返回数量
```

---

## Go 创建索引

```go
index := mongo.IndexModel{
    Keys: bson.D{
        {Key: "sku", Value: 1},
    },
}

name, err := collection.
    Indexes().
    CreateOne(ctx, index)
```

唯一索引：

```go
index := mongo.IndexModel{
    Keys: bson.D{
        {Key: "sku", Value: 1},
    },
    Options: options.Index().
        SetUnique(true),
}
```

---

## 索引的取舍

```text
索引
 ├─ 提高查询效率
 └─ 增加存储和写入维护成本
```

所以不是所有字段都需要建立索引。

通常优先考虑：

```text
经常作为查询条件的字段

经常参与排序的字段

需要唯一性约束的字段
```

---

## 本课重点

掌握三类索引：

```text
单字段索引
联合索引
唯一索引
```

掌握一个重要排查工具：

```javascript
.explain("executionStats")
```

慢查询时，可以先看：

```text
是否 COLLSCAN
        ↓
扫描了多少 Document
        ↓
是否存在合适索引
```