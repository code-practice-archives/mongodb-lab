# MongoDB Lesson 07：数据建模手册

## 核心认识

MongoDB 关联数据主要有两种组织方式：

```text
Embedded
→ 数据直接嵌入同一个 Document

Reference
→ 数据拆到不同 Collection，通过 ID 关联
```

---

## Embedded

适合：

- 经常一起查询
- 生命周期接近
- 数据规模不会无限增长

示例：

```javascript
{
    name: "MacBook Pro",
    attributes: {
        cpu: "M5",
        memory: "32GB"
    },
    tags: ["computer", "apple"]
}
```

查询嵌套字段：

```javascript
db.products.find({
    "attributes.memory": "32GB"
})
```

修改嵌套字段：

```javascript
db.products.updateOne(
    { name: "MacBook Pro" },
    {
        $set: {
            "attributes.memory": "64GB"
        }
    }
)
```

数组追加：

```javascript
db.products.updateOne(
    { name: "MacBook Pro" },
    {
        $push: {
            tags: "developer"
        }
    }
)
```

---

## Reference

适合：

- 数据可以独立存在
- 被多个 Document 共享
- 子数据可能很多或持续增长

示例：

```javascript
// categories
{
    _id: "computer",
    name: "电脑"
}
```

```javascript
// products
{
    name: "MacBook Pro",
    categoryId: "computer"
}
```

查询通常是：

```text
查 Product
  ↓
拿 categoryId
  ↓
再查 Category
```

MongoDB 不会像关系型数据库外键一样自动加载关联对象。

---

## Go Struct 表达

Embedded：

```go
type Attributes struct {
    CPU    string `bson:"cpu"`
    Memory string `bson:"memory"`
}

type Product struct {
    Name       string     `bson:"name"`
    Attributes Attributes `bson:"attributes"`
    Tags       []string   `bson:"tags"`
}
```

Reference：

```go
type Product struct {
    Name       string `bson:"name"`
    CategoryID string `bson:"categoryId"`
}
```

---

## 选择思路

```text
是否经常一起读取？
        ↓
      是
        ↓
数据是否较小、生命周期一致？
        ↓
      是
        ↓
    Embedded
```

否则更倾向：

```text
Reference
```

本课重点：

> MongoDB 建模不是先考虑“拆几张表”，而是先考虑数据平时怎么一起读、一起改，以及是否会持续增长。