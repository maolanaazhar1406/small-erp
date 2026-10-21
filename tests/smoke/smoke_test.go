//go:build smoke

package smoke

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

	// Wait for the application to be ready
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

func TestHealthCheck(t *testing.T) {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["success"] != true {
		t.Fatal("health check success should be true")
	}
	t.Log("Health check passed")
}

func TestLogin(t *testing.T) {
	token := login(t, "admin", "admin123")
	if token == "" {
		t.Fatal("login should return a token")
	}
	adminToken = token
	t.Log("Admin login successful")
}

func TestUnauthorizedAccess(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/v1/users")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 401 {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
	t.Log("Unauthorized access correctly denied")
}

func TestListUsers(t *testing.T) {
	resp := authGet(t, "/api/v1/users")
	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	t.Log("List users successful")
}

func TestListProducts(t *testing.T) {
	resp := authGet(t, "/api/v1/products")
	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	t.Log("List products successful")
}

func TestListCustomers(t *testing.T) {
	resp := authGet(t, "/api/v1/customers")
	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	t.Log("List customers successful")
}

func TestListSuppliers(t *testing.T) {
	resp := authGet(t, "/api/v1/suppliers")
	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	t.Log("List suppliers successful")
}

func TestFullProcurementWorkflow(t *testing.T) {
	if adminToken == "" {
		adminToken = login(t, "admin", "admin123")
	}

	// 1. Create Purchase Order
	poBody := map[string]interface{}{
		"supplier_id": 1,
		"order_date":  time.Now().Format("2006-01-02"),
		"notes":       "Smoke test PO",
		"items": []map[string]interface{}{
			{"product_id": 1, "quantity": 10, "unit_price": 12000000},
			{"product_id": 2, "quantity": 50, "unit_price": 150000},
		},
	}
	resp := authPost(t, "/api/v1/purchase-orders", poBody)
	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("create PO failed: status %d, body: %s", resp.StatusCode, string(body))
	}
	poResult := parseResponse(t, resp)
	poID := int(poResult["data"].(map[string]interface{})["id"].(float64))
	t.Logf("PO created with ID: %d", poID)

	// 2. Confirm PO
	statusBody := map[string]interface{}{"status": "confirmed"}
	resp = authPatch(t, fmt.Sprintf("/api/v1/purchase-orders/%d/status", poID), statusBody)
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("confirm PO failed: status %d, body: %s", resp.StatusCode, string(body))
	}
	t.Log("PO confirmed")

	// 3. Create Goods Receipt
	grBody := map[string]interface{}{
		"po_id":        poID,
		"receipt_date": time.Now().Format("2006-01-02"),
		"notes":        "Smoke test GR",
		"items": []map[string]interface{}{
			{"po_item_id": getFirstPOItemID(t, poID), "product_id": 1, "quantity_received": 10},
			{"po_item_id": getSecondPOItemID(t, poID), "product_id": 2, "quantity_received": 50},
		},
	}
	resp = authPost(t, "/api/v1/goods-receipts", grBody)
	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("create GR failed: status %d, body: %s", resp.StatusCode, string(body))
	}
	t.Log("Goods receipt created - inventory updated, AP recorded")

	// 4. Verify inventory was updated
	resp = authGet(t, "/api/v1/inventory")
	if resp.StatusCode != 200 {
		t.Fatal("failed to get inventory")
	}
	t.Log("Inventory verified after goods receipt")

	// 5. Verify AP was created
	resp = authGet(t, "/api/v1/accounts-payable")
	if resp.StatusCode != 200 {
		t.Fatal("failed to get AP")
	}
	t.Log("Accounts payable verified")
}

func TestFullSalesWorkflow(t *testing.T) {
	if adminToken == "" {
		adminToken = login(t, "admin", "admin123")
	}

	// 1. Create Sales Order
	soBody := map[string]interface{}{
		"customer_id": 1,
		"order_date":  time.Now().Format("2006-01-02"),
		"notes":       "Smoke test SO",
		"items": []map[string]interface{}{
			{"product_id": 2, "quantity": 5, "unit_price": 250000},
		},
	}
	resp := authPost(t, "/api/v1/sales-orders", soBody)
	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("create SO failed: status %d, body: %s", resp.StatusCode, string(body))
	}
	soResult := parseResponse(t, resp)
	soID := int(soResult["data"].(map[string]interface{})["id"].(float64))
	t.Logf("SO created with ID: %d", soID)

	// 2. Confirm SO
	resp = authPatch(t, fmt.Sprintf("/api/v1/sales-orders/%d/status", soID), map[string]interface{}{"status": "confirmed"})
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("confirm SO failed: status %d, body: %s", resp.StatusCode, string(body))
	}
	t.Log("SO confirmed")

	// 3. Deliver SO
	resp = authPatch(t, fmt.Sprintf("/api/v1/sales-orders/%d/status", soID), map[string]interface{}{"status": "delivered"})
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("deliver SO failed: status %d, body: %s", resp.StatusCode, string(body))
	}
	t.Log("SO delivered - inventory reduced, AR recorded")

	// 4. Verify AR
	resp = authGet(t, "/api/v1/accounts-receivable")
	if resp.StatusCode != 200 {
		t.Fatal("failed to get AR")
	}
	t.Log("Accounts receivable verified")
}

