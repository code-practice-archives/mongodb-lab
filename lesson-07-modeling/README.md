# Lesson 07：MongoDB Document 数据建模

前面我们已经会 CRUD、查询、索引和聚合了。现在开始一个 MongoDB 很重要、也和 MySQL 思维差异比较大的问题：

> **有关联的数据，到底应该放在一个 Document 里，还是拆成多个 Collection？**

MongoDB 最常见就是两种方式：

```text
Embedded
嵌入在同一个 Document

Reference
拆成不同 Document，通过 ID 关联
```

MongoDB 官方也把这两种方式作为关联数据建模的核心选择。

---

## 1. 创建 Lesson 07

```text
mongodb-lab/
├── lesson-01-basic/
├── ...
├── lesson-06-aggregation/
└── lesson-07-modeling/
```

```bash
mkdir lesson-07-modeling
cd lesson-07-modeling
```

`compose.yaml`：

```yaml
services:
  mongodb:
    image: mongodb/mongodb-community-server:8.0.29-ubi9-slim
    container_name: mongodb-lesson-07
    ports:
      - "27023:27017"
    volumes:
      - mongodb_lesson_07_data:/data/db

volumes:
  mongodb_lesson_07_data:
```

启动：

```bash
docker compose up -d

mongosh "mongodb://localhost:27023"
```

然后：

```javascript
use middleware_lab
```

---

# 2. Embedded：嵌入式 Document

假设我们保存商品信息。

如果是 MySQL，可能设计：

```text
products
product_attributes
```

MongoDB 可以直接：

```javascript
db.products.insertOne({
    name: "MacBook Pro",
    price: 15999,

    attributes: {
        cpu: "M5",
        memory: "32GB",
        storage: "1TB"
    }
})
```

也就是：

```text
Product
├── name
├── price
└── attributes
      ├── cpu
      ├── memory
      └── storage
```

`attributes` 就是一个 **Embedded Document**。

MongoDB 的嵌入模型可以让相关数据一次查询取出来，并且相关字段可以在一个 Document 的原子写操作中一起修改。

---

## 3. 查询嵌套字段

查询：

> 内存是 32GB 的商品。

使用 **dot notation**：

```javascript
db.products.find({
    "attributes.memory": "32GB"
})
```

修改：

```javascript
db.products.updateOne(
    {
        name: "MacBook Pro"
    },
    {
        $set: {
            "attributes.memory": "64GB"
        }
    }
)
```

所以：

```text
attributes.memory
```

就是访问嵌套 Document 中字段的方式。

---

# 4. Array 也是非常常见的嵌入

例如：

```javascript
db.products.insertOne({
    name: "LEGO Technic",
    price: 599,

    tags: [
        "toy",
        "technic",
        "lego"
    ]
})
```

查询包含 `lego` 标签的商品：

```javascript
db.products.find({
    tags: "lego"
})
```

向数组增加元素：

```javascript
db.products.updateOne(
    {
        name: "LEGO Technic"
    },
    {
        $push: {
            tags: "car"
        }
    }
)
```

所以 MongoDB Document 很常见的结构就是：

```text
Document
├── 普通 Field
├── Embedded Document
└── Array
```

---

# 5. Reference：引用

但不是所有数据都适合塞进一个 Document。

例如：

```text
Category
toy

Product
LEGO Technic
LEGO City
LEGO Creator
...
```

如果把完整 Category 信息嵌进每一个 Product：

```javascript
{
    name: "LEGO Technic",

    category: {
        id: "toy",
        name: "玩具",
        description: "儿童与收藏玩具..."
    }
}
```

每个商品都复制一份。

另一种方式就是：

### categories

```javascript
db.categories.insertOne({
    _id: "toy",
    name: "玩具",
    description: "儿童与收藏玩具"
})
```

### products

```javascript
db.products.insertOne({
    name: "LEGO Technic",
    price: 599,
    categoryId: "toy"
})
```

关系：

```text
products
{
    name: "LEGO Technic",
    categoryId: "toy"
}
        │
        │ reference
        ▼
categories
{
    _id: "toy",
    name: "玩具"
}
```

这就是 **Reference**。

MongoDB 的 Reference 本质上就是在 Document 中保存另一个 Document 的标识，并不会像关系型数据库外键那样自动帮你加载数据。

---

# 6. Reference 怎么查询？

先查 Product：

```javascript
let product = db.products.findOne({
    name: "LEGO Technic"
})
```

得到：

```javascript
{
    name: "LEGO Technic",
    categoryId: "toy"
}
```

再：

```javascript
db.categories.findOne({
    _id: product.categoryId
})
```

这和应用代码里的：

```text
查 Product
    ↓
拿 categoryId
    ↓
查 Category
```

非常接近。

MongoDB 也可以通过 `$lookup` 做类似 JOIN 的操作，不过我们前面已经学过 Aggregation，这里只知道它存在即可，不深入。

---

# 7. Embedded 还是 Reference？

这才是这一课真正重要的部分。

## Embedded

适合：

```text
经常一起查询
+
生命周期基本一致
+
数据规模不会无限增长
```

例如：

