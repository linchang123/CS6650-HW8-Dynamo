package main

import (
	"fmt"
	"math/rand"
	"strings"
	// "time"
)


// item represents data about a product item.
// (item struct used to store product item data in memory)
// struct tag (e.g. `json:"artist"`) specify what a field's name
// should be when the struct's contents are serialized into JSON.
type Item struct {
	ID           int     `json:"product_id"`
	SKU          string  `json:"sku"`
	Manufacturer string  `json:"manufacturer"`
	CategoryID   int     `json:"category_id"`
	Weight       float64 `json:"weight"`
	SomeOtherID  int     `json:"some_other_id"`
	Name         string  `json:"name"`
	Category     string	 `json:"category"`
	Description  string  `json:"description"`
	Brand		 string  `json:"brand"`
}


func GenerateProducts(count int) map[int]Item {
	// rand.Seed(time.Now().UnixNano())
	
	products := make(map[int]Item)
	usedSKUs := make(map[string]bool)
	
	manufacturers := []string{
		"Muji", "Pilot", "Jans Sports", "Nike", "Adidas",
		"Apple", "Samsung", "Sony", "Dell", "HP",
		"Lenovo", "Asus", "Microsoft", "Amazon", "Google",
		"Patagonia", "North Face", "Columbia", "Under Armour", "Puma",
		"Reebok", "New Balance", "Vans", "Converse", "Timberland",
	}

	categories := []string{
    "Stationery",        // Muji
    "Pen",              // Pilot
    "Backpacks",         // Jans Sports (JanSport)
    "Athletic Apparel",  // Nike
    "Athletic Apparel",  // Adidas
    "Electronic",       // Apple
    "Electronic",       // Samsung
    "Electronic",       // Sony
    "Computer",         // Dell
    "Computer",         // HP
    "Computer",         // Lenovo
    "Computer",         // Asus
    "Software",          // Microsoft
    "E-commerce",        // Amazon
    "Technology",        // Google
    "Outdoor Apparel",   // Patagonia
    "Outdoor Apparel",   // North Face
    "Outdoor Apparel",   // Columbia
    "Athletic Apparel",  // Under Armour
    "Athletic Apparel",  // Puma
    "Athletic Apparel",  // Reebok
    "Athletic Footwear", // New Balance
    "Footwear",          // Vans
    "Footwear",          // Converse
    "Footwear",          // Timberland
}
	
	for i := 1; i <= count; i++ {
		// Generate unique SKU
		sku := GenerateUniqueSKU(usedSKUs)
		usedSKUs[sku] = true
		
		// Random manufacturer
		random_index := rand.Intn(len(manufacturers))
		manufacturer := manufacturers[random_index]
		
		// Random category ID (100-999)
		categoryID := rand.Intn(900) + 100
		category := categories[random_index]
		
		// Random weight (0.1 to 50.0)
		weight := rand.Float64()*49.9 + 0.1
		weight = float64(int(weight*10)) / 10 // Round to 1 decimal place
		
		// Random some other ID (100-9999)
		someOtherID := rand.Intn(9900) + 100
		name := fmt.Sprintf("Product %s %d", manufacturer, i)
		description := fmt.Sprintf("%s %s %d", manufacturer, category, i)
		
		item := Item{
			ID:           i,
			SKU:          sku,
			Manufacturer: manufacturer,
			CategoryID:   categoryID,
			Weight:       weight,
			SomeOtherID:  someOtherID,
			Name:		  name,
			Category:     category,
			Description:  description,
			Brand:        manufacturer,
		}
		
		products[i] = item
	}
	
	return products
}

func GenerateUniqueSKU(usedSKUs map[string]bool) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	
	for {
		// Generate first part (4 characters)
		part1 := make([]byte, 4)
		for i := 0; i < 4; i++ {
			part1[i] = letters[rand.Intn(len(letters))]
		}
		
		// Generate second part (3 characters)
		part2 := make([]byte, 3)
		for i := 0; i < 3; i++ {
			part2[i] = letters[rand.Intn(len(letters))]
		}
		
		sku := string(part1) + "-" + string(part2)
		
		// Check if SKU is unique
		if !usedSKUs[sku] {
			return sku
		}
	}
}

func printProducts(products map[int]Item) {
	fmt.Println("var products = map[int]Item{")
	
	// Get sorted keys for consistent output
	keys := make([]int, 0, len(products))
	for k := range products {
		keys = append(keys, k)
	}
	
	// Simple insertion sort for small to medium sized maps
	for i := 1; i < len(keys); i++ {
		key := keys[i]
		j := i - 1
		for j >= 0 && keys[j] > key {
			keys[j+1] = keys[j]
			j--
		}
		keys[j+1] = key
	}
	
	// Print each item
	for i, id := range keys {
		item := products[id]
		fmt.Printf("\t%d: {ID: %d, SKU: \"%s\", Manufacturer: \"%s\", CategoryID: %d, Weight: %.1f, SomeOtherID: %d}",
			id, item.ID, item.SKU, item.Manufacturer, item.CategoryID, item.Weight, item.SomeOtherID)
		
		if i < len(keys)-1 {
			fmt.Println(",")
		} else {
			fmt.Println(",")
		}
	}
	
	fmt.Println("}")
}

// Alternative: If you want to print just the first few items as a sample
func printSample(products map[int]Item, sampleSize int) {
	fmt.Printf("\n// Sample of first %d items:\n", sampleSize)
	fmt.Println("var productsSample = map[int]Item{")
	
	count := 0
	for i := 1; i <= len(products) && count < sampleSize; i++ {
		if item, exists := products[i]; exists {
			fmt.Printf("\t%d: {ID: %d, SKU: \"%s\", Manufacturer: \"%s\", CategoryID: %d, Weight: %.1f, SomeOtherID: %d,Name: \"%s\", Category: \"%s\", Description: \"%s\", Brand: \"%s\"}",
				i, item.ID, item.SKU, item.Manufacturer, item.CategoryID, item.Weight, item.SomeOtherID,
			item.Name, item.Category, item.Description, item.Brand)

			count++
			if count < sampleSize && i < len(products) {
				fmt.Println(",")
			} else {
				fmt.Println(",")
			}
		}
	}
	
	fmt.Println("}")
}

// Helper function to format the map as a string (useful for writing to file)
func FormatProductsAsString(products map[int]Item) string {
	var sb strings.Builder
	sb.WriteString("var products = map[int]Item{\n")
	
	// Get sorted keys
	keys := make([]int, 0, len(products))
	for k := range products {
		keys = append(keys, k)
	}
	
	// Sort keys
	for i := 1; i < len(keys); i++ {
		key := keys[i]
		j := i - 1
		for j >= 0 && keys[j] > key {
			keys[j+1] = keys[j]
			j--
		}
		keys[j+1] = key
	}
	
	for i, id := range keys {
		item := products[id]
		sb.WriteString(fmt.Sprintf("\t%d: {ID: %d, SKU: \"%s\", Manufacturer: \"%s\", CategoryID: %d, Weight: %.1f, SomeOtherID: %d}",
			id, item.ID, item.SKU, item.Manufacturer, item.CategoryID, item.Weight, item.SomeOtherID))
		
		if i < len(keys)-1 {
			sb.WriteString(",\n")
		} else {
			sb.WriteString(",\n")
		}
	}
	
	sb.WriteString("}")
	return sb.String()
}