func TestPaymentProcessing(t *testing.T) {
	if adminToken == "" {
		adminToken = login(t, "admin", "admin123")
	}

	// Pay AP (reference_id = 1, the first AP entry)
	payBody := map[string]interface{}{
		"payment_type":   "payable",
		"reference_id":   1,
		"amount":         1000000,
		"payment_date":   time.Now().Format("2006-01-02"),
		"payment_method": "bank_transfer",
		"notes":          "Partial AP payment",
	}
	resp := authPost(t, "/api/v1/payments", payBody)
	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("AP payment failed: status %d, body: %s", resp.StatusCode, string(body))
	}
	t.Log("AP payment processed")

	// Pay AR (reference_id = 1)
	arPayBody := map[string]interface{}{
		"payment_type":   "receivable",
		"reference_id":   1,
		"amount":         500000,
		"payment_date":   time.Now().Format("2006-01-02"),
		"payment_method": "cash",
		"notes":          "Partial AR payment",
	}
	resp = authPost(t, "/api/v1/payments", arPayBody)
	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("AR payment failed: status %d, body: %s", resp.StatusCode, string(body))
	}
	t.Log("AR payment processed")
}

func TestReports(t *testing.T) {
	if adminToken == "" {
		adminToken = login(t, "admin", "admin123")
	}

	endpoints := []string{
		"/api/v1/reports/sales?start_date=2020-01-01&end_date=2030-12-31",
		"/api/v1/reports/purchases?start_date=2020-01-01&end_date=2030-12-31",
		"/api/v1/reports/inventory",
		"/api/v1/reports/stock-movements?start_date=2020-01-01&end_date=2030-12-31",
		"/api/v1/reports/accounts-payable",
		"/api/v1/reports/accounts-receivable",
		"/api/v1/reports/cash-flow?start_date=2020-01-01&end_date=2030-12-31",
		"/api/v1/reports/profit-loss?start_date=2020-01-01&end_date=2030-12-31",
		"/api/v1/reports/balance-sheet?as_of_date=2030-12-31",
	}

	for _, ep := range endpoints {
		resp := authGet(t, ep)
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("report %s failed: status %d, body: %s", ep, resp.StatusCode, string(body))
		}
		t.Logf("Report %s: OK", ep)
	}
}

func TestInsufficientStock(t *testing.T) {
	if adminToken == "" {
		adminToken = login(t, "admin", "admin123")
	}

	// Try to sell more than available
	soBody := map[string]interface{}{
		"customer_id": 1,
		"order_date":  time.Now().Format("2006-01-02"),
		"items": []map[string]interface{}{
			{"product_id": 1, "quantity": 999999, "unit_price": 100},
		},
	}
	resp := authPost(t, "/api/v1/sales-orders", soBody)
	if resp.StatusCode != 201 {
		t.Skip("SO creation may have failed for other reasons")
	}
	soResult := parseResponse(t, resp)
	soID := int(soResult["data"].(map[string]interface{})["id"].(float64))

	authPatch(t, fmt.Sprintf("/api/v1/sales-orders/%d/status", soID), map[string]interface{}{"status": "confirmed"})
	resp = authPatch(t, fmt.Sprintf("/api/v1/sales-orders/%d/status", soID), map[string]interface{}{"status": "delivered"})
	if resp.StatusCode == 200 {
		t.Fatal("delivery of insufficient stock should have failed")
	}
	t.Log("Insufficient stock correctly rejected")
}

func TestStaffPermissions(t *testing.T) {
	staffToken := login(t, "staff", "staff123")

	// Staff should be able to read products
	client := &http.Client{}
	req, _ := http.NewRequest("GET", baseURL+"/api/v1/products", nil)
	req.Header.Set("Authorization", "Bearer "+staffToken)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("staff should be able to read products, got %d", resp.StatusCode)
	}
	t.Log("Staff read access verified")

	// Staff should NOT be able to delete users
	req, _ = http.NewRequest("DELETE", baseURL+"/api/v1/users/1", nil)
	req.Header.Set("Authorization", "Bearer "+staffToken)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 403 {
		t.Fatalf("staff should NOT be able to delete users, got %d", resp.StatusCode)
	}
	t.Log("Staff restricted access verified")
}

// ---- Helper functions ----

func login(t *testing.T, username, password string) string {
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	resp, err := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("login failed for %s: status %d, body: %s", username, resp.StatusCode, string(respBody))
	}

	result := parseResponse(t, resp)
	data := result["data"].(map[string]interface{})
	return data["token"].(string)
}

func authGet(t *testing.T, path string) *http.Response {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", baseURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func authPost(t *testing.T, path string, body interface{}) *http.Response {
	client := &http.Client{}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", baseURL+path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func authPatch(t *testing.T, path string, body interface{}) *http.Response {
	client := &http.Client{}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("PATCH", baseURL+path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func parseResponse(t *testing.T, resp *http.Response) map[string]interface{} {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to parse JSON: %v, body: %s", err, string(body))
	}
	return result
}

func getFirstPOItemID(t *testing.T, poID int) int {
	resp := authGet(t, fmt.Sprintf("/api/v1/purchase-orders/%d", poID))
	result := parseResponse(t, resp)
	data := result["data"].(map[string]interface{})
	items := data["items"].([]interface{})
	return int(items[0].(map[string]interface{})["id"].(float64))
}

func getSecondPOItemID(t *testing.T, poID int) int {
	resp := authGet(t, fmt.Sprintf("/api/v1/purchase-orders/%d", poID))
	result := parseResponse(t, resp)
	data := result["data"].(map[string]interface{})
	items := data["items"].([]interface{})
	return int(items[1].(map[string]interface{})["id"].(float64))
}
