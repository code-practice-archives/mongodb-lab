# Lesson 11：MongoDB 选型——什么时候该用，什么时候不该用

这一课不学新的 MongoDB API，目标是把前面学的东西串起来，最终能回答面试里最常见的问题：

> **为什么不用 MySQL，而要选择 MongoDB？**

---

## 第一步：先建立选型认识

MongoDB 最核心的优势不是“比 MySQL 快”，而是：

> **Document 模型非常适合结构灵活、嵌套明显、经常整体读写的数据。**

例如商品：

```javascript
{
    name: "MacBook Pro",
    type: "computer",
    attributes: {
        cpu: "M5",
        memory: "32GB"
    }
}
```

另一个商品：

```javascript
{
    name: "LEGO Technic",
    type: "toy",
    attributes: {
        pieces: 1500,
        age: "10+"
    }
}
```

如果放 MySQL：

```text
products
├── cpu
├── memory
├── pieces
├── age
├── color
├── size
├── ...
```

会出现大量不同类型商品对应不同字段的问题。

MongoDB：

```text
Product
└── attributes
      ↓
每种商品可以拥有不同结构
```

这就是 MongoDB 很典型的价值。

### 简单选型表

| 场景 | 更倾向 |
|---|---|
| 强事务、复杂关联 | MySQL |
| 结构固定的核心业务数据 | MySQL |
| 灵活 Document、嵌套数据 | MongoDB |
| 商品属性、内容、配置等结构变化较多 | MongoDB |
| 缓存、计数器、Session | Redis |
| 极低延迟 KV | Redis |

不要记成：

```text
MongoDB = 大数据
```

更准确是：

```text
MongoDB
=
Document 数据模型
+
Flexible Schema
+
比较强的查询/索引/聚合能力
```

---

# 第二步：做一个选型实验

仍然为这一课单独建目录：

```text
lesson-11-selection/
└── compose.yaml
```

`compose.yaml`：

```yaml
services:
  mongodb:
    image: mongodb/mongodb-community-server:8.0.29-ubi9-slim
    container_name: mongodb-lesson-11

    ports:
      - "27034:27017"

    volumes:
      - mongodb_lesson_11_data:/data/db

volumes:
  mongodb_lesson_11_data:
```

启动：

```bash
docker compose up -d
```

连接：

```bash
mongosh "mongodb://localhost:27034"
```

创建：

```javascript
use middleware_lab
```

插入完全不同类型的商品：

```javascript
db.products.insertMany([
    {
        name: "MacBook Pro",
        type: "computer",
        price: 15999,
        attributes: {
            cpu: "M5",
            memory: "32GB",
            storage: "1TB"
        }
    },

    {
        name: "LEGO Technic",
        type: "toy",
        price: 599,
        attributes: {
            pieces: 1500,
            age: "10+"
        }
    },

    {
        name: "T-Shirt",
        type: "clothing",
        price: 199,
        attributes: {
            color: "black",
            size: ["M", "L", "XL"]
        }
    }
])
```

你现在看到的其实就是 MongoDB 一个很典型的业务场景：

```text
products Collection

Computer Document
├── cpu
├── memory
└── storage

Toy Document
├── pieces
└── age

Clothing Document
├── color
└── size
```

但依然能统一查询：

```javascript
db.products.find({
    price: {
        $gt: 500
    }
})
```

也能查某类特有字段：

```javascript
db.products.find({
    "attributes.memory": "32GB"
})
```

这就是：

> **结构灵活，但依然保留数据库的查询、索引和聚合能力。**

这也是 MongoDB 和“直接存 JSON 文件”完全不同的地方。

---

# 第三步：Go 里有什么体现？

Go 里最明显的问题是：

如果 `attributes` 每种商品完全不同，就不一定适合定义一个死板 Struct。

可以：

```go
type Product struct {
	Name       string         `bson:"name"`
	Type       string         `bson:"type"`
	Price      int            `bson:"price"`
	Attributes map[string]any `bson:"attributes"`
}
```

例如电脑：

```go
computer := Product{
	Name:  "MacBook Pro",
	Type:  "computer",
	Price: 15999,

	Attributes: map[string]any{
		"cpu":     "M5",
		"memory":  "32GB",
		"storage": "1TB",
	},
}
```

玩具：

```go
toy := Product{
	Name:  "LEGO Technic",
	Type:  "toy",
	Price: 599,

	Attributes: map[string]any{
		"pieces": 1500,
		"age":    "10+",
	},
}
```

然后仍然：

```go
collection.InsertOne(ctx, computer)
collection.InsertOne(ctx, toy)
```

这就很好地体现：

```text
Go
↓
灵活的数据结构
↓
BSON Document
↓
MongoDB
```

---

# 最后形成一个选型判断

以后碰到一个业务，可以先问：

```text
数据结构稳定吗？
        ↓
大量关联和强事务吗？
        ↓
是
→ MySQL
```

如果：

```text
数据天然像一个 Document
+
经常整体读取
+
字段可能因类型不同而变化
+
嵌套结构很多
        ↓
MongoDB
```

如果只是：

```text
Key → Value
+
要求极低延迟
+
可以接受内存型存储
        ↓
Redis
```

所以这一整个 MongoDB 基础学习最后最值得你带走的一句话是：

> **MongoDB 的核心价值不是替代 MySQL，而是在数据天然呈 Document 结构、Schema 灵活且经常整体读写时，提供一种比关系模型更自然的数据组织方式。**

到这里，我们最初那份 **MongoDB 基础学习清单其实已经基本跑完了**。下一步比较适合做的不是继续深挖 MongoDB，而是换下一个你工作里没接触过的中间件，用同样的方法继续学。