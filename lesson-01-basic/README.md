# Lesson 01：启动 MongoDB 与认识 Document

## 这一课要完成什么

这一课只做四件事：

1. 使用 Docker 启动一个 MongoDB。
2. 使用 `mongosh` 连接 MongoDB。
3. 理解 `Database / Collection / Document / Field`。
4. 插入并查询第一批数据，感受 MongoDB 的灵活 Schema。

暂时不涉及：

- Go Driver
- 完整 CRUD
- 索引
- 聚合
- 事务
- Replica Set
- Sharding

这些全部留给后面的课程。

---

# 1. 创建独立课程目录

整个学习仓库：

```text
mongodb-lab/
└── lesson-01-basic/
    ├── compose.yaml
    └── README.md
```

创建：

```bash
mkdir -p mongodb-lab/lesson-01-basic
cd mongodb-lab/lesson-01-basic
```

以后：

```text
lesson-01-basic
lesson-02-crud
lesson-03-go-driver
lesson-04-index
...
```

互相不覆盖。

---

# 2. 创建 compose.yaml

创建：

```bash
touch compose.yaml
```

内容：

```yaml
services:
  mongodb:
    image: mongodb/mongodb-community-server:8.0.29-ubi9-slim
    container_name: mongodb-lesson-01

    ports:
      - "27017:27017"

    volumes:
      - mongodb_lesson_01_data:/data/db

volumes:
  mongodb_lesson_01_data:
```

这里有两个隔离点。

容器：

```text
mongodb-lesson-01
```

数据卷：

```text
mongodb_lesson_01_data
```

以后 Lesson 02 会使用自己的容器和自己的 Volume，不会覆盖第一课的数据。

---

# 3. 启动 MongoDB

执行：

```bash
docker compose up -d
```

检查：

```bash
docker compose ps
```

正常情况下能看到：

```text
mongodb-lesson-01   running
```

也可以：

```bash
docker ps
```

查看日志：

```bash
docker compose logs mongodb
```

看到 MongoDB 正常启动即可。

---

# 4. 准备 mongosh

`mongod` 是：

```text
MongoDB Server
```

而：

```text
mongosh
```

是我们操作 MongoDB 的命令行客户端。

可以类比：

```text
MySQL Server       → mysqld
MySQL CLI          → mysql

MongoDB Server     → mongod
MongoDB CLI        → mongosh
```

如果本机已经存在：

```bash
mongosh --version
```

就不用安装。

macOS 如果没有，可以：

```bash
brew tap mongodb/brew
brew install mongosh
```

---

# 5. 连接 MongoDB

执行：

```bash
mongosh "mongodb://localhost:27017"
```

连接成功后会进入：

```text
test>
```

此时我们已经完成：

```text
Terminal
   │
   │ mongosh
   ▼
localhost:27017
   │
   ▼
MongoDB
```

先验证服务器：

```javascript
db.runCommand({ hello: 1 })
```

看到：

```javascript
ok: 1
```

说明 MongoDB 工作正常。

---

# 6. MongoDB 最基础的数据结构

先建立和 MySQL 的对应关系：

| MySQL | MongoDB |
|---|---|
| Database | Database |
| Table | Collection |
| Row | Document |
| Column | Field |

例如 MySQL：

```sql
CREATE TABLE products (
    id BIGINT,
    name VARCHAR(100),
    category VARCHAR(50),
    price DECIMAL(10,2)
);
```

MongoDB 不需要先创建这么一张固定结构的表。

它存的是 Document：

```javascript
{
    name: "LEGO Technic",
    category: "toy",
    price: 599
}
```

整体关系：

```text
Database
   │
   └── Collection
          │
          ├── Document
          │      ├── Field
          │      └── Field
          │
          └── Document
```

例如：

```text
middleware_lab
      │
      └── products
             │
             ├── LEGO Document
             └── MacBook Document
```

目前先简单记：

> Collection ≈ MySQL Table  
> Document ≈ MySQL Row

但它们并不是完全等价的。

---

# 7. 选择 Database

执行：

```javascript
use middleware_lab
```

应该看到：

```text
switched to db middleware_lab
```

查看当前 Database：

```javascript
db
```

返回：

```text
middleware_lab
```

这时候执行：

```javascript
show dbs
```

可能还看不到：

```text
middleware_lab
```

原因是：

> 现在只是切换到了这个 Database，还没有真正写入数据。

这是和 MySQL 使用体验比较明显的区别之一。

---

# 8. 插入第一条 Document

执行：

```javascript
db.products.insertOne({
    name: "LEGO Technic",
    category: "toy",
    price: 599
})
```

会得到类似：

```javascript
{
    acknowledged: true,
    insertedId: ObjectId("...")
}
```

我们没有提前执行：

```sql
CREATE TABLE products
```

但是 MongoDB 已经创建了：

```text
middleware_lab
      │
      └── products
```

并存入了一条 Document。

---

# 9. 查询数据

执行：

```javascript
db.products.find()
```

会看到类似：

```javascript
[
  {
    _id: ObjectId("..."),
    name: "LEGO Technic",
    category: "toy",
    price: 599
  }
]
```

