package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	// CRITICAL: These values must match Part I (MySQL test)
	NumCreateCart = 50
	NumAddItems   = 50
	NumGetCart    = 50
	TotalOps      = NumCreateCart + NumAddItems + NumGetCart // 150 total
	
	NumWorkers    = 10 // Concurrent workers
	TimeLimit     = 5 * time.Minute
)

// TestResult represents a single operation result
type TestResult struct {
	Operation    string  `json:"operation"`
	ResponseTime float64 `json:"response_time"` // in milliseconds
	Success      bool    `json:"success"`
	StatusCode   int     `json:"status_code"`
	Timestamp    string  `json:"timestamp"`
	CustomerID   int     `json:"customer_id,omitempty"`
}

// TestOutput represents the complete test output
type TestOutput struct {
	Results    []TestResult       `json:"results"`
	Statistics map[string]OpStats `json:"statistics"`
}

// OpStats represents statistics for an operation type
type OpStats struct {
	Count             int     `json:"count"`
	Successful        int     `json:"successful"`
	Failed            int     `json:"failed"`
	AvgResponseTime   float64 `json:"avg_response_time"`
	MinResponseTime   float64 `json:"min_response_time"`
	MaxResponseTime   float64 `json:"max_response_time"`
	TotalResponseTime float64 `json:"total_response_time"`
}

var (
	baseURL        string
	results        []TestResult
	resultsMutex   sync.Mutex
	httpClient     *http.Client
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run dynamodb_test_concurrent.go <ALB_URL>")
		fmt.Println("Example: go run dynamodb_test_concurrent.go http://your-alb.amazonaws.com")
		os.Exit(1)
	}

	baseURL = os.Args[1]
	httpClient = &http.Client{Timeout: 30 * time.Second}

	printHeader()

	// Test connectivity
	if !testConnectivity() {
		fmt.Println("✗ Service is not accessible")
		os.Exit(1)
	}
	fmt.Println("✓ Service is healthy")

	// Generate unique customer IDs
	baseCustomerID := rand.Intn(100000) + 10000
	fmt.Printf("Using customer IDs: %d - %d\n\n", baseCustomerID, baseCustomerID+NumCreateCart-1)

	startTime := time.Now()

	// Phase 1: Create carts concurrently
	fmt.Println("Phase 1: Creating shopping carts concurrently...")
	customerIDs := make([]int, NumCreateCart)
	for i := 0; i < NumCreateCart; i++ {
		customerIDs[i] = baseCustomerID + i
	}
	
	runConcurrent(NumCreateCart, func(i int) {
		createCart(customerIDs[i])
	})
	fmt.Println("✓ Phase 1 complete")

	// Phase 2: Add items concurrently
	fmt.Println("Phase 2: Adding items to carts concurrently...")
	runConcurrent(NumAddItems, func(i int) {
		customerID := customerIDs[i%len(customerIDs)]
		addItemToCart(customerID)
	})
	fmt.Println("✓ Phase 2 complete")

	// Phase 3: Get carts concurrently
	fmt.Println("Phase 3: Retrieving carts concurrently...")
	runConcurrent(NumGetCart, func(i int) {
		customerID := customerIDs[i%len(customerIDs)]
		getCart(customerID)
	})
	fmt.Println("✓ Phase 3 complete")

	duration := time.Since(startTime)

	// Calculate statistics
	stats := calculateStatistics()

	// Create output
	output := TestOutput{
		Results:    results,
		Statistics: stats,
	}

	// Save to JSON
	saveResults(output, "dynamodb_test_results.json")

	// Print summary
	printSummary(duration, stats)

	// Check time limit
	if duration > TimeLimit {
		fmt.Printf("⚠ WARNING: Test took longer than 5 minutes (%.2fs)\n", duration.Seconds())
	} else {
		fmt.Println("✓ Test completed within 5 minutes")
	}

	// Check success rate
	successRate := float64(countSuccessful()) / float64(len(results)) * 100
	if successRate == 100 {
		fmt.Println("✓ All operations successful")
	} else {
		fmt.Printf("⚠ Success rate: %.2f%%\n", successRate)
	}

	fmt.Println("============================================================")
}

func printHeader() {
	fmt.Println("============================================================")
	fmt.Println("Concurrent DynamoDB Shopping Cart Test (Go)")
	fmt.Println("============================================================")
	fmt.Printf("Target: %s\n", baseURL)
	fmt.Printf("Concurrent Workers: %d\n", NumWorkers)
	fmt.Printf("Total Operations: %d\n", TotalOps)
	fmt.Printf("  - Create Cart: %d\n", NumCreateCart)
	fmt.Printf("  - Add Items: %d\n", NumAddItems)
	fmt.Printf("  - Get Cart: %d\n", NumGetCart)
	fmt.Println("Output: dynamodb_test_results.json")
	fmt.Println("============================================================")
}

