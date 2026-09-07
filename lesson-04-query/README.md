# Lesson 04：Go 查询进阶 —— Find、Cursor、排序与分页

这一课还是基于 Go Driver，不增加复杂概念。目标是把实际业务里最常见的查询补齐：

```text
Find 多条数据
    ↓
Cursor 获取结果
    ↓
排序
    ↓
分页
    ↓
指定返回字段
```

MongoDB Go Driver 的 `Find()` 会返回所有符合 Filter 的 Document，并通过 `Cursor` 读取；排序、`skip`、`limit`、projection 都通过 `FindOptions` 设置。

---

## 1. 创建独立 Lesson

目录：

```text
mongodb-lab/
├── lesson-01-basic/
├── lesson-02-crud/
├── lesson-03-go-driver/
└── lesson-04-query/
```

```bash
mkdir lesson-04-query
cd lesson-04-query

go mod init mongodb-lab/lesson-04-query
go get go.mongodb.org/mongo-driver/v2/mongo
```

`compose.yaml`：

```yaml
services:
  mongodb:
    image: mongodb/mongodb-community-server:8.0.29-ubi9-slim
    container_name: mongodb-lesson-04
    ports:
      - "27020:27017"
    volumes:
      - mongodb_lesson_04_data:/data/db

volumes:
  mongodb_lesson_04_data:
```

启动：

```bash
docker compose up -d
```

这一课连接：

```text
mongodb://localhost:27020
```

---

# 2. 准备 Product

`main.go`：

```go
package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Product struct {
	Name     string `bson:"name"`
	Category string `bson:"category"`
	Price    int    `bson:"price"`
	Stock    int    `bson:"stock"`
}

func main() {
	ctx := context.Background()

	client, err := mongo.Connect(
		options.Client().
			ApplyURI("mongodb://localhost:27020"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.
		Database("middleware_lab").
		Collection("products")

	_ = collection

	fmt.Println("MongoDB connected")
}
```

先运行：

```bash
go run .
```

---

# 3. 准备测试数据

我们仍然使用熟悉的商品。

先用 `mongosh`：

```bash
mongosh "mongodb://localhost:27020"
```

然后：

```javascript
use middleware_lab
```

插入：

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

现在进入真正的 Go 查询。

---

# 4. Find：查询多条数据

上一课：

```go
FindOne()
```

最多拿一条。

这一课：

```go
Find()
```

拿所有符合条件的数据。

例如查询所有 toy：

```go
cursor, err := collection.Find(
	ctx,
	bson.M{
		"category": "toy",
	},
)
if err != nil {
	log.Fatal(err)
}

defer cursor.Close(ctx)
```

注意：

```go
collection.Find(...)
```

并没有直接返回：

```go
[]Product
```

而是返回：

```text
Cursor
```

可以先理解：

> Cursor 是 MongoDB 查询结果的“读取器”。

官方 Go Driver 的 `Find()` 就是返回 `Cursor`，再通过 Cursor 逐条或批量读取结果。

---

# 5. Cursor.All：一次读完

最简单的处理方式：

```go
var products []Product

err = cursor.All(ctx, &products)
if err != nil {
	log.Fatal(err)
}

for _, product := range products {
	fmt.Printf("%+v\n", product)
}
```

结果：

```text
{Name:LEGO Technic Category:toy Price:599 Stock:20}
{Name:LEGO City Category:toy Price:299 Stock:50}
```

整体过程：

```text
Find
 ↓
Cursor
 ↓
cursor.All()
 ↓
[]Product
```

这个写法对于：

> **结果数量比较确定、不会特别大**

的普通业务查询非常方便。

---

# 6. Cursor.Next：逐条读取

Cursor 也可以逐条处理：

```go
cursor, err := collection.Find(
	ctx,
	bson.M{"category": "toy"},
)
if err != nil {
	log.Fatal(err)
}

defer cursor.Close(ctx)

for cursor.Next(ctx) {

	var product Product

	if err := cursor.Decode(&product); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", product)
}

if err := cursor.Err(); err != nil {
	log.Fatal(err)
}
```

这个模型就是：

```text
Cursor
  ↓
Next()
  ↓
Decode()
  ↓
处理一条
  ↓
Next()
```

官方 Driver 同时支持 `cursor.All()` 和 `cursor.Next() + Decode()` 两种消费方式。

现在不用过度纠结怎么选。

简单建立印象：

```text
结果少、想直接拿 []Product
→ cursor.All()

结果很多、希望逐条处理
→ cursor.Next()
```

---

# 7. 排序 Sort

例如：

> 查询商品，按照价格从低到高。

```go
opts := options.Find().
	SetSort(
		bson.D{
			{"price", 1},
		},
	)

cursor, err := collection.Find(
	ctx,
	bson.M{},
	opts,
)
```

这里：

```text
1  = 升序
-1 = 降序
```

所以：

```go
bson.D{{"price", 1}}
```

类似：

```sql
ORDER BY price ASC
```

而：

```go
bson.D{{"price", -1}}
```

就是：

```sql
ORDER BY price DESC
```

MongoDB 官方同样使用 `SetSort()` 配置排序。

---

# 8. `bson.M` 和 `bson.D`

