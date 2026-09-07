好，第二课我们正式进入 **MongoDB CRUD**。这一课仍然只用 `mongosh`，暂时不写 Go。

目标是让你真正建立 MongoDB 最基础的操作手感：

> **准备数据 → 查询 → 条件查询 → 修改 → 删除**

MongoDB 官方 CRUD 主要就是 `find()`、`updateOne/updateMany()`、`deleteOne/deleteMany()` 这些方法；写操作在**单个 Document 级别具有原子性**。

---

# Lesson 02：MongoDB CRUD

## 1. 创建独立课程

目录变成：

```text
mongodb-lab/
├── lesson-01-basic/
│   └── compose.yaml
│
└── lesson-02-crud/
    └── compose.yaml
```

创建：

```bash
mkdir lesson-02-crud
cd lesson-02-crud
```

`compose.yaml`：

```yaml
services:
  mongodb:
    image: mongodb/mongodb-community-server:8.0.29-ubi9-slim
    container_name: mongodb-lesson-02
    ports:
      - "27018:27017"
    volumes:
      - mongodb_lesson_02_data:/data/db

volumes:
  mongodb_lesson_02_data:
```

注意我把宿主机端口改成：

```text
27018
```

这样即使第一课 MongoDB 还在运行，也不会冲突：

```text
Lesson 01 → localhost:27017

Lesson 02 → localhost:27018
```

启动：

```bash
docker compose up -d
```

连接：

```bash
mongosh "mongodb://localhost:27018"
```

---

# 2. 准备这一课的数据

切换数据库：

```javascript
use middleware_lab
```

这次我们一次插入几条数据：

```javascript
db.products.insertMany([
    {
        name: "LEGO Technic",
        category: "toy",
        price: 599,
        stock: 20
    },
    {
        name: "LEGO City",
        category: "toy",
        price: 299,
        stock: 50
    },
    {
        name: "MacBook Pro",
        category: "computer",
        price: 15999,
        stock: 10
    },
    {
        name: "ThinkPad X1",
        category: "computer",
        price: 9999,
        stock: 15
    },
    {
        name: "iPhone",
        category: "phone",
        price: 7999,
        stock: 30
    }
])
```

`insertMany()` 就是：

> 一次插入多个 Document。

如果没有手动指定 `_id`，MongoDB 会自动生成。

现在：

```javascript
db.products.find()
```

应该可以看到 5 条商品。

---

# 3. 查询：find()

第一课我们已经用过：

```javascript
db.products.find()
```

相当于：

```sql
SELECT * FROM products;
```

现在开始加条件。

### 查询玩具

```javascript
db.products.find({
    category: "toy"
})
```

类比：

```sql
SELECT *
FROM products
WHERE category = 'toy';
```

MongoDB 这里非常重要的一个东西叫：

> **Filter**

也就是：

```javascript
{
    category: "toy"
}
```

后面的查询、修改、删除都会反复使用 Filter。

---

# 4. 条件查询

### price 大于 5000

```javascript
db.products.find({
    price: {
        $gt: 5000
    }
})
```

其中：

```text
$gt = greater than
```

就是：

```sql
price > 5000
```

常见操作符先认识这几个：

| MongoDB | 含义 | SQL |
|---|---|---|
| `$gt` | 大于 | `>` |
| `$gte` | 大于等于 | `>=` |
| `$lt` | 小于 | `<` |
| `$lte` | 小于等于 | `<=` |
| `$ne` | 不等于 | `!=` |
| `$in` | 在某个集合中 | `IN` |

例如：

```javascript
db.products.find({
    price: {
        $gte: 500,
        $lte: 10000
    }
})
```

就是：

```sql
WHERE price >= 500
AND price <= 10000
```

---

# 5. 多条件查询

例如：

> computer，并且价格低于 12000。

```javascript
db.products.find({
    category: "computer",
    price: {
        $lt: 12000
    }
})
```

这种写法默认就是：

```text
AND
```

对应：

```sql
WHERE category = 'computer'
AND price < 12000;
```

查询本身可以针对普通字段、嵌套字段和数组建立 Filter。

---

# 6. 修改：updateOne()

假设 LEGO Technic 涨价：

```javascript
db.products.updateOne(
    {
        name: "LEGO Technic"
    },
    {
        $set: {
            price: 699
        }
    }
)
```

这里第一次看到 MongoDB 更新操作非常重要的结构：

```text
updateOne(
    filter,
    update
)
```

也就是：

```text
找到谁
   ↓
怎么修改
```

这里：

```javascript
{
    name: "LEGO Technic"
}
```

是 Filter。

而：

```javascript
{
    $set: {
        price: 699
    }
}
```

是修改内容。

可以理解成：

```sql
UPDATE products
SET price = 699
WHERE name = 'LEGO Technic';
```

MongoDB 当前提供 `updateOne()`、`updateMany()` 和 `replaceOne()` 等更新方式。

验证：

```javascript
db.products.find({
    name: "LEGO Technic"
})
```

---

# 7. `$inc`：非常常见

例如库存增加 10：

```javascript
db.products.updateOne(
    {
        name: "LEGO Technic"
    },
    {
        $inc: {
            stock: 10
        }
    }
)
```

原来：

```text
stock = 20
```

现在：

```text
stock = 30
```

这里的：

```text
$inc
```

就是：

> 在原值基础上增加。

这个以后做库存、计数器之类的场景会经常看到。

---

# 8. updateMany()

假设所有 `toy` 商品库存增加 5：

```javascript
db.products.updateMany(
    {
        category: "toy"
    },
    {
        $inc: {
            stock: 5
        }
    }
)
```

区别很直接：

```text
updateOne
    ↓
只修改一条匹配记录

updateMany
    ↓
修改所有匹配记录
```

---

# 9. 删除

删除一个：

```javascript
db.products.deleteOne({
    name: "LEGO City"
})
```

类似：

```sql
DELETE FROM products
WHERE name = 'LEGO City'
LIMIT 1;
```

MongoDB 的 `deleteOne()` 删除一个匹配 Document，并返回诸如 `deletedCount` 这样的结果信息。

删除所有 `computer`：

```javascript
db.products.deleteMany({
    category: "computer"
})
```

`deleteMany()` 会删除所有满足 Filter 的 Document。

---

# 10. 第二课最重要的认识

这一课其实不用记很多东西。

MongoDB CRUD 可以统一理解成：

```text
新增
insertOne()
insertMany()

查询
find(filter)

修改
updateOne(filter, update)
updateMany(filter, update)

删除
deleteOne(filter)
deleteMany(filter)
```

而其中最重要的概念其实是：

```text
Filter
```

比如：

```javascript
{
    category: "toy",
    price: {
        $lt: 1000
    }
}
```

它不只是 `find()` 使用。

而是：

```text
find(filter)

updateOne(filter, ...)

updateMany(filter, ...)

deleteOne(filter)

deleteMany(filter)
```

**MongoDB CRUD 很大程度上就是：先学会写 Filter，再决定对匹配出来的 Document 做什么。**

---

你现在可以先实际跑到这里。尤其建议你自己手敲一遍：

```javascript
db.products.find({
    category: "toy",
    price: {
        $lt: 1000
    }
})
```

然后：

```javascript
db.products.updateOne(
    { name: "LEGO Technic" },
    { $inc: { stock: 10 } }
)
```

这两个操作能理解清楚，MongoDB CRUD 的基本模型其实就已经建立起来了。下一步第二课还可以继续补 **排序、分页、指定返回字段，以及嵌套字段/数组查询**，这些才会让 `find()` 真正接近业务代码中的查询。