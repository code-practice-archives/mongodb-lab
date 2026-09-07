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
	SKU      string `bson:"sku"`
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
			ApplyURI("mongodb://localhost:27021"),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(ctx)

	collection := client.
		Database("middleware_lab").
		Collection("products")

	// 2. 创建 sku 索引
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "sku", Value: 1},
		},
	}

	indexName, err := collection.
		Indexes().
		CreateOne(ctx, indexModel)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("index created:", indexName)

	// 3. 使用索引字段查询
	var product Product

	err = collection.FindOne(
		ctx,
		bson.M{
			"sku": "SKU-09000",
		},
	).Decode(&product)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("product: %+v\n", product)
}
