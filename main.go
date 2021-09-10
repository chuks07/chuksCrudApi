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
	Title string `json:"title"`
	Author string `json:"author"`
	Year int `json:"year"`
	Category string `json:"category"`
}
var Books []Book
var dbClient *mongo.Client 

func main(){
	fmt.Println("Hello world")

	ctx, cancel:= context.WithTimeout(context.Background(),10*time.Second)
	defer cancel()

	//find Ulr address
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err !=nil{
		log.Fatalf("Could not connect to the db: %v\n",err)
	}

	dbClient = client
	err = dbClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("mongo db not available: %v\n", err)
	}

	router := gin.Default()

	router.POST("/createnewBook", createNewBook)

	router.GET("/getBook/:title", getSingleBook)

	router.GET("/getAllBooks", getAllBooks)

	router.PATCH("/updateABook/:title", updateABook)

	router.DELETE("/deleteABook/:title", deleteABook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	_ = router.Run(":" + port)


}

func createNewBook (c *gin.Context){
	var book Book

	err := c.ShouldBindJSON(&book)
	if err != nil{
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

	c.JSON(200, gin.H{
		"message": "new book has sucessfully been created",
		"data": book,
	})
}
func getSingleBook( c *gin.Context) {
	title := c.Param("title")
	
	var book Book

	query := bson.M{
		"title" :title,
	}

	err :=dbClient.Database("ChuksLibrary").Collection("ChuksBooks").FindOne(context.Background(), query).Decode(&book)

	if err != nil {
		fmt.Println("book not found", err)
		c.JSON(400, gin.H{
			"error" : "no book with title found :" +title,
		})
		return
	}
	
	c.JSON(200, gin.H{
		"message":"success",
		"data": book,
	})
}
func getAllBooks (c *gin.Context) {

	var Books []Book

	cursor, err := dbClient.Database("ChuksLibrary").Collection("ChuksBooks").Find(context.Background(), bson.M{})

	if err != nil{
		c.JSON(500, gin.H{
			"error": "could not process request, could'nt get books",
		})
		return
	}

	err = cursor.All(context.Background(), &Books)
	if err != nil{
		c.JSON(500, gin.H{
			"error": "could not process request, could'nt get books",
		})
	}
	
	c.JSON(200, gin.H{
		"message":"Welcome reader",
		"data": Books,
	})
}

func updateABook(c *gin.Context){

	title := c.Param("title")
	var book Book

	err :=c.ShouldBindJSON(&book)

	if err != nil{
		c.JSON(400,gin.H{
			"error": "invalid request data",
		})
		return
	}	

	filterQuery := bson.M{
		"title": title,
	}

	updateQuery :=bson.M{
		"$set": bson.M{
			"title":book.Title,
			"author":book.Author,
			"category":book.Category,
			"year":book.Year,
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

// delete is still having errors

func deleteABook (c *gin.Context){
	title := c.Param("title")

	query:= bson.M{
		"title": title,
	}
	_, err :=dbClient.Database("ChuksLibrary").Collection("ChuksBooks").DeleteOne(context.Background(), query)
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