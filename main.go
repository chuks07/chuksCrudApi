package main

import( 
	"fmt"
	"github.com/gin-gonic/gin"
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

func main(){
	fmt.Println("Hello world")
}

func createBookHandler (c *gin.Context){
	var book Book

	err := c.ShouldBindJSON(&book)
	if err !=nil{
		c.JSON(400,gin.H{
			"error": "invalid data request",
		})
	}
}