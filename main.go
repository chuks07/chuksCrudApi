package main

import( 
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"os"
	"log"
	"time"
	"context"
)
type Book struct{
	Title string `json:"Title"`
	Author string `json:"Author"`
	Year int `json:"Year"`
	Category string `json:"Category"`
}
var Books []Book
var dbClient *mongo.Client 

func main(){
	fmt.Println("Hello world")

	ctx,cancel:=context.WithTimeout(context.Background(),10*time.Second)
	defer cancel()

	//find Ulr address
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err !=nil{
		log.Fatalf("Could not connect to the db: %v\n",err)
	}

	dbClient =client
	err = dbClient.Ping(ctx,readpref.Primary())
	if err != nil {
		log.Fatalf("mongo db not available: %v\n",err)
	}

	router := gin.Default()

	router.POST("/createnewBook", createNewBook)

	router.GET("/getBook/:Title", getSingleBook)

	router.GET("/getAllBooks", getAllBooks)

	router.PATCH("/updateABook/:Title", updateABook)

	router.DELETE("/deleteABook/:Title", deleteABook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	_= router.Run(":" + port)


}

func createNewBook (c *gin.Context){
	var book Book

	err := c.ShouldBindJSON(&book)
	if err !=nil{
		c.JSON(400,gin.H{
			"error": "invalid data request",
		})
		return
	}
	//create database and collection on mongodb compass later
	_, err = dbClient.Database("ChuksLibrary").Collection("ChuksBooks").InsertOne(context.Background(), book)

	if err !=nil {
		fmt.Println("error saving book", err)
		c.JSON(500, gin.H{
			"error": "Could not process request, coould not save new book",
		})
		return 
	} 
	Books = append(Books, book)

	c.JSON(200, gin.H{
		"message": "new book has sucessfully been created",
		"data": book,
	})
}
func getSingleBook( c *gin.Context) {
	Title := c.Param("Title")
	
	fmt.Println("Title", Title)

	var book Book

	bookAvailable :=false

	for _, value := range Books {

		if value.Title == Title{

			book = value

			bookAvailable = true
		}
	}
	if !bookAvailable {
		c.JSON(404, gin.H{
			"error":"no book with Title found:" + Title,
		})
		return
	}
	c.JSON(200, gin.H{
		"message":"success",
		"data": book,
	})
}
func getAllBooks (c *gin.Context) {
	
	c.JSON(200, gin.H{
		"message":"Welcome reader",
		"data": Books,
	})
}

func updateABook(c *gin.Context){

	Title := c.Param("Title")
	var book Book
	bookAvailable :=false

	for _, value := range Books {

		if value.Title == Title{

			book = value

			bookAvailable = true
		}
	}
	if !bookAvailable {
		c.JSON(404, gin.H{
			"error":"no book with Title found:" + Title,
		})
		return
	}

	err :=c.ShouldBindJSON(&book)

	

	filterQuery := bson.M{
		"Title": Title,
	}

	updateQuery :=bson.M{
		"$set": bson.M{
			"Title":book.Title,
			"Author":book.Author,
			"Category":book.Category,
			"Year":book.Year,
		},
	}
	// get  database name from mongodb compass
	_,err = dbClient.Database("ChuksLibrary").Collection("ChuksBooks").UpdateOne(context.Background(),filterQuery,updateQuery)
	if err !=nil {
		c.JSON(500, gin.H{
			"error":"could not process request, database has not been updated",
		})
		return 
	}
	c.JSON(200, gin.H{
		"message":"Book has been updated",
	})
}
func deleteABook (c *gin.Context){
	Title := c.Param("Title")

	query:= bson.M{
		"Title": Title,
	}
	_,err :=dbClient.Database("ChuksLibrary").Collection("Chuksbooks").DeleteOne(context.Background(),query)
	if err != nil{
		c.JSON(500, gin.H{
			"error":"could not process command, Book has not been deleted",
		})
		return
	}
	c.JSON(200,gin.H{
		"message":"Book has been deleted",
	})
}