```javascript
{
    name: "MacBook Pro",

    attributes: {
        cpu: "...",
        memory: "..."
    }
}
```

很自然。

官方也推荐在“contains / has-a”、经常一起查询或一起修改的数据上优先考虑嵌入。

---

## Reference

适合：

```text
数据本身可以独立存在

或者

被很多 Document 共同引用

或者

嵌入数据可能无限增长
```

例如：

```text
Category
   ↑
   ├── Product A
   ├── Product B
   ├── Product C
   └── ...
```

如果 Category 经常修改，复制到几万个 Product 里显然不方便。

复杂多对多关系、子数据数量非常大等情况，也更适合考虑 Reference。

---

# 8. 一个非常实用的判断方法

以后遇到两个对象：

```text
A
B
```

可以先问：

> **业务读取 A 的时候，是不是绝大多数时候都需要 B？**

如果：

```text
是
+
B 很小
+
B 不会无限增长
```

可以优先考虑：

```text
Embedded
```

如果：

```text
B 可以独立存在

或者

很多 A 共享 B

或者

B 数量可能非常大
```

更倾向：

```text
Reference
```

MongoDB 数据建模一个很重要的思想就是：

> **围绕应用的读取和修改方式设计 Document，而不是先像关系型数据库一样追求数据完全规范化。**

官方数据建模指南也强调，要根据应用如何访问相关数据决定嵌入还是引用。

---

# 9. 一个真实一点的订单例子

这个例子很能体现 MongoDB 思维。

假设订单有：

```text
Order
+
Shipping Address
```

地址很适合直接嵌进去：

```javascript
db.orders.insertOne({
    orderNo: "ORDER-001",
    totalPrice: 1298,

    shippingAddress: {
        province: "Tokyo",
        city: "Shinjuku",
        detail: "..."
    }
})
```

为什么？

因为订单查看时：

```text
订单
+
当时的收货地址
```

通常就是一起读。

而且我们甚至希望：

> 用户以后修改地址，也不要改变历史订单里的地址。

这种情况下复制一份地址“快照”反而是正确业务行为。

---

# 10. 一个需要注意的限制

MongoDB 单个 BSON Document 有：

> **16 MiB 大小限制。** 

所以这种设计要小心：

```javascript
{
    userId: 1,

    logs: [
        // 无限增加
        ...
    ]
}
```

如果数组会无限增长：

```text
100
10000
1000000
...
```

就不应该一直 Embedded。

更合理可能是：

```text
users Collection

logs Collection
    ↓
userId Reference
```

所以：

> **Embedded 不是“全部塞进一个 Document”。**

---

# 11. Go 中表示 Embedded

初始化：

```bash
go mod init mongodb-lab/lesson-07-modeling
go get go.mongodb.org/mongo-driver/v2/mongo
```

Go Struct 非常自然：

```go
type Attributes struct {
	CPU     string `bson:"cpu"`
	Memory  string `bson:"memory"`
	Storage string `bson:"storage"`
}

type Product struct {
	Name       string     `bson:"name"`
	Price      int        `bson:"price"`
	Attributes Attributes `bson:"attributes"`
	Tags       []string   `bson:"tags"`
}
```

写入：

```go
product := Product{
	Name:  "MacBook Pro",
	Price: 15999,

	Attributes: Attributes{
		CPU:     "M5",
		Memory:  "32GB",
		Storage: "1TB",
	},

	Tags: []string{
		"computer",
		"apple",
	},
}

_, err := collection.InsertOne(
	ctx,
	product,
)
```

最后 MongoDB 中就是：

```javascript
{
    name: "MacBook Pro",
    price: 15999,

    attributes: {
        cpu: "M5",
        memory: "32GB",
        storage: "1TB"
    },

    tags: [
        "computer",
        "apple"
    ]
}
```

---

# 12. Go 中表示 Reference

例如：

```go
type Product struct {
	Name       string `bson:"name"`
	Price      int    `bson:"price"`
	CategoryID string `bson:"categoryId"`
}
```

写入：

```go
product := Product{
	Name:       "LEGO Technic",
	Price:      599,
	CategoryID: "toy",
}
```

本质还是：

```text
Product.CategoryID
      ↓
Category._id
```

MongoDB 并不会因为你定义了 `CategoryID` 就自动建立外键关系。

---

# 本课最重要的认识

MongoDB 数据建模不要先想：

```text
应该有几张表？
```

而是先想：

```text
这些数据平时怎么一起读？
这些数据是否一起修改？
它们是不是同一个生命周期？
数据会不会无限增长？
```

然后选择：

```text
                 关联数据
                    │
          ┌─────────┴─────────┐
          ↓                   ↓
       Embedded            Reference

      经常一起读            独立存在
      数据比较小            大量共享
      生命周期一致          可能无限增长
```

这一课最核心就两个词：

> **Embedded：把相关数据放在一起。**  
> **Reference：拆开保存，通过 ID 建立关系。**

掌握到这里已经够用了。下一课我们会进入 **MongoDB Transaction（事务）**，实际用 Go 做一次“两个 Document 要么一起成功，要么一起失败”的操作。