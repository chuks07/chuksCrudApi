package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
)

const (
	DbName ="ChuksLibrary"
	TaskCollection = "library"
	BookCollection = "books"
	jwtSecret ="secretname"

)
type Book struct{
	Title  		string `json:"title" bson:"title"`
	Author 		string `json:"author" bson:"author"`
	Email  		string `json:"email" bson:"email"`
	Password 	string `json:"-,oimtempty" bson:"password"`
	Ts   		time.Time `json:"timestamp" bson:"timestamp"`

}

type Task struct {
	ID 				string 		`json:"id"`
	Owner			string 		`json:"owner"`
	Name 			string 		`json:"name"`
	Description		string 		`json:"description"`
	Ts				time.Time 	`json:"timestamp"`
}

type Claim struct {
	BookId string `json:"book_id"`
	jwt.StandardClaims 
}


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

	router.POST("/createTask", createTaskHandler)

	router.GET("/getTask/:title", getSingleBook)

	router.GET("/getAllBooks", getAllBooks)

	router.GET("/getTasks", getAllTasksHandler)

	router.PATCH("/updateABook/:title", updateABook)

	router.DELETE("/deleteABook/:title", deleteABook)

	router.POST("/login", loginHandler)

	router.POST("/signup", signUpHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	_ = router.Run(":" + port)


}
func welcomeHandler (c *gin.Context) {
	c.JSON(200, gin.H{
		"message":"welcome to task manger API",
	})
}

func createTaskHandler (c *gin.Context){
	authorization :=c.Request.Header.Get("Authorization")
	fmt.Println(authorization)

	jwtToken:= ""

			sp := strings.Split(authorization, " ")
			if len(sp) > 1{
				jwtToken = sp[1]
			}

	claims	:= &Claim{}
	keyFunc := func (token *jwt.Token)(i interface{}, e error)  {
			return [] byte(jwtSecret), nil
	}

	token, err := jwt.ParseWithClaims(jwtToken,claims, keyFunc)
	if !token.Valid {
		c.JSON(400, gin.H{
			"error" : "invalid jwt token",
		})
		return 
	}
	var taskReq Task


	err = c.ShouldBindJSON(&taskReq)
	if err != nil{
		c.JSON(400,gin.H{
			"error": "invalid data request",
		})
		return
	}
	
	taskId := uuid.NewV4().String()

	task := Task{
		ID: taskId,
		Owner: claims.BookId,
		Name: taskReq.Name,
		Description: taskReq.Description,
		Ts: time.Now(),
	}
		_, err = dbClient.Database(DbName).Collection(TaskCollection).InsertOne(context.Background(), task)

	if err !=nil {
		fmt.Println("error saving task", err)
		c.JSON(500, gin.H{
			"error": "Could not process request, coould not save new task",
		})
		return 
	} 

	c.JSON(200, gin.H{
		"message": "new book has sucessfully been created",
		"data": task,
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

	cursor, err := dbClient.Database(DbName).Collection("ChuksBooks").Find(context.Background(), bson.M{})

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
func  getAllTasksHandler(c *gin.Context){
	authorization := c.Request.Header.Get("Authorization")
	if authorization == "" {
		c.JSON(401, gin.H{
			"error" : "auth token required",
		})
		return
	}
	jwtToken := ""
	sp := strings.Split(authorization, " ")
	if len(sp) > 1{
		jwtToken = sp[1]
	}
	claims := &Claim{}
	keyFunc := func(token *jwt.Token) (i interface{}, e error) {
			return []byte(jwtSecret), nil 
	}

	token, err := jwt.ParseWithClaims(jwtToken, claims, keyFunc)
	if !token.Valid {
		c.JSON(401, gin.H{
			"error": "invalid jwt token",
		})
		return 
	}
	var tasks []Task
	query := bson.M{
		"owner": claims.BookId,
	}

	cursor, err :=dbClient.Database(DbName).Collection(TaskCollection).Find(context.Background(), query)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Could not process request, could get task",
		})
		return 
	}
	err = cursor.All(context.Background(), &tasks)
	if err != nil{
		c.JSON(500, gin.H{
			"error": "Could not process request, could get task",
		})
		return 
	}
	
	c.JSON(200, gin.H{
		"message": "success",
		"data": tasks,
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
			
		},
	}
	// get  database name from mongodb compass
	_,err = dbClient.Database(DbName).Collection("ChuksBooks").UpdateOne(context.Background(),filterQuery,updateQuery)
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
	_, err :=dbClient.Database(DbName).Collection("ChuksBooks").DeleteOne(context.Background(), query)
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
func loginHandler (c *gin.Context) {
	loginReq := struct {
		Email string `json:"email"`
		Password string `json:"password"`
	} {}

	err := c.ShouldBindJSON(&loginReq)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "invalid request data",
		})
		return 
	}

	var book Book
	query := bson.M{ 
		"email": loginReq.Email,
	}

	err = dbClient.Database(DbName).Collection("ChuksLibrary").FindOne(context.Background(),query).Decode(&book)
	if err !=nil {
		fmt.Printf("error getting book from db: %v\n", err)
		c.JSON(500, gin.H{
			"error":"could not process request, could not get book",
		})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(book.Password),[]byte(loginReq.Password))
	if err !=nil {
		fmt.Printf("error getting book from db: %v\n", err)
		c.JSON(500, gin.H{
			"error":"invalid login details",
		})
		return
	}
	claims :=&Claim{
		BookId: book.Title,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
		},
	}

	token :=jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtTokenString, err := token.SignedString([]byte(jwtSecret))

	c.JSON(200, gin.H{
		"message": "sign up successful",
		"token": jwtTokenString,
		"data":    book,
	})
}

func signUpHandler (c *gin.Context){

	type SignupRequest struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var signupReq SignupRequest

	err := c.ShouldBindJSON(&signupReq)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "invalid request data",
		})
		return
	}

	query := bson.M {
		"email": signupReq.Email,
	}
	count, err := dbClient.Database(DbName).Collection("ChuksLibrary").CountDocuments(context.Background(), query)

	if err != nil {
		fmt.Println("error searching for book: ", err)

		c.JSON(500, gin.H{
			"error": "Could not process request, please try again later",
		})
		return
	}

	if count > 0 {
		c.JSON(500, gin.H{
			"error": "Email already exits, please use a different email",
		})
		return
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(signupReq.Password), bcrypt.DefaultCost)
	hashPassword := string(bytes)

	bookTitle := uuid.NewV4().String()

	book := Book {
		Title: bookTitle,
		Author: signupReq.Name,
		Email:signupReq.Email,
		Password:hashPassword,
		Ts: time.Now(),
	}
	_, err = dbClient.Database(DbName).Collection("ChuksLibrary").InsertOne(context.Background(), book)
	if err != nil {
		fmt.Println("error saving book", err)

		c.JSON(500, gin.H{
			"error": "Could not process request, could not save book",
		})
		return
	}
	claims :=&Claim{
		BookId: book.Title,
		StandardClaims: jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour *1).Unix(),
		},
	}
	
	token:= jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtTokenString, err:= token.SignedString([]byte(jwtSecret))

	c.JSON(200, gin.H{
		"message": "sign up successful",
		"token": jwtTokenString,
		"data":    book,
	})
}