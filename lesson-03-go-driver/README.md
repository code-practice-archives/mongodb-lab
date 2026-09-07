好，第三课我们开始 **Go Driver 实践**。这一课不引入索引、聚合之类的新主题，只把前两课的 MongoDB 基础操作迁移到 Go 代码里。

MongoDB 当前官方 Go Driver 使用 `go.mongodb.org/mongo-driver/v2/...` 包路径；连接、CRUD 的基本 API 仍然是 `mongo.Connect`、`InsertOne`、`Find`、`UpdateOne` 等。

# Lesson 03：使用 Go 操作 MongoDB

## 1. 创建独立目录

```text
mongodb-lab/
├── lesson-01-basic/
├── lesson-02-crud/
└── lesson-03-go-driver/
```

创建：

```bash
mkdir lesson-03-go-driver
cd lesson-03-go-driver

go mod init mongodb-lab/lesson-03-go-driver
```

这一课仍然启动自己的 MongoDB。

`compose.yaml`：

```yaml
services:
  mongodb:
    image: mongodb/mongodb-community-server:8.0.29-ubi9-slim
    container_name: mongodb-lesson-03
    ports:
      - "27019:27017"
    volumes:
      - mongodb_lesson_03_data:/data/db

volumes:
  mongodb_lesson_03_data:
```

启动：

```bash
docker compose up -d
```

所以这一课的地址是：

```text
mongodb://localhost:27019
```

---

## 2. 安装 Go Driver

执行：

```bash
go get go.mongodb.org/mongo-driver/v2/mongo
```

官方当前 Go Driver 就是通过这个 v2 模块使用。

---

# 3. 第一个 Go 程序：连接 MongoDB

先创建：

```text
lesson-03-go-driver/
├── compose.yaml
├── go.mod
├── go.sum
└── main.go
```

`main.go`：

```go
package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	const uri = "mongodb://localhost:27019"

	client, err := mongo.Connect(
		options.Client().ApplyURI(uri),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Println(err)
		}
	}()

	// Connect 创建 Client 后，再通过 Ping 验证 MongoDB 是否真的可访问。
	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatal(err)
	}

	fmt.Println("MongoDB connected")
}
```

运行：

```bash
go run .
```

看到：

```text
MongoDB connected
```

就说明 Go 已经成功连接 MongoDB。

官方当前示例也是通过 Connection URI 创建 `MongoClient`，并用 `Ping()` 验证连接。

这里先建立一个对应关系：

```text
mongosh
mongodb://localhost:27019

        ↓

Go

mongo.Connect(...)
```

---

# 4. 获取 Database 和 Collection

前面 `mongosh` 中：

```javascript
use middleware_lab
db.products
```

Go 中则是：

```go
db := client.Database("middleware_lab")

collection := db.Collection("products")
```

甚至通常直接写：

```go
collection := client.
	Database("middleware_lab").
	Collection("products")
```

官方 Driver 就是通过：

```go
client.Database(...)
```

获取 Database，再通过：

```go
Collection(...)
```

获取 Collection。

---

# 5. 定义 Product

创建一个我们熟悉的结构：

```go
type Product struct {
	Name     string `bson:"name"`
	Category string `bson:"category"`
	Price    int    `bson:"price"`
	Stock    int    `bson:"stock"`
}
```

这里最值得注意的是：

```go
bson:"name"
```

和你平时写 JSON：

```go
json:"name"
```

很像。

比如：

```go
Product{
	Name:  "LEGO Technic",
	Price: 599,
}
```

写入 MongoDB 后大致就是：

```javascript
{
    name: "LEGO Technic",
    price: 599
}
```

这就是：

> Go Struct ↔ BSON Document

---

# 6. InsertOne

在 `Ping` 后增加：

```go
collection := client.
	Database("middleware_lab").
	Collection("products")

product := Product{
	Name:     "LEGO Technic",
	Category: "toy",
	Price:    599,
	Stock:    20,
}

result, err := collection.InsertOne(
	context.Background(),
	product,
)
if err != nil {
	log.Fatal(err)
}

fmt.Println("inserted id:", result.InsertedID)
```

运行：

```bash
go run .
```

会得到：

```text
inserted id: ObjectID(...)
```

如果没有指定 `_id`，Driver 会自动生成 `ObjectId`。

它对应我们第二课的：