这里第一次值得区分一下。

我们一直在使用：

```go
bson.M{
	"category": "toy",
}
```

`M` 可以简单理解成：

```go
map[string]any
```

也就是：

> 不强调字段顺序。

所以写 Filter 非常方便：

```go
bson.M{
	"category": "toy",
	"price": bson.M{
		"$lt": 1000,
	},
}
```

---

而：

```go
bson.D{
	{"price", -1},
	{"name", 1},
}
```

是一个：

> **有顺序的 BSON Document。**

为什么排序经常使用 `bson.D`？

因为：

```text
先 price DESC
再 name ASC
```

这里顺序是有意义的。

所以目前可以这样记：

```text
普通 Filter
→ bson.M

需要明确字段顺序
→ bson.D
```

例如：

```go
// Filter
bson.M{
	"category": "toy",
}
```

排序：

```go
bson.D{
	{"price", -1},
	{"name", 1},
}
```

---

# 9. Limit

例如：

> 只取价格最高的两个商品。

```go
opts := options.Find().
	SetSort(bson.D{
		{"price", -1},
	}).
	SetLimit(2)

cursor, err := collection.Find(
	ctx,
	bson.M{},
	opts,
)
```

对应 SQL：

```sql
SELECT *
FROM products
ORDER BY price DESC
LIMIT 2;
```

`SetLimit()` 控制最多返回多少 Document。

---

# 10. Skip：实现最基础分页

假设：

```text
page     = 2
pageSize = 2
```

计算：

```go
skip := int64((page - 1) * pageSize)
```

然后：

```go
opts := options.Find().
	SetSort(bson.D{
		{"price", -1},
	}).
	SetSkip(skip).
	SetLimit(int64(pageSize))
```

完整：

```go
page := 2
pageSize := 2

opts := options.Find().
	SetSort(bson.D{
		{"price", -1},
	}).
	SetSkip(int64((page - 1) * pageSize)).
	SetLimit(int64(pageSize))

cursor, err := collection.Find(
	ctx,
	bson.M{},
	opts,
)
```

基本模型和 MySQL 非常像：

```text
MongoDB             MySQL

SetSkip()           OFFSET
SetLimit()          LIMIT
SetSort()           ORDER BY
```

MongoDB 官方特别提醒：使用 `skip` 做分页时最好配合稳定的排序，否则结果顺序没有保证。

---

# 11. Projection：只返回需要的字段

比如接口只需要：

```text
name
price
```

不想返回：

```text
category
stock
```

可以：

```go
opts := options.Find().
	SetProjection(
		bson.M{
			"name":  1,
			"price": 1,
			"_id":   0,
		},
	)

cursor, err := collection.Find(
	ctx,
	bson.M{},
	opts,
)
```

这里：

```text
1 → 返回字段
0 → 不返回字段
```

类似 MySQL：

```sql
SELECT name, price
FROM products;
```

而不是：

```sql
SELECT *
```

官方 Go Driver 使用 `SetProjection()` 控制返回 Document 中包含哪些字段。

---

# 12. 一个比较像真实接口的查询

现在组合起来：

> 查询 toy 商品，按照价格从高到低排序，每页 10 条。

```go
func FindProducts(
	ctx context.Context,
	collection *mongo.Collection,
	category string,
	page int,
	pageSize int,
) ([]Product, error) {

	filter := bson.M{
		"category": category,
	}

	opts := options.Find().
		SetSort(
			bson.D{{"price", -1}},
		).
		SetSkip(
			int64((page - 1) * pageSize),
		).
		SetLimit(
			int64(pageSize),
		)

	cursor, err := collection.Find(
		ctx,
		filter,
		opts,
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []Product

	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	return products, nil
}
```

调用：

```go
products, err := FindProducts(
	ctx,
	collection,
	"toy",
	1,
	10,
)
```

这已经非常像实际项目里的：

```text
GET /products
    ?category=toy
    &page=1
    &pageSize=10
```

后端 Repository 查询代码了。

---

# 这一课最重要的模型

把 Go 的 MongoDB 查询理解成：

```text
Filter
  ↓
collection.Find()
  ↓
FindOptions
├── Sort
├── Skip
├── Limit
└── Projection
  ↓
Cursor
  ↓
All / Next
  ↓
[]Product / Product
```

尤其记住下面几个对应关系：

| MongoDB Go Driver | MySQL 思维 |
|---|---|
| `bson.M` | WHERE 条件 |
| `Find()` | SELECT 多条 |
| `Cursor` | 查询结果游标 |
| `SetSort()` | ORDER BY |
| `SetSkip()` | OFFSET |
| `SetLimit()` | LIMIT |
| `SetProjection()` | SELECT 指定字段 |

做到这里，**MongoDB 最基础的 Go CRUD 和查询操作基本就完整了**。

下一课我们就可以真正进入一个非常重要的新主题：

> **Lesson 05：MongoDB 索引**

到时候会实际做一个非常直观的实验：

```text
没有索引
    ↓
explain()
    ↓
创建索引
    ↓
再次 explain()
    ↓
对比查询扫描了多少 Document
```

这样你会真正看到 MongoDB 索引解决了什么问题，而不只是背“索引能加快查询”。