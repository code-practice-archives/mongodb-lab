# MongoDB Lesson 04：查询与 Cursor 手册

## 核心认识

```text
Find()
  ↓
发起查询
  ↓
返回 Cursor
  ↓
All() / Next()
  ↓
消费查询结果
```

`Find()` 已经向 MongoDB 发起查询，但通常不会一次性把所有结果都加载到 Go 内存中。

---

## Find 与 Cursor

| API | 作用 |
|---|---|
| `Find()` | 查询多条数据，返回 `Cursor` |
| `FindOne()` | 查询一条数据 |
| `cursor.All()` | 读取全部剩余结果并解码到 slice |
| `cursor.Next()` | 逐条读取 |
| `cursor.Decode()` | 将当前 Document 解码到 Struct |
| `cursor.Err()` | 检查遍历过程中是否出错 |
| `cursor.Close()` | 关闭 Cursor |

### 一次读完

```go
cursor, err := collection.Find(ctx, filter)

var products []Product
err = cursor.All(ctx, &products)
```

适合结果数量较小、明确可控的查询。

### 逐条读取

```go
for cursor.Next(ctx) {
    var product Product

    if err := cursor.Decode(&product); err != nil {
        return err
    }

    fmt.Println(product)
}
```

适合数据量较大或需要边读取边处理的场景。

---

## 查询选项

| API | 类似 SQL | 作用 |
|---|---|---|
| `SetSort()` | `ORDER BY` | 排序 |
| `SetSkip()` | `OFFSET` | 跳过指定数量 |
| `SetLimit()` | `LIMIT` | 限制返回数量 |
| `SetProjection()` | `SELECT field...` | 指定返回字段 |

### 排序

```go
options.Find().
    SetSort(
        bson.D{
            {Key: "price", Value: -1},
        },
    )
```

```text
1  → 升序
-1 → 降序
```

### 分页

```go
page := 2
pageSize := 10

opts := options.Find().
    SetSkip(int64((page - 1) * pageSize)).
    SetLimit(int64(pageSize))
```

### Projection

```go
options.Find().
    SetProjection(
        bson.M{
            "name":  1,
            "price": 1,
            "_id":   0,
        },
    )
```

---

## bson.M 与 bson.D

| 类型 | 特点 | 常见用途 |
|---|---|---|
| `bson.M` | 不强调字段顺序 | Filter、普通 Document |
| `bson.D` | 保留字段顺序 | Sort 等顺序有意义的场景 |

```go
// Filter
bson.M{
    "category": "toy",
}
```

```go
// Sort
bson.D{
    {Key: "price", Value: -1},
    {Key: "name", Value: 1},
}
```

---

## 常见查询组合

```go
filter := bson.M{
    "category": "toy",
}

opts := options.Find().
    SetSort(
        bson.D{
            {Key: "price", Value: -1},
        },
    ).
    SetSkip(0).
    SetLimit(10)

cursor, err := collection.Find(
    ctx,
    filter,
    opts,
)
```

整体模型：

```text
Filter
  +
FindOptions
  ↓
Find()
  ↓
Cursor
  ↓
All / Next
```

## 本课重点

- `Find()` 返回的是 `Cursor`，不是 `[]Product`。
- `All()` 会把结果全部读取到 slice。
- `Next()` 适合逐条消费结果。
- 排序、分页、字段裁剪都通过 `FindOptions` 完成。
- 普通 Filter 常用 `bson.M`，需要顺序时使用 `bson.D`。