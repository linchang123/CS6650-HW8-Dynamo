package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

// Configuration
const (
	BaseURL     = "http://cs6650l2-alb-1263136994.us-west-2.elb.amazonaws.com" // Update this!
	OutputFile  = "test_results.json"
	NumWorkers  = 10 // Number of concurrent workers
	NumCreateCart = 50
	NumAddItems   = 50
	NumGetCart    = 50
)

// TestResult represents a single test operation result
type TestResult struct {
	Operation    string  `json:"operation"`
	ResponseTime float64 `json:"response_time"`
	Success      bool    `json:"success"`
	StatusCode   int     `json:"status_code"`
	Timestamp    string  `json:"timestamp"`
	CustomerID   int     `json:"customer_id,omitempty"`
	ProductID    int     `json:"product_id,omitempty"`
	Quantity     int     `json:"quantity,omitempty"`
	Error        string  `json:"error,omitempty"`
}

// TestOutput represents the final JSON output
type TestOutput struct {
	TestMetadata TestMetadata `json:"test_metadata"`
	Statistics   Statistics   `json:"statistics"`
	Results      []TestResult `json:"results"`
}

type TestMetadata struct {
	BaseURL              string  `json:"base_url"`
	StartTime            string  `json:"start_time"`
	TotalDurationSeconds float64 `json:"total_duration_seconds"`
	TotalOperations      int     `json:"total_operations"`
	ConcurrentWorkers    int     `json:"concurrent_workers"`
}

type Statistics struct {
	TotalOperations      int                    `json:"total_operations"`
	SuccessfulOperations int                    `json:"successful_operations"`
	FailedOperations     int                    `json:"failed_operations"`
	SuccessRate          float64                `json:"success_rate"`
	Operations           map[string]OperationStats `json:"operations"`
}

type OperationStats struct {
	Count           int     `json:"count"`
	Successful      int     `json:"successful"`
	Failed          int     `json:"failed"`
	AvgResponseTime float64 `json:"avg_response_time"`
	MinResponseTime float64 `json:"min_response_time"`
	MaxResponseTime float64 `json:"max_response_time"`
}

// Thread-safe results storage
type SafeResults struct {
	mu      sync.Mutex
	results []TestResult
}

func (sr *SafeResults) Add(result TestResult) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.results = append(sr.results, result)
}

func (sr *SafeResults) GetAll() []TestResult {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	return sr.results
}

func getTimestamp() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