注意这里多出了：

```javascript
_id: ObjectId("...")
```

这是 MongoDB 自动产生的唯一标识。

现在可以暂时类比成：

```sql
id BIGINT PRIMARY KEY
```

这里只建立印象，后面再专门了解 `ObjectId`。

---

# 10. 做一个重要实验：不同结构的 Document

继续插入：

```javascript
db.products.insertOne({
    name: "MacBook Pro",
    category: "computer",
    price: 15999,
    cpu: "M5",
    memory: "32GB",
    tags: ["computer", "apple", "laptop"]
})
```

再次：

```javascript
db.products.find()
```

现在 Collection 中有两条数据。

第一条：

```javascript
{
    name: "LEGO Technic",
    category: "toy",
    price: 599
}
```

第二条：

```javascript
{
    name: "MacBook Pro",
    category: "computer",
    price: 15999,
    cpu: "M5",
    memory: "32GB",
    tags: [
        "computer",
        "apple",
        "laptop"
    ]
}
```

注意：

```text
LEGO
├── name
├── category
└── price


MacBook
├── name
├── category
├── price
├── cpu
├── memory
└── tags
```

它们都存在：

```text
products
```

这个 Collection 里面。

这就是 MongoDB 非常重要的特征：

> 同一个 Collection 中的 Document，不要求字段完全一样。

这就是我们开始理解 MongoDB **Flexible Schema** 的第一步。

---

# 11. 再体验一下嵌套 Document

MongoDB 的 Field 不只能存简单字符串和数字。

再插一件商品：

```javascript
db.products.insertOne({
    name: "iPhone",
    category: "phone",
    price: 7999,

    attributes: {
        color: "black",
        storage: "256GB"
    },

    tags: [
        "phone",
        "apple"
    ]
})
```

查询：

```javascript
db.products.find()
```

注意这里：

```javascript
attributes: {
    color: "black",
    storage: "256GB"
}
```

Document 里面还能嵌套 Document。

数组也可以直接存在 Field 中：

```javascript
tags: [
    "phone",
    "apple"
]
```

这也是 MongoDB 和传统关系型数据库非常明显的使用体验差异。

暂时不用思考底层怎么实现。

先建立一个感觉：

> MongoDB 天然非常适合保存这种类似 JSON 的、有嵌套结构的数据。

---

# 12. 查看 Collection

执行：

```javascript
show collections
```

应该看到：

```text
products
```

查看 Database：

```javascript
show dbs
```

现在应该可以看到：

```text
middleware_lab
```

因为我们已经真正写入数据了。

---

# 13. 测试数据是否持久化

退出：

```javascript
exit
```

停止容器：

```bash
docker compose down
```

重新启动：

```bash
docker compose up -d
```

再次进入：

```bash
mongosh "mongodb://localhost:27017"
```

执行：

```javascript
use middleware_lab
```

然后：

```javascript
db.products.find()
```

数据依然存在。

原因就是 Compose 中：

```yaml
volumes:
  - mongodb_lesson_01_data:/data/db
```

MongoDB 的数据并没有跟着 Container 一起删除。

---

# 14. 第一课结束后的状态

现在你的目录：

```text
mongodb-lab/
└── lesson-01-basic/
    ├── compose.yaml
    └── README.md
```

Docker 中：

```text
Container
mongodb-lesson-01

Volume
mongodb_lesson_01_data
```

MongoDB 中：

```text
middleware_lab
      │
      └── products
             │
             ├── LEGO Technic
             ├── MacBook Pro
             └── iPhone
```

这就是 **Lesson 01 的最终快照**。

以后不要再修改它。

---

# 15. 第一课需要真正记住什么

不是记命令，而是建立下面几个认识。

### MongoDB 的数据层级

```text
Database
   ↓
Collection
   ↓
Document
   ↓
Field
```

和 MySQL 粗略对应：

```text
Database
   ↓
Table
   ↓
Row
   ↓
Column
```

### Document

MongoDB 最核心的数据单位是：

```javascript
{
    name: "MacBook Pro",
    price: 15999
}
```

这种 BSON Document。

### Flexible Schema

同一个 Collection：

```text
products
```

可以有：

```javascript
{
    name: "LEGO",
    pieces: 1500
}
```

也可以有：

```javascript
{
    name: "MacBook",
    cpu: "M5",
    memory: "32GB"
}
```

字段不要求完全一样。

### 支持嵌套和数组

可以直接：

```javascript
{
    attributes: {
        color: "black"
    },

    tags: [
        "apple",
        "phone"
    ]
}
```

这是后面理解 MongoDB 数据建模的重要基础。

---

# 16. 第一课完成检查

执行过下面这些东西，就算完成：

```text
☑ Docker 启动 MongoDB

☑ mongosh 成功连接

☑ 创建 middleware_lab

☑ 创建 products Collection

☑ insertOne()

☑ find()

☑ 看到自动生成的 _id

☑ 插入结构不同的 Document

☑ 使用嵌套 Document

☑ 使用 Array

☑ 重启 Container 后数据仍然存在
```

第一课到此停止。

不要继续加 CRUD，也不要开始写 Go。
