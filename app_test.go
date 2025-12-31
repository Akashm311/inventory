package main


import (
	"fmt"
	"log"
	"testing"
	"net/http"
	"net/http/httptest"
	"bytes"
	"encoding/json"

)


var a App


func TestMain(m *testing.M) {
	err := a.Initialise(DbUser, DbPassword, "test")
	if err != nil {
		log.Fatal("Error occurred while initialising the database")
	}

	// Additional setup (e.g., creating a test table) can be done here

	createTable()

	m.Run()
}





func createTable() {
	createTableQuery := `CREATE TABLE IF NOT EXISTS products (
		id int NOT NULL AUTO_INCREMENT,
		name varchar(255) NOT NULL,
		quantity int,
		price float(10,7),
		PRIMARY KEY (id)
	);`
	_, err := a.DB.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	// Delete all records from the products table.
	_, err := a.DB.Exec("DELETE FROM products")
	if err != nil {
		log.Printf("Error clearing table: %v", err)
	}
	// Reset the auto-increment counter.
	_, err = a.DB.Exec("ALTER TABLE products AUTO_INCREMENT=1")
	if err != nil {
		log.Printf("Error resetting auto_increment: %v", err)
	}
	log.Println("clearTable completed")
}


func addProduct(name string, quantity int, price float64) {
	query := fmt.Sprintf("INSERT INTO products(name, quantity, price) VALUES('%v', %v, %v)", name, quantity, price)
	_, err := a.DB.Exec(query)
	if err != nil {
		log.Printf("Error adding product: %v", err)
	}
}

func TestGetProduct(t *testing.T) {
	clearTable()
	// Add a new product. With the auto-increment reset, the new product’s ID will be 1.
	addProduct("keyboard", 100, 500)
	
	// Create a new GET request for the product with ID 1.
	request, err := http.NewRequest("GET", "/product/1", nil)
	if err != nil {
		t.Fatalf("Couldn't create request: %v", err)
	}
	
	// Send the request and get the response.
	response := sendRequest(request)
	
	// Verify the HTTP status code.
	checkStatusCode(t, http.StatusOK, response.Code)
}


func sendRequest(request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	a.Router.ServeHTTP(recorder, request)
	return recorder
}

func checkStatusCode(t *testing.T, expectedStatusCode int, actualStatusCode int) {
	if expectedStatusCode != actualStatusCode {
		t.Errorf("Expected status: %v, Received: %v", expectedStatusCode, actualStatusCode)
	}
}

func TestCreateProduct(t *testing.T) {
    clearTable()
    var product = []byte(`{"name":"chair", "quantity":1, "price":100}`)
    req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(product))
    req.Header.Set("Content-Type", "application/json")


    response := sendRequest(req)
    checkStatusCode(t, http.StatusCreated, response.Code)


    var m map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &m)


    if m["name"] != "chair" {
        t.Errorf("Expected name: %v, Got: %v", "chair", m["name"])
    }
    if m["quantity"] != 1.0 { // Numbers are unmarshaled as float64.
        t.Errorf("Expected quantity: %v, Got: %v", 1.0, m["quantity"])
    }
}


func TestDeleteProduct(t *testing.T) {
    // Clear existing data.
    clearTable()


    // Add a product "connector" with quantity 10 and price 10.
    addProduct("connector", 10, 10)


    // Retrieve the product using GET request.
    req, _ := http.NewRequest("GET", "/product/1", nil)
    response := sendRequest(req)
    checkStatusCode(t, http.StatusOK, response.Code)


    // Validate the retrieved product details.
    var m map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &m)
    if m["name"] != "connector" {
        t.Errorf("Expected name: %v, Got: %v", "connector", m["name"])
    }
    if m["quantity"] != 10.0 {
        t.Errorf("Expected quantity: %v, Got: %v", 10.0, m["quantity"])
    }


    // Delete the product using DELETE request.
    req, _ = http.NewRequest("DELETE", "/product/1", nil)
    response = sendRequest(req)
    checkStatusCode(t, http.StatusOK, response.Code)


    // Try to retrieve the deleted product.
    req, _ = http.NewRequest("GET", "/product/1", nil)
    response = sendRequest(req)
    // Expect a 404 Not Found status since the product has been deleted.
    checkStatusCode(t, http.StatusNotFound, response.Code)
}

func TestUpdateProduct(t *testing.T) {
    clearTable()
    addProduct("connector", 10, 10)


    // Retrieve the original product.
    req, _ := http.NewRequest("GET", "/product/1", nil)
    response := sendRequest(req)
    checkStatusCode(t, http.StatusOK, response.Code)


    var oldValue map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &oldValue)


    // Prepare updated product payload; here, the quantity is changed while keeping the name consistent.
    product := []byte(`{"name":"connector", "quantity":1, "price":10}`)
    req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(product))
    req.Header.Set("Content-Type", "application/json")
    response = sendRequest(req)


    var newValue map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &newValue)


    // Verify that the product ID remains unchanged.
    if oldValue["id"] != newValue["id"] {
        t.Errorf("Expected id: %v, Got: %v", oldValue["id"], newValue["id"])
    }


    // Validate updated fields.
    if newValue["name"] != "connector" {
        t.Errorf("Expected name: connector, Got: %v", newValue["name"])
    }
    if newValue["price"] != float64(10) {
        t.Errorf("Expected price: 10, Got: %v", newValue["price"])
    }
    // Confirm that the quantity has been updated.
    if oldValue["quantity"] == newValue["quantity"] {
        t.Errorf("Expected quantity to change from %v, but got %v", oldValue["quantity"], newValue["quantity"])
    }
}