func testConnectivity() bool {
	resp, err := httpClient.Get(baseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func runConcurrent(count int, taskFunc func(int)) []int {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, NumWorkers)
	customerIDs := make([]int, count)

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			taskFunc(index)
			<-semaphore                    // Release
		}(i)
	}

	wg.Wait()
	return customerIDs
}

func createCart(customerID int) {
	startTime := time.Now()

	payload := map[string]int{"customer_id": customerID}
	jsonData, _ := json.Marshal(payload)

	resp, err := httpClient.Post(
		baseURL+"/shopping-carts",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	duration := time.Since(startTime).Seconds() * 1000 // Convert to milliseconds

	result := TestResult{
		Operation:    "create_cart",
		ResponseTime: duration,
		Success:      err == nil && (resp.StatusCode == 200 || resp.StatusCode == 201),
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		CustomerID:   customerID,
	}

	if resp != nil {
		result.StatusCode = resp.StatusCode
		resp.Body.Close()
	}

	addResult(result)
}

func addItemToCart(customerID int) {
	startTime := time.Now()

	productID := rand.Intn(100000) + 1
	quantity := rand.Intn(5) + 1

	payload := map[string]int{
		"product_id": productID,
		"quantity":   quantity,
	}
	jsonData, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/shopping-carts/%d/items", baseURL, customerID)
	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))

	duration := time.Since(startTime).Seconds() * 1000

	result := TestResult{
		Operation:    "add_items",
		ResponseTime: duration,
		Success:      err == nil && (resp.StatusCode == 200 || resp.StatusCode == 201),
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		CustomerID:   customerID,
	}

	if resp != nil {
		result.StatusCode = resp.StatusCode
		resp.Body.Close()
	}

	addResult(result)
}

func getCart(customerID int) {
	startTime := time.Now()

	url := fmt.Sprintf("%s/shopping-carts/%d", baseURL, customerID)
	resp, err := httpClient.Get(url)

	duration := time.Since(startTime).Seconds() * 1000

	result := TestResult{
		Operation:    "get_cart",
		ResponseTime: duration,
		Success:      err == nil && resp.StatusCode == 200,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		CustomerID:   customerID,
	}

	if resp != nil {
		result.StatusCode = resp.StatusCode
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	addResult(result)
}

func addResult(result TestResult) {
	resultsMutex.Lock()
	defer resultsMutex.Unlock()
	results = append(results, result)
}

func calculateStatistics() map[string]OpStats {
	stats := make(map[string]OpStats)
	opTypes := []string{"create_cart", "add_items", "get_cart"}

	for _, opType := range opTypes {
		stat := OpStats{
			MinResponseTime: 999999,
		}

		for _, result := range results {
			if result.Operation == opType {
				stat.Count++
				stat.TotalResponseTime += result.ResponseTime

				if result.Success {
					stat.Successful++
				} else {
					stat.Failed++
				}

				if result.ResponseTime < stat.MinResponseTime {
					stat.MinResponseTime = result.ResponseTime
				}
				if result.ResponseTime > stat.MaxResponseTime {
					stat.MaxResponseTime = result.ResponseTime
				}
			}
		}

		if stat.Count > 0 {
			stat.AvgResponseTime = stat.TotalResponseTime / float64(stat.Count)
		}

		stats[opType] = stat
	}

	return stats
}

func saveResults(output TestOutput, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fmt.Printf("Error encoding JSON: %v\n", err)
		return
	}

	fmt.Printf("\nResults saved to: %s\n", filename)
}

func printSummary(duration time.Duration, stats map[string]OpStats) {
	fmt.Println("============================================================")
	fmt.Println("TEST SUMMARY")
	fmt.Println("============================================================")
	fmt.Printf("Total Duration: %.2f seconds\n", duration.Seconds())
	fmt.Printf("Total Operations: %d\n", len(results))
	fmt.Printf("Successful: %d\n", countSuccessful())
	fmt.Printf("Failed: %d\n", len(results)-countSuccessful())
	fmt.Printf("Success Rate: %.2f%%\n\n", float64(countSuccessful())/float64(len(results))*100)

	for opType, stat := range stats {
		fmt.Printf("%s:\n", opType)
		fmt.Printf("  Count: %d\n", stat.Count)
		fmt.Printf("  Success: %d/%d\n", stat.Successful, stat.Count)
		fmt.Printf("  Avg Response Time: %.2f ms\n", stat.AvgResponseTime)
		fmt.Printf("  Min/Max: %.2f/%.2f ms\n\n", stat.MinResponseTime, stat.MaxResponseTime)
	}
}

func countSuccessful() int {
	count := 0
	for _, result := range results {
		if result.Success {
			count++
		}
	}
	return count
}