```javascript
db.products.insertOne({
    name: "LEGO Technic",
    category: "toy",
    price: 599,
    stock: 20
})
```

---

# 7. FindOne

现在查询这条数据。

需要增加：

```go
"go.mongodb.org/mongo-driver/v2/bson"
```

代码：

```go
var product Product

err := collection.FindOne(
	context.Background(),
	bson.M{
		"name": "LEGO Technic",
	},
).Decode(&product)

if err != nil {
	log.Fatal(err)
}

fmt.Printf("%+v\n", product)
```

这里：

```go
bson.M{
	"name": "LEGO Technic",
}
```

其实就是前面学习过的：

```javascript
{
    name: "LEGO Technic"
}
```

也就是：

> Filter。

所以你会发现前两课并没有白学。

```text
mongosh

db.products.find({
    name: "LEGO Technic"
})
```

对应 Go：

```go
collection.FindOne(
	ctx,
	bson.M{"name": "LEGO Technic"},
)
```

官方 Driver 中 `FindOne()` 查询后通常通过 `Decode()` 解码进 Go 变量。

---

# 8. UpdateOne

把价格：

```text
599
```

修改成：

```text
699
```

代码：

```go
result, err := collection.UpdateOne(
	context.Background(),

	// filter
	bson.M{
		"name": "LEGO Technic",
	},

	// update
	bson.M{
		"$set": bson.M{
			"price": 699,
		},
	},
)

if err != nil {
	log.Fatal(err)
}

fmt.Println("modified:", result.ModifiedCount)
```

其实和 `mongosh` 几乎完全一样：

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

所以 MongoDB Go Driver 很好理解：

> MongoDB 原本那些 Filter、`$set`、`$inc` 等语法没有消失，只是换成 Go 的 BSON 数据结构表达。

官方的 `UpdateOne()` 也是 `filter + update` 这种形式。

---

# 9. DeleteOne

删除：

```go
result, err := collection.DeleteOne(
	context.Background(),
	bson.M{
		"name": "LEGO Technic",
	},
)

if err != nil {
	log.Fatal(err)
}

fmt.Println("deleted:", result.DeletedCount)
```

对应：

```javascript
db.products.deleteOne({
    name: "LEGO Technic"
})
```

---

# 10. 最终完整 Demo

现在可以把这一课收敛成一个完整程序：

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

	// 1. 连接 MongoDB
	client, err := mongo.Connect(
		options.Client().
			ApplyURI("mongodb://localhost:27019"),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	// 2. 获取 Collection
	collection := client.
		Database("middleware_lab").
		Collection("products")

	// 3. Insert
	product := Product{
		Name:     "LEGO Technic",
		Category: "toy",
		Price:    599,
		Stock:    20,
	}

	insertResult, err := collection.InsertOne(ctx, product)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("insert:", insertResult.InsertedID)

	// 4. Find
	var result Product

	err = collection.FindOne(
		ctx,
		bson.M{"name": "LEGO Technic"},
	).Decode(&result)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("find: %+v\n", result)

	// 5. Update
	updateResult, err := collection.UpdateOne(
		ctx,
		bson.M{"name": "LEGO Technic"},
		bson.M{
			"$set": bson.M{
				"price": 699,
			},
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("update:", updateResult.ModifiedCount)

	// 6. Delete
	deleteResult, err := collection.DeleteOne(
		ctx,
		bson.M{"name": "LEGO Technic"},
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("delete:", deleteResult.DeletedCount)
}
```

执行：

```bash
go run .
```

应该看到类似：

```text
insert: ObjectID(...)
find: {Name:LEGO Technic Category:toy Price:599 Stock:20}
update: 1
delete: 1
```

---

## 这一课最重要的不是代码，而是这个映射

```text
Mongo Shell                  Go Driver

use db                       client.Database()

db.products                  db.Collection()

insertOne()                  InsertOne()

find()                       Find / FindOne()

updateOne()                  UpdateOne()

deleteOne()                  DeleteOne()

{ category: "toy" }          bson.M{"category": "toy"}

Document                     Go Struct
                             +
                             bson tag
```

所以你现在应该产生一种感觉：

> **Go Driver 并没有创造一套新的 MongoDB 查询语言。MongoDB 的 Filter 和操作符还是那些，只是使用 `bson.M`、Struct 等 Go 类型来表达。**
