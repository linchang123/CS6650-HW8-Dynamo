package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	dynamoClient    *dynamodb.Client
	productsTable   string
	cartsTable      string
)

type ProductItem struct {
	ID           int     `dynamodbav:"product_id"`
	SKU          string  `dynamodbav:"sku"`
	Manufacturer string  `dynamodbav:"manufacturer"`
	CategoryID   int     `dynamodbav:"category_id"`
	Weight       float64 `dynamodbav:"weight"`
	SomeOtherID  int     `dynamodbav:"some_other_id"`
	Name         string  `dynamodbav:"name"`
	Category     string  `dynamodbav:"category"`
	Description  string  `dynamodbav:"description"`
	Brand        string  `dynamodbav:"brand"`
}


type CartItem struct {
	CustomerID int           `dynamodbav:"customer_id"`
	Items      []CartProduct `dynamodbav:"items"`
	CreatedAt  string        `dynamodbav:"created_at"`
	UpdatedAt  string        `dynamodbav:"updated_at"`
}

type CartProduct struct {
	ID           int     `dynamodbav:"product_id"`
	// SKU          string  `dynamodbav:"sku"`
	Manufacturer string  `dynamodbav:"manufacturer"`
	// CategoryID   int     `dynamodbav:"category_id"`
	// Weight       float64 `dynamodbav:"weight"`
	// SomeOtherID  int     `dynamodbav:"some_other_id"`
	// Name         string  `dynamodbav:"name"`
	Category     string  `dynamodbav:"category"`
	// Description  string  `dynamodbav:"description"`
	// Brand        string  `dynamodbav:"brand"`
	Quantity     int     `dynamodbav:"quantity"`
}

type CustomerItem struct {
	CustomerID int    `dynamodbav:"customer_id"`
	Name       string `dynamodbav:"name"`
	Email      string `dynamodbav:"email"`
	CreatedAt  string `dynamodbav:"created_at"`
}

// InitDynamoDB initializes the DynamoDB client and table names
func InitDynamoDB() error {
	ctx := context.Background()

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		return fmt.Errorf("unable to load SDK config: %v", err)
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)

	// Get table names from environment
	productsTable = os.Getenv("PRODUCTS_TABLE")
	cartsTable = os.Getenv("CARTS_TABLE")

	if productsTable == "" || cartsTable == "" {
		return fmt.Errorf("table names not set in environment variables")
	}

	log.Printf("DynamoDB initialized with tables: %s, %s", 
		productsTable, cartsTable)

	return nil
}

// GetProduct retrieves a product by ID
func GetProduct(productID int) (*ProductItem, error) {
	ctx := context.Background()

	result, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(productsTable),
		Key: map[string]types.AttributeValue{
			"product_id": &types.AttributeValueMemberN{Value: strconv.Itoa(productID)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %v", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("product not found")
	}

	var product ProductItem
	err = attributevalue.UnmarshalMap(result.Item, &product)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal product: %v", err)
	}

	return &product, nil
}


// GetCart retrieves a customer's cart
func GetCart(customerID int) (*CartItem, error) {
	ctx := context.Background()

	result, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(cartsTable),
		Key: map[string]types.AttributeValue{
			"customer_id": &types.AttributeValueMemberN{Value: strconv.Itoa(customerID)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %v", err)
	}

	if result.Item == nil {
		// Cart not found in DynamoDB - return error instead of empty cart
		return nil, fmt.Errorf("cart not found for customer %d", customerID)
	}

	var cart CartItem
	err = attributevalue.UnmarshalMap(result.Item, &cart)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart: %v", err)
	}

	return &cart, nil
}

// AddToCart adds a product to the customer's cart
func AddToCart(customerID, productID, quantity int) error {
	ctx := context.Background()

	// Get product details
	product, err := GetProduct(productID)
	if err != nil {
		return fmt.Errorf("product not found: %v", err)
	}

	// Get existing cart
	cart, err := GetCart(customerID)
	if err != nil {
		return fmt.Errorf("failed to get cart: %v", err)
	}

	// Check if product already in cart
	found := false
	for i, item := range cart.Items {
		if item.ID == productID {
			cart.Items[i].Quantity += quantity
			found = true
			break
		}
	}

	// Add new item if not found
	if !found {
		cart.Items = append(cart.Items, CartProduct{
			ID:           product.ID,
			// SKU:          product.SKU,
			Manufacturer: product.Manufacturer,
			// CategoryID:   product.CategoryID,
			// Weight:       product.Weight,
			// SomeOtherID:  product.SomeOtherID,
			// Name:         product.Name,
			Category:     product.Category,
			// Description:  product.Description,
			// Brand:        product.Brand,
			Quantity:     quantity,
		})
	}

	cart.UpdatedAt = time.Now().Format(time.RFC3339)

	// Marshal cart to DynamoDB format
	item, err := attributevalue.MarshalMap(cart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart: %v", err)
	}

	// Put cart back to DynamoDB
	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(cartsTable),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update cart: %v", err)
	}

	return nil
}

// SeedData populates DynamoDB with sample data using your existing GenerateProducts function
func SeedData(productsMap map[int]Item) error {
	ctx := context.Background()

	log.Println("Seeding DynamoDB tables...")

	
	log.Printf("Starting batch write to DynamoDB...")

	// Convert map to slice and batch write (max 25 items per batch)
	batchCount := 0
	writeRequests := make([]types.WriteRequest, 0, 25)
	
	for _, product := range productsMap {
		// Convert Item struct to DynamoDB ProductItem format (same structure, just with dynamodb tags)
		dynamoProduct := ProductItem{
			ID:           product.ID,
			SKU:          product.SKU,
			Manufacturer: product.Manufacturer,
			CategoryID:   product.CategoryID,
			Weight:       product.Weight,
			SomeOtherID:  product.SomeOtherID,
			Name:         product.Name,
			Category:     product.Category,
			Description:  product.Description,
			Brand:        product.Brand,
		}
		
		item, err := attributevalue.MarshalMap(dynamoProduct)
		if err != nil {
			log.Printf("Warning: failed to marshal product %d: %v", product.ID, err)
			continue
		}

		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: item,
			},
		})

		// When we have 25 items, write the batch
		if len(writeRequests) == 25 {
			_, err := dynamoClient.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]types.WriteRequest{
					productsTable: writeRequests,
				},
			})
			if err != nil {
				log.Printf("Warning: failed to batch write products: %v", err)
			}
			
			batchCount++
			if batchCount%100 == 0 {
				log.Printf("Seeded %d products...", batchCount*25)
			}
			
			// Reset for next batch
			writeRequests = make([]types.WriteRequest, 0, 25)
		}
	}
	
	// Write any remaining items
	if len(writeRequests) > 0 {
		_, err := dynamoClient.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				productsTable: writeRequests,
			},
		})
		if err != nil {
			log.Printf("Warning: failed to batch write final products: %v", err)
		}
		batchCount++
	}

	log.Printf("Database seeding completed! Seeded %d products in %d batches", len(productsMap), batchCount)
	return nil
}