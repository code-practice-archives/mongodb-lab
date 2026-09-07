明白，这篇就应该像一张**第一课手册页**，只保留“学了什么、命令是什么、干什么”，不再重复展开。

# MongoDB Lesson 01：基础手册

## 核心概念

| MongoDB | 类比 MySQL | 说明 |
|---|---|---|
| Database | Database | 数据库 |
| Collection | Table | Document 的集合 |
| Document | Row | 一条数据，类似 JSON |
| Field | Column | Document 中的字段 |
| `_id` | Primary Key | Document 唯一标识 |
| BSON | — | MongoDB 实际存储的数据格式 |

### 本课重点

- MongoDB 以 **Document** 为核心存储数据。
- 同一个 Collection 中的 Document 可以拥有不同字段，即 **Flexible Schema**。
- Document 可以直接包含嵌套对象和数组。
- Database 和 Collection 可以在第一次写入数据时自动创建。

---

## MongoDB 常用命令

| 命令 | 作用 |
|---|---|
| `mongosh "mongodb://localhost:27017"` | 连接 MongoDB |
| `db` | 查看当前 Database |
| `use middleware_lab` | 切换 Database |
| `show dbs` | 查看已有 Database |
| `show collections` | 查看当前 Database 的 Collection |
| `db.products.insertOne({...})` | 插入一个 Document |
| `db.products.find()` | 查询 Collection 中的 Document |
| `db.runCommand({ hello: 1 })` | 检查 MongoDB 服务 |
| `exit` | 退出 mongosh |

### 插入示例

```javascript
db.products.insertOne({
    name: "LEGO Technic",
    category: "toy",
    price: 599
})
```

### 嵌套与数组

```javascript
db.products.insertOne({
    name: "iPhone",
    attributes: {
        color: "black",
        storage: "256GB"
    },
    tags: ["phone", "apple"]
})
```

---

## Docker 环境

Compose 核心配置：

```yaml
services:
  mongodb:
    image: mongodb/mongodb-community-server:8.0.29-ubi9-slim
    ports:
      - "27017:27017"
    volumes:
      - mongodb_lesson_01_data:/data/db

volumes:
  mongodb_lesson_01_data:
```

| 命令 | 作用 |
|---|---|
| `docker compose up -d` | 启动 MongoDB |
| `docker compose ps` | 查看运行状态 |
| `docker compose logs mongodb` | 查看日志 |
| `docker compose down` | 停止容器，保留数据 |
| `docker volume ls` | 查看数据卷 |

> `mongodb_lesson_01_data:/data/db` 用于持久化 MongoDB 数据。