可以，第三课手册继续保持“短、归纳、可速查”。

# MongoDB Lesson 03：Go Driver 手册

## 核心认识

Go 操作 MongoDB，本质上就是把前两课的 MongoDB 命令换成 Go API：

| mongosh | Go Driver |
|---|---|
| `use db` | `client.Database()` |
| `db.products` | `db.Collection()` |
| `insertOne()` | `InsertOne()` |
| `find()` | `Find()` / `FindOne()` |
| `updateOne()` | `UpdateOne()` |
| `deleteOne()` | `DeleteOne()` |
| `{ name: "xxx" }` | `bson.M{"name": "xxx"}` |

---

## 连接 MongoDB

```go
client, err := mongo.Connect(
    options.Client().
        ApplyURI("mongodb://localhost:27019"),
)
```

验证连接：

```go
err = client.Ping(context.Background(), nil)
```

释放连接：

```go
defer client.Disconnect(context.Background())
```

---

## 获取 Collection

```go
collection := client.
    Database("middleware_lab").
    Collection("products")
```

---

## Go Struct 与 Document

```go
type Product struct {
    Name     string `bson:"name"`
    Category string `bson:"category"`
    Price    int    `bson:"price"`
    Stock    int    `bson:"stock"`
}
```

核心关系：

```text
Go Struct
   ↓ BSON 编解码
MongoDB Document
```

`bson:"name"` 类似常见的 `json:"name"`。

---

## BSON Filter

```go
bson.M{
    "name": "LEGO Technic",
}
```

对应 MongoDB：

```javascript
{
    name: "LEGO Technic"
}
```

Filter 依然是 MongoDB CRUD 的核心。

---

## CRUD

### Insert

```go
result, err := collection.InsertOne(ctx, Product{
    Name:     "LEGO Technic",
    Category: "toy",
    Price:    599,
    Stock:    20,
})
```

### FindOne

```go
var product Product

err := collection.FindOne(
    ctx,
    bson.M{"name": "LEGO Technic"},
).Decode(&product)
```

### UpdateOne

```go
result, err := collection.UpdateOne(
    ctx,
    bson.M{"name": "LEGO Technic"},
    bson.M{
        "$set": bson.M{
            "price": 699,
        },
    },
)
```

### DeleteOne

```go
result, err := collection.DeleteOne(
    ctx,
    bson.M{"name": "LEGO Technic"},
)
```

---

## 本课涉及的主要类型

| 类型 | 作用 |
|---|---|
| `mongo.Client` | MongoDB 客户端 |
| `mongo.Database` | Database |
| `mongo.Collection` | Collection |
| `bson.M` | 表达 Document、Filter、Update 等 BSON 数据 |
| `context.Context` | 控制数据库操作生命周期 |
| Struct + `bson` tag | BSON 与 Go Struct 的映射 |

---

## 本课重点

```text
MongoDB 命令
      ↓
Go Driver API
```

最重要的是理解：

> Go Driver 没有改变 MongoDB 的查询模型。

`Filter`、`$set`、`$inc` 等 MongoDB 语法仍然存在，只是通过 Go 的 `bson.M` 和 Struct 表达。

也就是：

```go
bson.M{
    "price": bson.M{
        "$gt": 5000,
    },
}
```

本质上就是：

```javascript
{
    price: {
        $gt: 5000
    }
}
```