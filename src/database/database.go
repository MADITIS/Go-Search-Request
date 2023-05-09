package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/maditis/search-go/src/config"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Ctx = context.Background()
var Redis *redis.Client
var MongoDB *mongo.Collection

func InitRedis() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     config.EnvFields.RedisURL,
		Password: config.EnvFields.RedisPassword,
		DB:       0, // use default DB
	})
	fmt.Println("Init DB", Redis)
}

func GetValue(key string) (string, bool) {
	var value string
	value, err := Redis.Get(Ctx, key).Result()
	if err == redis.Nil {
		fmt.Println("Key Does Not Exists", key, "Error: ", err.Error())
		return value, false
	} else if err != nil {
		fmt.Println("Some Thing Horrible has gone wrong", key, "Error: ", err.Error())
		return value, false
	} else {
		return value, true
	}
}

func SetValue(key string, value []map[string]string) {
	val, er := json.Marshal(value)
	if er != nil {
		fmt.Printf("Could Not set value")
		return
	}
	err := Redis.Set(Ctx, key, val, 1*time.Hour).Err()
	// Redis.LSet()
	if err != nil {
		fmt.Println("Could Not set Key: ", key, "Error: ", err)
	}
	fmt.Println("key Set Successfully: ", key)
}

func InitMongo() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s", config.EnvFields.MongoUser, config.EnvFields.MongoPass, config.EnvFields.MongoDBURL)))
	if err != nil {
		panic(err)
	}
	err = client.Ping(Ctx, nil)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")


	MongoDB = client.Database("admin").Collection("Requests")
}

func InsertOne(userName string, userID int64, messageID int, datetime string, requestLink string) {
	item := bson.D{
		{Key: "RequestedBy", Value: bson.D{
			{Key: "UserName", Value: userName},
			{Key: "UserID", Value: userID},
			{Key: "DateTime", Value: datetime},
			{Key: "MessageID", Value: messageID},
			{Key: "MessageLink", Value: requestLink},
		}},
	}
	MongoDB.InsertOne(context.TODO(), item)
}

func UpdateOne(userKey string, userID int64, messageID int, datetime string, userValue string, requesterID int64) bool {
	var result RequestData
	filter := bson.D{{Key: "RequestedBy.UserID", Value: requesterID}, {Key: "RequestedBy.MessageID", Value: messageID}}
	fmt.Println(filter)
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "PickedBy", Value: bson.D{{Key: userKey, Value: userValue},
			{Key: "UserID", Value: userID}, {Key: "DateTime", Value: datetime}}}}},
	}

	err := MongoDB.FindOneAndUpdate(Ctx, filter, update).Decode(&result)
	return err == nil

}

type RequestInfo struct {
	UserName    string
	FirstName   string
	UserID      int64
	DateTime    string
	MessageLink string
	MessageID   int
}
type RequestData struct {
	ID          primitive.ObjectID `bson:"_id"`
	RequestedBy RequestInfo
	PickedBy    RequestInfo
	CompletedBy RequestInfo
}

func FindOne(messageID int, requesterID int64) (RequestData, error) {
	var result RequestData
	filter := bson.D{{Key: "RequestedBy.UserID", Value: requesterID}, {Key: "RequestedBy.MessageID", Value: messageID}}
	err := MongoDB.FindOne(Ctx, filter).Decode(&result)
	if err != nil {
		fmt.Println("Some error occured", err.Error())
		return result, err
	}
	return result, nil

}

func CanCancel(userid int64, messageID int) bool {
	var temp RequestData
	filter := bson.D{{Key: "PickedBy.UserID", Value: userid}, {Key: "RequestedBy.MessageID", Value: messageID}}
	update := bson.D{{Key: "$unset", Value: bson.D{{Key: "PickedBy", Value: ""}}}}

	err := MongoDB.FindOneAndUpdate(Ctx, filter, update).Decode(&temp)
	return err == nil

}

func DeleteRequest(messageID int) bool {
	var temp RequestData
	filter := bson.D{{Key: "RequestedBy.MessageID", Value: messageID}}
	err := MongoDB.FindOneAndDelete(Ctx, filter).Decode(&temp)

	return err == nil
}

func DeleteByUser(userID int64, messageId int) bool {
	var temp RequestData
	filter := bson.D{{Key: "RequestedBy.UserID", Value: userID}, {Key: "RequestedBy.MessageID", Value: messageId}}
	err := MongoDB.FindOneAndDelete(Ctx, filter).Decode(&temp)

	return err == nil
}

func AddComplete(userid int64, messageid int, userKey string, userValue string, datetime string) (RequestData, error) {
	var result RequestData
	filter := bson.D{{Key: "RequestedBy.MessageID", Value: messageid}, {Key: "PickedBy.UserID", Value: userid}}

	print("whattttttttttttttttttttttttt", filter, userid, messageid)
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "CompletedBy", Value: bson.D{{Key: userKey, Value: userValue},
			{Key: "UserID", Value: userid}, {Key: "DateTime", Value: datetime}}}}},
	}

	err := MongoDB.FindOneAndUpdate(Ctx, filter, update).Decode(&result)
	if err != nil {
		fmt.Println("Some error occured", err.Error())
		return result, err
	}
	return result, nil
	// if err != nil {
	// 	fmt.Println("Cant update", err.Error())
	// }
	// if r != nil {
	// 	// print(r)
	// 	id, _ := r.UpsertedID.(primitive.ObjectID)
	// 	print("idiiiiiii", id.Hex(), "hmm",r.UpsertedID)
	// 	err = MongoDB.FindOne(Ctx, bson.D{{Key: "ID", Value: id}}).Decode(&result)
	// 	if err != nil {
	// 		fmt.Println("Some error occured", err.Error())
	// 		return result, err
	// 	}
	// 	fmt.Println("Successfully Completed")
	// 	return result, nil
	// }
	// return result, nil
}

func ClearRequests() bool {
	_, err := MongoDB.DeleteMany(Ctx, bson.D{})

	return err == nil
}

func GetPendingRequests(userID int64, total bool, completed bool) ([]RequestData, error) {
	var (
		results []RequestData
		filter  primitive.D
	)
	if !completed {
		switch total {
		case true:
			filter = bson.D{{Key: "RequestedBy.UserID", Value: userID}}
		case false:
			filter = bson.D{{Key: "RequestedBy.UserID", Value: userID}, {Key: "CompletedBy", Value: bson.D{{Key: "$exists", Value: false}}}}
		}
	} else {
		filter = bson.D{
			{Key: "CompletedBy", Value: bson.D{{Key: "$exists", Value: true}}},
			{Key: "CompletedBy.UserID", Value: userID},
		}
	}
	cursor, err := MongoDB.Find(Ctx, filter)
	if err != nil {
		fmt.Println("Some Horrible Things has happened")
		return results, err
	}
	if err = cursor.All(Ctx, &results); err != nil {
		fmt.Println("Some Horrible Things has happened")
		return results, err
	}

	for _, result := range results {
		cursor.Decode(&result)
		output, err := json.MarshalIndent(result, "", "    ")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", output)
	}
	return results, nil

}
