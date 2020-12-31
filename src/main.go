package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/go-cmp/cmp"
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
	Phone     string
}

type CheckUser struct {
	UserName string
	Password string
}

type UpdateUser struct {
	FirstName string
	LastName  string
	UserName  string
	Email     string
	Phone     string
}

type Message struct {
	State   bool
	Message string
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
var loginCollection = database.Collection("login")
var reviewCollection = database.Collection("review")

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
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/updateUser", updateUser)
	http.HandleFunc("/writereview", insertreview)
	http.HandleFunc("/getreview", getreview)
	http.HandleFunc("/updatereview", updatereview)
	http.HandleFunc("/deletereview", deletereview)
	http.HandleFunc("/confirmPassword", confirmPassword)
	http.HandleFunc("/changePassword", changePassword)
	http.Handle("/", http.FileServer(http.Dir("./build")))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./build/static"))))

	http.ListenAndServe(":8000", nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	// var result User
	decoder := json.NewDecoder(r.Body)
	var mem CheckUser
	var result User
	err := decoder.Decode(&mem)

	if err != nil {
		panic(err)
	}
	fmt.Println(mem.UserName)
	query := bson.M{"username": mem.UserName, "password": mem.Password}

	err = loginCollection.FindOne(context.TODO(), query).Decode(&result)
	if err != nil {
		var msg = Message{false, "You type wrong userid or Password. Please try again"}
		js, err := json.Marshal(msg)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)

		return

	}
	js, err2 := json.Marshal(result)
	if err2 != nil {
		w.Write([]byte("fail"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func confirmPassword(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var mem, check CheckUser

	err := decoder.Decode(&mem)
	if err != nil {
		panic(err)
	}
	query := bson.M{"username": mem.UserName, "password": mem.Password}
	err = loginCollection.FindOne(context.TODO(), query).Decode(&check)
	var msg Message
	if err != nil {
		msg = Message{false, "Password is not correct"}
	} else {
		msg = Message{true, " "}

	}
	js, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func changePassword(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var mem CheckUser

	err := decoder.Decode(&mem)
	if err != nil {
		panic(err)
	}
	query := bson.M{"username": mem.UserName}
	_, err = loginCollection.UpdateOne(context.TODO(), query, bson.D{{"$set", mem}})
	var msg Message
	if err != nil {
		msg = Message{false, "Fail to change the password. Try again"}
	} else {
		msg = Message{true, "Password is changed successfully!"}
	}
	js, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func signup(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	fmt.Println(decoder)

	var user, empty, search User
	err := decoder.Decode(&user)
	if err != nil {
		panic(err)
	}

	query := bson.M{"username": user.UserName}
	loginCollection.FindOne(context.TODO(), query).Decode(&search)
	fmt.Println(cmp.Equal(search, empty))

	if cmp.Equal(search, empty) {
		insertResult, err := loginCollection.InsertOne(context.TODO(), user)
		if err != nil {

			log.Fatal(err)
		}

		fmt.Println("Inserted a single document: ", insertResult.InsertedID)
		var message Message = Message{true, "Your sign up is succeed!"}
		packet, _ := json.Marshal(message)
		w.Write(packet)
		os.Mkdir("./Image/"+user.UserName, os.ModePerm)
		noImage, err := os.Open("./Image/noImage.jpg")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer noImage.Close()
		userImage, err := os.OpenFile("./Image/"+user.UserName+"/"+user.UserName+".jpg", os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer noImage.Close()
		io.Copy(userImage, noImage)

	} else {
		var message Message = Message{false, "The userid is existed already. Please use another userid"}
		packet, _ := json.Marshal(message)
		w.Write(packet)

	}
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	update := UpdateUser{
		r.Form["firstName"][0],
		r.Form["lastName"][0],
		r.Form["username"][0],
		r.Form["email"][0],
		r.Form["phone"][0],
	}
	query := bson.M{"username": update.UserName}
	fmt.Println(update.UserName)
	result, err := loginCollection.UpdateOne(context.TODO(), query, bson.D{
		{"$set", update}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Replaced %v Documents!\n", result.ModifiedCount)

	var user UpdateUser
	err = loginCollection.FindOne(context.TODO(), query).Decode(&user)
	if err != nil {
		w.Write([]byte("fail"))

	}
	js, err2 := json.Marshal(user)
	if err2 != nil {
		w.Write([]byte("fail"))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	fmt.Println("Update User and return the data")
	file, handler, err := r.FormFile("userImage")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()

	f, err := os.OpenFile("./Image/"+update.UserName+"/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer f.Close()
	io.Copy(f, file)

}

func insertreview(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	currenttime := time.Now()
	var review Review
	err := decoder.Decode(&review)
	review.Date = currenttime.Format("2006-01-02")
	review.ID = primitive.NewObjectID()
	fmt.Println(review.ID)

	insertResult, err := reviewCollection.InsertOne(context.TODO(), review)
	if err != nil {

		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", insertResult.InsertedID)

	query := bson.M{"movieid": review.MovieID}
	cursor, err := reviewCollection.Find(context.TODO(), query)
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
	res, err := reviewCollection.DeleteOne(context.TODO(), bson.M{"_id": docid})
	if err != nil {
		fmt.Println("This is Delete result error")
	}
	if res.DeletedCount == 0 {
		fmt.Println("Delete document is not found", res)
	} else {
		fmt.Println("DeleteOne Result:", res)
	}

	query := bson.M{"movieid": elem["movieid"]}
	cursor, err := reviewCollection.Find(context.TODO(), query)
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
	result, err := reviewCollection.ReplaceOne(context.TODO(), query, update)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Replaced %v Documents!\n", result.ModifiedCount)

	query2 := bson.M{"movieid": update.MovieID}
	cursor, err := reviewCollection.Find(context.TODO(), query2)
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
	cursor, err := reviewCollection.Find(context.TODO(), query)
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
