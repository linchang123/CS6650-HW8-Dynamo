package main

import (
    "log"
    "net/http"
    "strconv"
    "time"
    "math/rand"
    "fmt"
    "strings"
    "context"
    "github.com/gin-gonic/gin"
    "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
    "github.com/aws/aws-sdk-go-v2/aws"
)

// CartItem represents an item in the shopping cart
type CartItemResponse struct {
    ID          int     `json:"id"`
    ProductID   int     `json:"product_id"`
    Manufacturer string  `json:"manufacturer"`
    Category     string	 `json:"category"`
    Quantity    int     `json:"quantity"`
    CreatedAt   string  `json:"created_at"`
    UpdatedAt   string  `json:"updated_at"`
}

// ShoppingCart represents a complete shopping cart
type ShoppingCartResponse struct {
    ID         int        `json:"id"`
    CustomerID int        `json:"customer_id"`
    Items      []CartItemResponse `json:"items"`
    CreatedAt  string     `json:"created_at"`
    UpdatedAt  string     `json:"updated_at"`
}

// createShoppingCart creates a new shopping cart
// POST /shopping-carts
func createShoppingCart(c *gin.Context) {
    var input struct {
        CustomerID int `json:"customer_id" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "customer_id is required",
        })
        return
    }
    
    // Try to get existing cart from DynamoDB
    ctx := context.Background()
    result, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
        TableName: aws.String(cartsTable),
        Key: map[string]types.AttributeValue{
            "customer_id": &types.AttributeValueMemberN{Value: strconv.Itoa(input.CustomerID)},
        },
    })
    
    // If cart exists in DynamoDB, return message
    if err == nil && result.Item != nil {
        c.JSON(http.StatusOK, gin.H{
            "message":     "Shopping cart already exists for this customer",
            "id":          input.CustomerID,
            "customer_id": input.CustomerID,
        })
        return
    }
    
    // Create and save new empty cart to DynamoDB
    newCart := &CartItem{
        CustomerID: input.CustomerID,
        Items:      []CartProduct{},
        CreatedAt:  time.Now().Format(time.RFC3339),
        UpdatedAt:  time.Now().Format(time.RFC3339),
    }
    
    // Marshal and save to DynamoDB
    item, err := attributevalue.MarshalMap(newCart)
    if err != nil {
        log.Printf("Error marshaling cart: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to create cart",
        })
        return
    }
    
    _, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
        TableName: aws.String(cartsTable),
        Item:      item,
    })
    if err != nil {
        log.Printf("Error saving cart to DynamoDB: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to create cart",
        })
        return
    }
    
    // Return the created cart
    c.JSON(http.StatusCreated, gin.H{
        "id":          input.CustomerID,
        "customer_id": input.CustomerID,
        "message":     fmt.Sprintf("shopping cart created for customer %d", input.CustomerID),
        "created_at":  newCart.CreatedAt,
    })
}

// getShoppingCart retrieves a shopping cart with all items by customer ID
// GET /shopping-carts/:id (where id is customer_id)
func getShoppingCart(c *gin.Context) {
    customerIDParam := c.Param("id")
    
    // Convert to integer
    customerID, err := strconv.Atoi(customerIDParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid customer ID",
        })
        return
    }
    
    // Get cart from DynamoDB
    cart, err := GetCart(customerID)
    if err != nil {
        log.Printf("Error retrieving cart: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Internal server error",
        })
        return
    }
    
    // Convert DynamoDB cart to response format
    response := ShoppingCartResponse{
        ID:         customerID, // Using customer_id as cart ID
        CustomerID: cart.CustomerID,
        CreatedAt:  cart.CreatedAt,
        UpdatedAt:  cart.UpdatedAt,
        Items:      []CartItemResponse{},
    }
    
    // Convert cart items to response format
    for i, item := range cart.Items {
        response.Items = append(response.Items, CartItemResponse{
            ID:           i + 1, // Generate sequential IDs for items
            ProductID:    item.ID,
            Manufacturer: item.Manufacturer, // Map name to manufacturer for compatibility
            Category:     item.Category, // Map description to category for compatibility
            Quantity:     item.Quantity,
            CreatedAt:    cart.CreatedAt,
            UpdatedAt:    cart.UpdatedAt,
        })
    }
    
    // Return the cart with all items
    c.JSON(http.StatusOK, response)
}

// addItemToCart adds or updates an item in the shopping cart by customer ID
// POST /shopping-carts/:id/items (where id is customer_id)
func addItemToCart(c *gin.Context) {
    customerIDParam := c.Param("id")
    
    // Convert to integer
    customerID, err := strconv.Atoi(customerIDParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid customer ID",
        })
        return
    }
    
    // Parse request body
    var input struct {
        ProductID int `json:"product_id" binding:"required"`
        Quantity  int `json:"quantity" binding:"required,min=1"`
    }
    
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "product_id and quantity (min 1) are required",
        })
        return
    }
    
    // Verify product exists in DynamoDB
    product, err := GetProduct(input.ProductID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Product not found",
        })
        return
    }
    
    // Add item to cart using DynamoDB function
    err = AddToCart(customerID, input.ProductID, input.Quantity)
    if err != nil {
        log.Printf("Error adding item to cart: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to add item to cart",
        })
        return
    }
    
    // Get updated cart to return
    cart, err := GetCart(customerID)
    if err != nil {
        log.Printf("Error retrieving updated cart: %v", err)
        c.JSON(http.StatusOK, gin.H{
            "message":    "Item added to cart",
            "product_id": input.ProductID,
            "quantity":   input.Quantity,
        })
        return
    }
    
    // Find the added/updated item in the cart
    var addedItem CartItemResponse
    for i, item := range cart.Items {
        if item.ID == input.ProductID {
            addedItem = CartItemResponse{
                ID:           i + 1,
                ProductID:    item.ID,
                Manufacturer: product.Manufacturer,
                Category:     product.Category,
                Quantity:     item.Quantity,
                CreatedAt:    cart.CreatedAt,
                UpdatedAt:    cart.UpdatedAt,
            }
            break
        }
    }
    
    c.JSON(http.StatusOK, gin.H{
        "message": "Item added to cart successfully",
        "item":    addedItem,
    })
}

func searchProducts(c *gin.Context) {
    defer func() {
        if r := recover(); r != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "INTERNAL_SERVER_ERROR",
                "message": "something went wrong",
                "details": fmt.Sprintf("%v", r),
            })
        }
    }()
    startTime := time.Now()

    // Extract query parameter
    query := c.Query("q")
    if query == "" {
        c.JSON(400, gin.H{"error": "Query parameter 'q' is required"})
        return
    }
    // Convert query to lowercase for case-insensitive search
    queryLower := strings.ToLower(query)

    // Generate 100 random product IDs (1-100000)
    randomIDs := generateRandomIDs(100, 1, 100000)

    // Search for matching products
    var matchingProducts []Item
    totalFound := 0
    totalSearched := 0

    for _, productID := range randomIDs {
        // Check if product exists in map
        totalSearched++
        if value, exists := syncProducts.Load(productID); exists {
            // Check if query matches name, category, or brand (case-insensitive)
            item := value.(Item)
            nameLower := strings.ToLower(item.Name)
            categoryLower := strings.ToLower(item.Category)
            brandLower := strings.ToLower(item.Brand)

            if strings.Contains(nameLower, queryLower) ||
                strings.Contains(categoryLower, queryLower) ||
                strings.Contains(brandLower, queryLower) {

                totalFound++

                // Add to results if we haven't reached 20 items yet
                if len(matchingProducts) < 20 {
                    matchingProducts = append(matchingProducts, item)
                }
            }
        }
    }

    // Calculate search duration
    duration := time.Since(startTime)
    searchTime := fmt.Sprintf("%.3fs", duration.Seconds())

    // Create response
    response := SearchResponse{
        Products:      matchingProducts,
        TotalFound:    totalFound,
        TotalSearched: totalSearched,
        SearchTime:    searchTime,
    }

    // Return empty array instead of null if no products found
    if response.Products == nil {
        response.Products = []Item{}
    }

    c.JSON(200, response)
}

// generateRandomIDs generates n random integers between min and max (inclusive)
func generateRandomIDs(n, min, max int) []int {
    ids := make([]int, n)
    for i := 0; i < n; i++ {
        ids[i] = rand.Intn(max-min+1) + min
    }
    return ids
}

// postAlbums adds an album from JSON received in the request body.
func postItem(c *gin.Context) {

    defer func() {
        if r := recover(); r != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "INTERNAL_SERVER_ERROR",
                "message": "something went wrong",
                "details": fmt.Sprintf("%v", r),
            })
        }
    }()

    // Extract product ID from route
    productIDStr := c.Param("productId")
    productID, err := strconv.Atoi(productIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "INVALID_INPUT",
            "message": "data input invalid",
            "details": "invalid productId",
        })
        return
    }

    // Check if product exists in map
    _, exists := syncProducts.Load(productID)
    if !exists {
        c.JSON(http.StatusNotFound, gin.H{
            "error":   "NOT_FOUND",
            "message": "product not found",
            "details": fmt.Sprintf("no item with ID %d", productID),
        })
        return
    }

    // Call BindJSON to bind the received JSON (from request body) to
    // newItem.
    // if err := c.BindJSON(&newItem); err != nil {
    // 	return
    // }
    var newDetails Item
    if err := c.ShouldBindJSON(&newDetails); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "INVALID_INPUT",
            "message": "The provided input data is invalid",
            "details": err.Error(), // tells why decoding failed
        })
        return
    }

    // Ensure the product ID in body matches the route parameter
    if newDetails.ID != productID {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "INVALID_INPUT",
            "message": "data input invalid",
            "details": "product_id in body does not match route parameter",
        })
        return
    }

    // Add the new details to the corresponding product.
    syncProducts.Store(productID, newDetails)

    c.Status(http.StatusNoContent)
}

// getItemByID locates the item whose ID value matches the productId
// parameter sent by the client, then returns that item as a response.
func getItemByID(c *gin.Context) {

    defer func() {
        if r := recover(); r != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "INTERNAL_SERVER_ERROR",
                "message": "something went wrong",
                "details": fmt.Sprintf("%v", r),
            })
        }
    }()

    // id := c.Param("productId") // "Context.Param()" retrieves the productId path parameter from the URL

    // Extract product ID from route
    productIDStr := c.Param("productId")
    productID, err := strconv.Atoi(productIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "INVALID_INPUT",
            "message": "data input invalid",
            "details": "invalid productID",
        })
        return
    }
    // Check if product exists in map
    value, exists := syncProducts.Load(productID)
    if !exists {
        c.JSON(http.StatusNotFound, gin.H{
            "error":   "INVALID_INPUT",
            "message": "product not found",
            "details": fmt.Sprintf("no item with ID %d", productID),
        })
        return
    }

    // return "404 not found error" if the album is not found
    c.IndentedJSON(http.StatusOK, value.(Item))

}