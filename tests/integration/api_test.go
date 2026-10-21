//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

var baseURL string
var adminToken string

func TestMain(m *testing.M) {
	baseURL = os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Wait for app readiness
	for i := 0; i < 30; i++ {
		resp, err := http.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			break
		}
		time.Sleep(2 * time.Second)
	}

	os.Exit(m.Run())
}

func loginAdmin(t *testing.T) string {
	if adminToken != "" {
		return adminToken
	}
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin123"})
	resp, err := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	defer resp.Body.Close()
	result := parseJSON(t, resp)
	adminToken = result["data"].(map[string]interface{})["token"].(string)
	return adminToken
}

func parseJSON(t *testing.T, resp *http.Response) map[string]interface{} {
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return result
}

func doGet(t *testing.T, path string) *http.Response {
	token := loginAdmin(t)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", baseURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("GET %s failed: %v", path, err)
	}
	return resp
}

func doPost(t *testing.T, path string, data interface{}) *http.Response {
	token := loginAdmin(t)
	client := &http.Client{}
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", baseURL+path, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST %s failed: %v", path, err)
	}
	return resp
}

func doPut(t *testing.T, path string, data interface{}) *http.Response {
	token := loginAdmin(t)
	client := &http.Client{}
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest("PUT", baseURL+path, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("PUT %s failed: %v", path, err)
	}
	return resp
}

// ---- Auth Tests ----

func TestLoginSuccess(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin123"})
	resp, err := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	result := parseJSON(t, resp)
	if result["success"] != true {
		t.Fatal("login should succeed")
	}
	data := result["data"].(map[string]interface{})
	if data["token"] == nil || data["token"] == "" {
		t.Fatal("token should not be empty")
	}
}

func TestLoginBadPassword(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "wrongpassword"})
	resp, _ := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
	defer resp.Body.Close()

	if resp.StatusCode != 401 {
		t.Fatalf("expected 401 for wrong password, got %d", resp.StatusCode)
	}
}

func TestLoginBadUsername(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"username": "nonexistent", "password": "admin123"})
	resp, _ := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
	defer resp.Body.Close()

	if resp.StatusCode != 401 {
		t.Fatalf("expected 401 for wrong username, got %d", resp.StatusCode)
	}
}

func TestLoginEmptyBody(t *testing.T) {
	body, _ := json.Marshal(map[string]string{})
	resp, _ := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Fatalf("expected 400 for empty body, got %d", resp.StatusCode)
	}
}

// ---- CRUD Tests ----

func TestCategoryLifecycle(t *testing.T) {
	// Create
	resp := doPost(t, "/api/v1/product-categories", map[string]string{
		"name":        "Integration Test Category",
		"description": "Created by integration test",
	})
	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("create category: expected 201, got %d: %s", resp.StatusCode, body)
	}
	result := parseJSON(t, resp)
	catID := int(result["data"].(map[string]interface{})["id"].(float64))

	// Read
	resp = doGet(t, fmt.Sprintf("/api/v1/product-categories/%d", catID))
	if resp.StatusCode != 200 {
		t.Fatalf("get category: expected 200, got %d", resp.StatusCode)
	}

	// Update
	resp = doPut(t, fmt.Sprintf("/api/v1/product-categories/%d", catID), map[string]string{
		"name":        "Updated Category",
		"description": "Updated by integration test",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("update category: expected 200, got %d", resp.StatusCode)
	}

	// List
	resp = doGet(t, "/api/v1/product-categories")
	if resp.StatusCode != 200 {
		t.Fatalf("list categories: expected 200, got %d", resp.StatusCode)
	}
}

func TestUnitLifecycle(t *testing.T) {
	resp := doPost(t, "/api/v1/units", map[string]string{
		"name":         "Integration Test Unit",
		"abbreviation": "itu",
	})
	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("create unit: expected 201, got %d: %s", resp.StatusCode, body)
	}
}

func TestProductLifecycle(t *testing.T) {
	resp := doPost(t, "/api/v1/products", map[string]interface{}{
		"code":  "INT-001",
		"name":  "Integration Test Product",
		"price": 100000,
		"cost":  50000,
	})
	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("create product: expected 201, got %d: %s", resp.StatusCode, body)
	}
	result := parseJSON(t, resp)
	prodID := int(result["data"].(map[string]interface{})["id"].(float64))

	// Can be read
	resp = doGet(t, fmt.Sprintf("/api/v1/products/%d", prodID))
	if resp.StatusCode != 200 {
		t.Fatalf("get product: expected 200, got %d", resp.StatusCode)
	}

	// List with pagination
	resp = doGet(t, "/api/v1/products?page=1&per_page=5")
	if resp.StatusCode != 200 {
		t.Fatalf("list products: expected 200, got %d", resp.StatusCode)
	}
}

