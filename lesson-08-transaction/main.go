package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Account struct {
	ID      string `bson:"_id"`
	Balance int    `bson:"balance"`
}

func main() {
	ctx := context.Background()

	// 1. 连接 Replica Set
	client, err := mongo.Connect(
		options.Client().ApplyURI(
			"mongodb://localhost:27024/?replicaSet=rs0&directConnection=true",
		),
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

	accounts := client.
		Database("middleware_lab").
		Collection("accounts")

	// 2. 每次运行 Demo 时重新准备数据
	if _, err := accounts.DeleteMany(ctx, bson.M{}); err != nil {
		log.Fatal(err)
	}

	_, err = accounts.InsertMany(
		ctx,
		[]any{
			Account{
				ID:      "A",
				Balance: 1000,
			},
			Account{
				ID:      "B",
				Balance: 500,
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n=== Initial ===")
	printAccounts(ctx, accounts)

	// =========================================
	// 3. 故意失败的事务
	// =========================================

	fmt.Println("\n=== Failed Transaction ===")

	err = transferWithError(
		ctx,
		client,
		accounts,
		"A",
		"B",
		100,
	)

	fmt.Println("transaction error:", err)

	fmt.Println("\n=== After Failed Transaction ===")
	printAccounts(ctx, accounts)

	// =========================================
	// 4. 正常事务
	// =========================================

	fmt.Println("\n=== Successful Transaction ===")

	err = transfer(
		ctx,
		client,
		accounts,
		"A",
		"B",
		100,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n=== After Successful Transaction ===")
	printAccounts(ctx, accounts)
}

// 正常转账
func transfer(
	ctx context.Context,
	client *mongo.Client,
	accounts *mongo.Collection,
	from string,
	to string,
	amount int,
) error {

	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(
		ctx,
		func(txCtx context.Context) (any, error) {

			// A 扣钱
			result, err := accounts.UpdateOne(
				txCtx,
				bson.M{
					"_id":     from,
					"balance": bson.M{"$gte": amount},
				},
				bson.M{
					"$inc": bson.M{
						"balance": -amount,
					},
				},
			)

			if err != nil {
				return nil, err
			}

			if result.MatchedCount == 0 {
				return nil, errors.New("账户不存在或余额不足")
			}

			// B 加钱
			result, err = accounts.UpdateOne(
				txCtx,
				bson.M{
					"_id": to,
				},
				bson.M{
					"$inc": bson.M{
						"balance": amount,
					},
				},
			)

			if err != nil {
				return nil, err
			}

			if result.MatchedCount == 0 {
				return nil, errors.New("目标账户不存在")
			}

			return nil, nil
		},
	)

	return err
}

// 故意制造错误，观察事务回滚
func transferWithError(
	ctx context.Context,
	client *mongo.Client,
	accounts *mongo.Collection,
	from string,
	to string,
	amount int,
) error {

	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(
		ctx,
		func(txCtx context.Context) (any, error) {

			// 第一步：A - 100
			_, err := accounts.UpdateOne(
				txCtx,
				bson.M{
					"_id": from,
				},
				bson.M{
					"$inc": bson.M{
						"balance": -amount,
					},
				},
			)

			if err != nil {
				return nil, err
			}

			// 模拟程序在两步操作之间发生异常
			return nil, errors.New("模拟程序异常")

			// B + 100 不会执行
		},
	)

	return err
}

// 打印账户余额
func printAccounts(
	ctx context.Context,
	accounts *mongo.Collection,
) {

	cursor, err := accounts.Find(
		ctx,
		bson.M{},
		options.Find().SetSort(
			bson.D{
				{Key: "_id", Value: 1},
			},
		),
	)

	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	var result []Account

	if err := cursor.All(ctx, &result); err != nil {
		log.Fatal(err)
	}

	for _, account := range result {
		fmt.Printf(
			"%s = %d\n",
			account.ID,
			account.Balance,
		)
	}
}
