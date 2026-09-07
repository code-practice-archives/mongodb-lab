package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Embedded Document
type Attributes struct {
	CPU     string `bson:"cpu"`
	Memory  string `bson:"memory"`
	Storage string `bson:"storage"`
}

// Category 单独存在于 categories Collection
type Category struct {
	ID          string `bson:"_id"`
	Name        string `bson:"name"`
	Description string `bson:"description"`
}

// Product 同时演示 Embedded 和 Reference
type Product struct {
	Name       string     `bson:"name"`
	Price      int        `bson:"price"`
	CategoryID string     `bson:"categoryId"` // Reference
	Attributes Attributes `bson:"attributes"` // Embedded
	Tags       []string   `bson:"tags"`       // Array
}

func main() {
	ctx := context.Background()

	// 1. 连接 MongoDB
	client, err := mongo.Connect(
		options.Client().
			ApplyURI("mongodb://localhost:27023"),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Println(err)
		}
	}()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	fmt.Println("MongoDB connected")

	db := client.Database("middleware_lab")

	products := db.Collection("products")
	categories := db.Collection("categories")

	// 为了方便重复运行 Demo，先清空本课数据
	if _, err := products.DeleteMany(ctx, bson.M{}); err != nil {
		log.Fatal(err)
	}

	if _, err := categories.DeleteMany(ctx, bson.M{}); err != nil {
		log.Fatal(err)
	}

	// =====================================================
	// 2. Reference：先创建一个独立 Category
	// =====================================================

	category := Category{
		ID:          "computer",
		Name:        "电脑",
		Description: "笔记本与桌面电脑",
	}

	_, err = categories.InsertOne(ctx, category)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("category inserted")

	// =====================================================
	// 3. Embedded + Reference：创建 Product
	// =====================================================

	product := Product{
		Name:       "MacBook Pro",
		Price:      15999,
		CategoryID: "computer",

		// Embedded Document
		Attributes: Attributes{
			CPU:     "M5",
			Memory:  "32GB",
			Storage: "1TB",
		},

		// Array
		Tags: []string{
			"computer",
			"apple",
			"laptop",
		},
	}

	result, err := products.InsertOne(ctx, product)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("product inserted:", result.InsertedID)

	// =====================================================
	// 4. 查询 Embedded Document
	//    attributes.memory = 32GB
	// =====================================================

	fmt.Println("\n=== Query Embedded Document ===")

	var embeddedResult Product

	err = products.FindOne(
		ctx,
		bson.M{
			"attributes.memory": "32GB",
		},
	).Decode(&embeddedResult)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(
		"name=%s cpu=%s memory=%s storage=%s\n",
		embeddedResult.Name,
		embeddedResult.Attributes.CPU,
		embeddedResult.Attributes.Memory,
		embeddedResult.Attributes.Storage,
	)

	// =====================================================
	// 5. 查询 Array
	//    tags 中包含 apple
	// =====================================================

	fmt.Println("\n=== Query Array ===")

	var tagResult Product

	err = products.FindOne(
		ctx,
		bson.M{
			"tags": "apple",
		},
	).Decode(&tagResult)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(
		"name=%s tags=%v\n",
		tagResult.Name,
		tagResult.Tags,
	)

	// =====================================================
	// 6. 修改 Embedded Document
	//    memory: 32GB -> 64GB
	// =====================================================

	fmt.Println("\n=== Update Embedded Document ===")

	updateResult, err := products.UpdateOne(
		ctx,

		bson.M{
			"name": "MacBook Pro",
		},

		bson.M{
			"$set": bson.M{
				"attributes.memory": "64GB",
			},
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("modified:", updateResult.ModifiedCount)

	// =====================================================
	// 7. 修改 Array
	//    添加一个 tag
	// =====================================================

	fmt.Println("\n=== Update Array ===")

	_, err = products.UpdateOne(
		ctx,

		bson.M{
			"name": "MacBook Pro",
		},

		bson.M{
			"$push": bson.M{
				"tags": "developer",
			},
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	// =====================================================
	// 8. Reference 查询
	//
	// Product
	//    ↓ categoryId
	// Category
	//
	// MongoDB 不会像 ORM 外键一样自动加载 Category
	// =====================================================

	fmt.Println("\n=== Query Reference ===")

	var productResult Product

	err = products.FindOne(
		ctx,
		bson.M{
			"name": "MacBook Pro",
		},
	).Decode(&productResult)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(
		"product categoryId:",
		productResult.CategoryID,
	)

	// 根据 categoryId 再查 categories
	var categoryResult Category

	err = categories.FindOne(
		ctx,
		bson.M{
			"_id": productResult.CategoryID,
		},
	).Decode(&categoryResult)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(
		"category: %+v\n",
		categoryResult,
	)

	// =====================================================
	// 9. 查询最终 Product
	// =====================================================

	fmt.Println("\n=== Final Product ===")

	var finalProduct Product

	err = products.FindOne(
		ctx,
		bson.M{
			"name": "MacBook Pro",
		},
	).Decode(&finalProduct)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", finalProduct)
}