// createCartWorker creates a shopping cart
func createCartWorker(customerID int, results *SafeResults, wg *sync.WaitGroup) {
	defer wg.Done()
	
	start := time.Now()
	
	payload := map[string]int{"customer_id": customerID}
	jsonData, _ := json.Marshal(payload)
	
	resp, err := http.Post(
		BaseURL+"/shopping-carts",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	
	responseTime := time.Since(start).Milliseconds()
	
	result := TestResult{
		Operation:    "create_cart",
		ResponseTime: float64(responseTime),
		CustomerID:   customerID,
		Timestamp:    getTimestamp(),
	}
	
	if err != nil {
		result.Success = false
		result.StatusCode = 0
		result.Error = err.Error()
	} else {
		defer resp.Body.Close()
		result.Success = resp.StatusCode == 200 || resp.StatusCode == 201
		result.StatusCode = resp.StatusCode
	}
	
	results.Add(result)
}

// addItemsWorker adds items to a cart
func addItemsWorker(customerID int, results *SafeResults, wg *sync.WaitGroup) {
	defer wg.Done()
	
	start := time.Now()
	
	productID := rand.Intn(100000) + 1
	quantity := rand.Intn(10) + 1
	
	payload := map[string]int{
		"product_id": productID,
		"quantity":   quantity,
	}
	jsonData, _ := json.Marshal(payload)
	
	url := fmt.Sprintf("%s/shopping-carts/%d/items", BaseURL, customerID)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	
	responseTime := time.Since(start).Milliseconds()
	
	result := TestResult{
		Operation:    "add_items",
		ResponseTime: float64(responseTime),
		CustomerID:   customerID,
		ProductID:    productID,
		Quantity:     quantity,
		Timestamp:    getTimestamp(),
	}
	
	if err != nil {
		result.Success = false
		result.StatusCode = 0
		result.Error = err.Error()
	} else {
		defer resp.Body.Close()
		result.Success = resp.StatusCode == 200 || resp.StatusCode == 201
		result.StatusCode = resp.StatusCode
	}
	
	results.Add(result)
}

// getCartWorker retrieves a cart
func getCartWorker(customerID int, results *SafeResults, wg *sync.WaitGroup) {
	defer wg.Done()
	
	start := time.Now()
	
	url := fmt.Sprintf("%s/shopping-carts/%d", BaseURL, customerID)
	resp, err := http.Get(url)
	
	responseTime := time.Since(start).Milliseconds()
	
	result := TestResult{
		Operation:    "get_cart",
		ResponseTime: float64(responseTime),
		CustomerID:   customerID,
		Timestamp:    getTimestamp(),
	}
	
	if err != nil {
		result.Success = false
		result.StatusCode = 0
		result.Error = err.Error()
	} else {
		defer resp.Body.Close()
		result.Success = resp.StatusCode == 200
		result.StatusCode = resp.StatusCode
	}
	
	results.Add(result)
}

// calculateStats computes statistics from results
func calculateStats(results []TestResult) Statistics {
	stats := Statistics{
		TotalOperations: len(results),
		Operations:      make(map[string]OperationStats),
	}
	
	// Count successes
	for _, r := range results {
		if r.Success {
			stats.SuccessfulOperations++
		} else {
			stats.FailedOperations++
		}
	}
	
	// Calculate success rate
	if stats.TotalOperations > 0 {
		stats.SuccessRate = float64(stats.SuccessfulOperations) / float64(stats.TotalOperations) * 100
	}
	
	// Calculate per-operation stats
	operations := []string{"create_cart", "add_items", "get_cart"}
	for _, opType := range operations {
		var opResults []TestResult
		for _, r := range results {
			if r.Operation == opType {
				opResults = append(opResults, r)
			}
		}
		
		if len(opResults) > 0 {
			opStats := OperationStats{
				Count:           len(opResults),
				MinResponseTime: opResults[0].ResponseTime,
				MaxResponseTime: opResults[0].ResponseTime,
			}
			
			var totalTime float64
			for _, r := range opResults {
				if r.Success {
					opStats.Successful++
				} else {
					opStats.Failed++
				}
				
				totalTime += r.ResponseTime
				
				if r.ResponseTime < opStats.MinResponseTime {
					opStats.MinResponseTime = r.ResponseTime
				}
				if r.ResponseTime > opStats.MaxResponseTime {
					opStats.MaxResponseTime = r.ResponseTime
				}
			}
			
			opStats.AvgResponseTime = totalTime / float64(len(opResults))
			stats.Operations[opType] = opStats
		}
	}
	
	return stats
}

// testConnectivity checks if the service is reachable
func testConnectivity() error {
	fmt.Println("\nTesting connectivity...")
	
	resp, err := http.Get(BaseURL + "/health")
	if err != nil {
		return fmt.Errorf("cannot reach service: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		fmt.Println("✓ Service is healthy")
		return nil
	}
	
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("service returned status %d: %s", resp.StatusCode, string(body))
}

func main() {
	fmt.Println("============================================================")
	fmt.Println("Concurrent MySQL Shopping Cart Test (Go)")
	fmt.Println("============================================================")
	fmt.Printf("Target: %s\n", BaseURL)
	fmt.Printf("Concurrent Workers: %d\n", NumWorkers)
	fmt.Printf("Total Operations: %d\n", NumCreateCart+NumAddItems+NumGetCart)
	fmt.Printf("  - Create Cart: %d\n", NumCreateCart)
	fmt.Printf("  - Add Items: %d\n", NumAddItems)
	fmt.Printf("  - Get Cart: %d\n", NumGetCart)
	fmt.Printf("Output: %s\n", OutputFile)
	fmt.Println("============================================================")
	
	// Test connectivity
	if err := testConnectivity(); err != nil {
		log.Fatalf("✗ %v", err)
	}
	
	// Initialize
	results := &SafeResults{}
	rand.Seed(time.Now().UnixNano())
	
	// Generate customer IDs
	baseCustomerID := rand.Intn(90000) + 10000
	customerIDs := make([]int, NumCreateCart)
	for i := 0; i < NumCreateCart; i++ {
		customerIDs[i] = baseCustomerID + i
	}
	
	fmt.Printf("\nUsing customer IDs: %d - %d\n", baseCustomerID, baseCustomerID+NumCreateCart-1)
	
	startTime := time.Now()
	
	// Phase 1: Create carts concurrently
	fmt.Println("\nPhase 1: Creating shopping carts concurrently...")
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, NumWorkers) // Limit concurrent workers
	
	for _, customerID := range customerIDs {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire
		go func(cid int) {
			defer func() { <-semaphore }() // Release
			createCartWorker(cid, results, &wg)
		}(customerID)
	}
	wg.Wait()
	fmt.Println("✓ Phase 1 complete")
	
	// Small delay to ensure carts are created
	time.Sleep(500 * time.Millisecond)
	
	// Phase 2: Add items concurrently
	fmt.Println("Phase 2: Adding items to carts concurrently...")
	for _, customerID := range customerIDs {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(cid int) {
			defer func() { <-semaphore }()
			addItemsWorker(cid, results, &wg)
		}(customerID)
	}
	wg.Wait()
	fmt.Println("✓ Phase 2 complete")
	
	// Small delay
	time.Sleep(500 * time.Millisecond)
	
	// Phase 3: Get carts concurrently
	fmt.Println("Phase 3: Retrieving carts concurrently...")
	for _, customerID := range customerIDs {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(cid int) {
			defer func() { <-semaphore }()
			getCartWorker(cid, results, &wg)
		}(customerID)
	}
	wg.Wait()
	fmt.Println("✓ Phase 3 complete")
	
	totalDuration := time.Since(startTime).Seconds()
	
	// Calculate statistics
	allResults := results.GetAll()
	stats := calculateStats(allResults)
	
	// Prepare output
	output := TestOutput{
		TestMetadata: TestMetadata{
			BaseURL:              BaseURL,
			StartTime:            startTime.UTC().Format("2006-01-02T15:04:05Z"),
			TotalDurationSeconds: totalDuration,
			TotalOperations:      NumCreateCart + NumAddItems + NumGetCart,
			ConcurrentWorkers:    NumWorkers,
		},
		Statistics: stats,
		Results:    allResults,
	}
	
	// Save to file
	file, err := os.Create(OutputFile)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		log.Fatalf("Error encoding JSON: %v", err)
	}
	
	// Print summary
	fmt.Println("\n============================================================")
	fmt.Println("TEST SUMMARY")
	fmt.Println("============================================================")
	fmt.Printf("Total Duration: %.2f seconds\n", totalDuration)
	fmt.Printf("Total Operations: %d\n", stats.TotalOperations)
	fmt.Printf("Successful: %d\n", stats.SuccessfulOperations)
	fmt.Printf("Failed: %d\n", stats.FailedOperations)
	fmt.Printf("Success Rate: %.2f%%\n", stats.SuccessRate)
	fmt.Println()
	
	for opType, opStats := range stats.Operations {
		fmt.Printf("%s:\n", opType)
		fmt.Printf("  Count: %d\n", opStats.Count)
		fmt.Printf("  Success: %d/%d\n", opStats.Successful, opStats.Count)
		fmt.Printf("  Avg Response Time: %.2f ms\n", opStats.AvgResponseTime)
		fmt.Printf("  Min/Max: %.2f/%.2f ms\n", opStats.MinResponseTime, opStats.MaxResponseTime)
		fmt.Println()
	}
	
	fmt.Printf("Results saved to: %s\n", OutputFile)
	
	// Check requirements
	if totalDuration > 300 {
		fmt.Printf("⚠ WARNING: Test took longer than 5 minutes (%.2fs)\n", totalDuration)
	} else {
		fmt.Println("✓ Test completed within 5 minutes")
	}
	
	if stats.SuccessRate == 100 {
		fmt.Println("✓ All operations successful")
	} else {
		fmt.Printf("⚠ %d operations failed\n", stats.FailedOperations)
	}
	
	fmt.Println("============================================================")
}