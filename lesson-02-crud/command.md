可以，第二课手册继续保持和第一课一样的风格：**短、归纳、突出重点，不重复教学过程**。

# MongoDB Lesson 02：CRUD 手册

## 核心认识

MongoDB CRUD 可以统一理解为：

```text
新增
insertOne / insertMany

查询
find(filter)

修改
updateOne / updateMany

删除
deleteOne / deleteMany
```

其中最重要的概念是：

> **Filter：先筛选出 Document，再决定查询、修改还是删除。**

---

## 常用 CRUD 命令

| 命令 | 作用 |
|---|---|
| `insertOne({...})` | 插入一个 Document |
| `insertMany([...])` | 插入多个 Document |
| `find(filter)` | 查询满足条件的 Document |
| `updateOne(filter, update)` | 修改一个匹配的 Document |
| `updateMany(filter, update)` | 修改所有匹配的 Document |
| `deleteOne(filter)` | 删除一个匹配的 Document |
| `deleteMany(filter)` | 删除所有匹配的 Document |

---

## Filter 示例

查询指定分类：

```javascript
db.products.find({
    category: "toy"
})
```

多条件默认是 AND：

```javascript
db.products.find({
    category: "computer",
    price: { $lt: 12000 }
})
```

---

## 常用比较操作符

| 操作符 | 含义 |
|---|---|
| `$gt` | 大于 |
| `$gte` | 大于等于 |
| `$lt` | 小于 |
| `$lte` | 小于等于 |
| `$ne` | 不等于 |
| `$in` | 包含于指定集合 |

示例：

```javascript
db.products.find({
    price: {
        $gte: 500,
        $lte: 10000
    }
})
```

---

## 更新操作

修改字段：

```javascript
db.products.updateOne(
    { name: "LEGO Technic" },
    { $set: { price: 699 } }
)
```

数值递增：

```javascript
db.products.updateOne(
    { name: "LEGO Technic" },
    { $inc: { stock: 10 } }
)
```

常用更新操作符：

| 操作符 | 作用 |
|---|---|
| `$set` | 设置字段值 |
| `$inc` | 在原值基础上增加 |

---

## 删除操作

删除一个：

```javascript
db.products.deleteOne({
    name: "LEGO City"
})
```

删除多个：

```javascript
db.products.deleteMany({
    category: "computer"
})
```

---

## 本课重点

```text
Filter
   ↓
匹配 Document
   ↓
find / update / delete
```

以及：

```text
insertOne / insertMany
find
updateOne / updateMany
deleteOne / deleteMany
```

这就是 MongoDB 最基础的 CRUD 操作模型。

这一篇就可以直接作为第二课的课后速查页。