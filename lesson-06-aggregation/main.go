package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CategoryStats struct {
	Category     string  `bson:"category"`
	ProductCount int     `bson:"productCount"`
	AvgPrice     float64 `bson:"avgPrice"`
	TotalStock   int     `bson:"totalStock"`
}

func main() {
	ctx := context.Background()

	// 1. 连接 MongoDB
	client, err := mongo.Connect(
		options.Client().
			ApplyURI("mongodb://localhost:27022"),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(ctx)

	collection := client.
		Database("middleware_lab").
		Collection("products")

	// 2. 定义 Aggregation Pipeline
	pipeline := mongo.Pipeline{
		{
			{"$match", bson.D{
				{"stock", bson.D{
					{"$gt", 10},
				}},
			}},
		},

		{
			{"$group", bson.D{
				{"_id", "$category"},

				{"productCount", bson.D{
					{"$sum", 1},
				}},

				{"avgPrice", bson.D{
					{"$avg", "$price"},
				}},

				{"totalStock", bson.D{
					{"$sum", "$stock"},
				}},
			}},
		},

		{
			{"$sort", bson.D{
				{"avgPrice", -1},
			}},
		},

		{
			{"$project", bson.D{
				{"_id", 0},
				{"category", "$_id"},
				{"productCount", 1},
				{"avgPrice", 1},
				{"totalStock", 1},
			}},
		},
	}

	// 3. 执行 Aggregation
	cursor, err := collection.Aggregate(
		ctx,
		pipeline,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer cursor.Close(ctx)

	// 4. 读取结果
	var results []CategoryStats

	if err := cursor.All(ctx, &results); err != nil {
		log.Fatal(err)
	}

	// 5. 输出
	for _, result := range results {
		fmt.Printf(
			"category=%s count=%d avgPrice=%.2f totalStock=%d\n",
			result.Category,
			result.ProductCount,
			result.AvgPrice,
			result.TotalStock,
		)
	}
}
