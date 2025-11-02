package main

import (
	"sync"
	"log"
    "context"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// product map that stores all products
var syncProducts sync.Map
// var products map[int]Item

// Response structure
type SearchResponse struct {
	Products      []Item `json:"products"`
	TotalFound    int    `json:"total_found"`
	TotalSearched int    `json:"total_searched"`
	SearchTime    string `json:"search_time"`
}


func main() {
	// Load .env file
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using system environment variables")
    }

	// Initialize DynamoDB connection
	log.Println("Initializing DynamoDB...")
	if err := InitDynamoDB(); err != nil {
		log.Fatalf("Failed to initialize DynamoDB: %v", err)
	}

	// Generate products
    log.Println("Generating products...")
    products := GenerateProducts(100000)
    
    // Check if products table is empty, only seed if needed
    ctx := context.Background()
    result, _ := dynamoClient.Scan(ctx, &dynamodb.ScanInput{
        TableName: aws.String(productsTable),
        Limit:     aws.Int32(1), // Just check if any product exists
    })
    
    if result == nil || len(result.Items) == 0 {
        log.Println("Products table empty, seeding...")
        if err := SeedData(products); err != nil {
            log.Printf("Warning: failed to seed data: %v", err)
        }
    } else {
        log.Println("Products already seeded, skipping...")
    }

	for k, v := range products {
		syncProducts.Store(k, v)
	}

	// initialize Gin router using Default
	router := gin.Default()

	// Health endpoint - checks DynamoDB connection
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":   "healthy",
			"database": "dynamodb",
		})
	})

	// Shopping cart endpoints
    router.POST("/shopping-carts", createShoppingCart)
    router.GET("/shopping-carts/:id", getShoppingCart)
    router.POST("/shopping-carts/:id/items", addItemToCart)
	// associate GET HTTP method and "/products/{productId}" path with a handler function "getItemByID"
	router.GET("/products/:productId", getItemByID)
	// associate POST HTTP method and "/products/{productId}/details" path with a handler function "postItem"
	router.POST("/products/:productId/details", postItem)
	// associate GET HTTP method and "/products/search?q={query}" path with a handler function "searchProducts"
	router.GET("/products/search", searchProducts)
	printSample(products, 10)
	log.Printf("Total products: %d", len(products))
	// "Run()" attaches router to an http server and start the server
	router.Run(":8080")
}
