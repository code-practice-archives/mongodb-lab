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
			ApplyURI("mongodb://localhost:27020"),
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

	// 2. 获取 Collection
	collection := client.
		Database("middleware_lab").
		Collection("products")

	// -----------------------------
	// Find + Cursor.All
	// -----------------------------

	fmt.Println("=== 查询所有 toy ===")

	filter := bson.M{
		"category": "toy",
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	var products []Product

	if err := cursor.All(ctx, &products); err != nil {
		log.Fatal(err)
	}

	for _, product := range products {
		fmt.Printf("%+v\n", product)
	}

	// -----------------------------
	// Find + Cursor.Next
	// -----------------------------

	fmt.Println("\n=== 使用 Cursor.Next 逐条读取 ===")

	cursor2, err := collection.Find(
		ctx,
		bson.M{},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor2.Close(ctx)

	for cursor2.Next(ctx) {

		var product Product

		if err := cursor2.Decode(&product); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%+v\n", product)
	}

	if err := cursor2.Err(); err != nil {
		log.Fatal(err)
	}

	// -----------------------------
	// 排序 + 分页
	// -----------------------------

	fmt.Println("\n=== 排序 + 分页 ===")

	page := 1
	pageSize := 2

	opts := options.Find().
		SetSort(
			bson.D{
				{Key: "price", Value: -1},
			},
		).
		SetSkip(
			int64((page - 1) * pageSize),
		).
		SetLimit(
			int64(pageSize),
		)

	cursor3, err := collection.Find(
		ctx,
		bson.M{},
		opts,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor3.Close(ctx)

	var pageProducts []Product

	if err := cursor3.All(ctx, &pageProducts); err != nil {
		log.Fatal(err)
	}

	for _, product := range pageProducts {
		fmt.Printf("%+v\n", product)
	}

	// -----------------------------
	// Projection
	// -----------------------------

	fmt.Println("\n=== 只查询 name 和 price ===")

	projectionOpts := options.Find().
		SetProjection(
			bson.M{
				"_id":   0,
				"name":  1,
				"price": 1,
			},
		)

	cursor4, err := collection.Find(
		ctx,
		bson.M{},
		projectionOpts,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor4.Close(ctx)

	var projectionResults []bson.M

	if err := cursor4.All(
		ctx,
		&projectionResults,
	); err != nil {
		log.Fatal(err)
	}

	for _, result := range projectionResults {
		fmt.Println(result)
	}
}
