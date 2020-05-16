package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//User for user data
type User struct {
	FirstName string
	LastName  string
	UserName  string
	Password  string
	Email     string
}

//Review structure
type Review struct {
	ID       primitive.ObjectID `bson:"_id, omitempty"`
	MovieID  string
	Username string
	Rate     string
	Review   string
	Date     string
}

var clientOptions = options.Client().ApplyURI("mongodb://localhost:27017")
var client, err = mongo.Connect(context.TODO(), clientOptions)
var database = client.Database("movie")
var collection = database.Collection("login")
var reviewcollection = database.Collection("review")

func main() {

	// user := User{"JongHun", "Yun", "dbswhd82", "dbswhd82", "johas@gmail.com"}

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	http.HandleFunc("/login", login)
	http.HandleFunc("/register", signup)
	http.HandleFunc("/writereview", insertreview)
	http.HandleFunc("/getreview", getreview)
	http.HandleFunc("/updatereview", updatereview)
	http.HandleFunc("/deletereview", deletereview)
	http.ListenAndServe(":8000", nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	// var result User
	decoder := json.NewDecoder(r.Body)

	var mem, result User
	err := decoder.Decode(&mem)

	if err != nil {
		panic(err)
	}
	query := bson.M{"username": mem.UserName, "password": mem.Password}

	err = collection.FindOne(context.TODO(), query).Decode(&result)
	if err != nil {
		w.Write([]byte("fail"))

	}
	js, err2 := json.Marshal(result)
	if err2 != nil {
		w.Write([]byte("fail"))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func signup(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	user := User{
		r.Form["firstname"][0],
		r.Form["lastname"][0],
		r.Form["username"][0],
		r.Form["password"][0],
		r.Form["email"][0],
	}

	insertResult, err := collection.InsertOne(context.TODO(), user)
	if err != nil {

		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", insertResult.InsertedID)

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func insertreview(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	currenttime := time.Now()
	var review Review
	err := decoder.Decode(&review)
	review.Date = currenttime.Format("2006-01-02")
	review.ID = primitive.NewObjectID()
	fmt.Println(review.ID)

	insertResult, err := reviewcollection.InsertOne(context.TODO(), review)
	if err != nil {

		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", insertResult.InsertedID)

	query := bson.M{"movieid": review.MovieID}
	cursor, err := reviewcollection.Find(context.TODO(), query)
	if err != nil {
		log.Fatal(err)
	}
	var result []*Review
	fmt.Println("This is insertreivew")
	for cursor.Next(context.TODO()) {
		var elem Review
		err := cursor.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(elem.ID)
		result = append(result, &elem)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	js, err2 := json.Marshal(result)
	if err2 != nil {
		w.Write([]byte("No review"))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	fmt.Println("write Review and return the data")

}

func deletereview(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var elem map[string]string
	err := decoder.Decode(&elem)
	if err != nil {
		fmt.Println("Decoder Error")
	}
	docid, err := primitive.ObjectIDFromHex(elem["ID"])
	if err != nil {
		fmt.Println("this is id encoding error")
	}
	res, err := reviewcollection.DeleteOne(context.TODO(), bson.M{"_id": docid})
	if err != nil {
		fmt.Println("This is Delete result error")
	}
	if res.DeletedCount == 0 {
		fmt.Println("Delete document is not found", res)
	} else {
		fmt.Println("DeleteOne Result:", res)
	}

	query := bson.M{"movieid": elem["movieid"]}
	cursor, err := reviewcollection.Find(context.TODO(), query)
	if err != nil {
		log.Fatal(err)
	}
	var result []*Review
	for cursor.Next(context.TODO()) {
		var elem Review
		err := cursor.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		result = append(result, &elem)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	js, err2 := json.Marshal(result)
	fmt.Println(result)
	if err2 != nil {
		w.Write([]byte("No review"))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	fmt.Println("get Reivews")

}

func updatereview(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var update Review
	err := decoder.Decode(&update)
	if err != nil {
		log.Fatal("Decoding error")
	}
	currenttime := time.Now()
	update.Date = currenttime.Format("2006-01-02")
	query := bson.M{"_id": update.ID}
	result, err := reviewcollection.ReplaceOne(context.TODO(), query, update)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Replaced %v Documents!\n", result.ModifiedCount)

	query2 := bson.M{"movieid": update.MovieID}
	cursor, err := reviewcollection.Find(context.TODO(), query2)
	if err != nil {
		log.Fatal(err)
	}
	var reviews []*Review
	fmt.Println("This is updatetreivew")
	for cursor.Next(context.TODO()) {
		var elem Review
		err := cursor.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		reviews = append(reviews, &elem)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	js, err2 := json.Marshal(reviews)
	if err2 != nil {
		w.Write([]byte("No review"))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	fmt.Println("Update Review and return the data")

	// var elem map[string]string
	// err := decoder.Decode(&elem)
	// if err != nil {
	// 	fmt.Println("Decoder Error")
	// }
	// for key, val := range elem {
	// 	fmt.Println(key, val)
	// }

}

func getreview(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var id Review
	err := decoder.Decode(&id)
	fmt.Println(id.MovieID)
	query := bson.M{"movieid": id.MovieID}
	cursor, err := reviewcollection.Find(context.TODO(), query)
	if err != nil {
		log.Fatal(err)
	}
	var result []*Review
	for cursor.Next(context.TODO()) {
		var elem Review
		err := cursor.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		result = append(result, &elem)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	js, err2 := json.Marshal(result)
	fmt.Println(result)
	if err2 != nil {
		w.Write([]byte("No review"))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	fmt.Println("get Reivews")

}