func TestCustomerAndSupplierLifecycle(t *testing.T) {
	// Customer
	resp := doPost(t, "/api/v1/customers", map[string]string{
		"code":  "INT-CUS-001",
		"name":  "Integration Test Customer",
		"email": "intcust@test.com",
	})
	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("create customer: expected 201, got %d: %s", resp.StatusCode, body)
	}

	// Supplier
	resp = doPost(t, "/api/v1/suppliers", map[string]string{
		"code":  "INT-SUP-001",
		"name":  "Integration Test Supplier",
		"email": "intsup@test.com",
	})
	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("create supplier: expected 201, got %d: %s", resp.StatusCode, body)
	}
}

// ---- Validation Tests ----

func TestCreateProductMissingFields(t *testing.T) {
	resp := doPost(t, "/api/v1/products", map[string]string{})
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400 for missing fields, got %d", resp.StatusCode)
	}
}

func TestCreateUserMissingFields(t *testing.T) {
	resp := doPost(t, "/api/v1/users", map[string]string{})
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400 for missing fields, got %d", resp.StatusCode)
	}
}

func TestDuplicatePOReceipt(t *testing.T) {
	// This test is run after smoke tests so a PO has already been received
	// Trying to receipt PO #1 again should fail
	grBody := map[string]interface{}{
		"po_id":        1,
		"receipt_date": time.Now().Format("2006-01-02"),
		"items": []map[string]interface{}{
			{"po_item_id": 1, "product_id": 1, "quantity_received": 5},
		},
	}
	resp := doPost(t, "/api/v1/goods-receipts", grBody)
	if resp.StatusCode == 201 {
		t.Fatal("duplicate goods receipt should have been rejected")
	}
	t.Logf("Duplicate GR correctly rejected with status %d", resp.StatusCode)
}

func TestJournalEntryUnbalanced(t *testing.T) {
	jeBody := map[string]interface{}{
		"entry_date":  time.Now().Format("2006-01-02"),
		"description": "Unbalanced entry",
		"lines": []map[string]interface{}{
			{"account_id": 1, "debit": 1000, "credit": 0, "description": "debit"},
			{"account_id": 2, "debit": 0, "credit": 500, "description": "credit"},
		},
	}
	resp := doPost(t, "/api/v1/journal-entries", jeBody)
	if resp.StatusCode == 201 {
		t.Fatal("unbalanced journal entry should have been rejected")
	}
	t.Logf("Unbalanced JE correctly rejected with status %d", resp.StatusCode)
}

func TestInvalidPOStatusTransition(t *testing.T) {
	// Create a PO
	poBody := map[string]interface{}{
		"supplier_id": 1,
		"order_date":  time.Now().Format("2006-01-02"),
		"items": []map[string]interface{}{
			{"product_id": 1, "quantity": 1, "unit_price": 100},
		},
	}
	resp := doPost(t, "/api/v1/purchase-orders", poBody)
	result := parseJSON(t, resp)
	poID := int(result["data"].(map[string]interface{})["id"].(float64))

	// Try to transition draft → received (invalid, must go through confirmed)
	client := &http.Client{}
	body, _ := json.Marshal(map[string]string{"status": "received"})
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("%s/api/v1/purchase-orders/%d/status", baseURL, poID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)

	if resp.StatusCode == 200 {
		t.Fatal("invalid status transition should have been rejected")
	}
	t.Logf("Invalid status transition correctly rejected with status %d", resp.StatusCode)
}

func TestPaymentExceedingBalance(t *testing.T) {
	payBody := map[string]interface{}{
		"payment_type":   "payable",
		"reference_id":   1,
		"amount":         999999999999,
		"payment_date":   time.Now().Format("2006-01-02"),
		"payment_method": "cash",
	}
	resp := doPost(t, "/api/v1/payments", payBody)
	if resp.StatusCode == 201 {
		t.Fatal("payment exceeding balance should have been rejected")
	}
	t.Logf("Excessive payment correctly rejected with status %d", resp.StatusCode)
}
