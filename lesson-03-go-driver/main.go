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
