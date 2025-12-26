package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	authServiceURL    = "http://localhost:8083"
	catalogServiceURL = "http://localhost:8082"
)

type LegacyRegisterRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Phone          string `json:"phone"`
	OrganizationID string `json:"organization_id"`
}

type LegacyLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LegacyLoginResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	User        struct {
		ID             string `json:"id"`
		Email          string `json:"email"`
		FirstName      string `json:"first_name"`
		LastName       string `json:"last_name"`
		OrganizationID string `json:"organization_id"`
	} `json:"user"`
}

type LegacyCreateProductRequest struct {
	OrganizationID  string `json:"organization_id"`
	SKU             string `json:"sku"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Category        string `json:"category"`
	UnitOfMeasure   string `json:"unit_of_measure"`
	BufferProfileID string `json:"buffer_profile_id,omitempty"`
}

type LegacyProductResponse struct {
	ID              string `json:"id"`
	OrganizationID  string `json:"organization_id"`
	SKU             string `json:"sku"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Category        string `json:"category"`
	UnitOfMeasure   string `json:"unit_of_measure"`
	Status          string `json:"status"`
	BufferProfileID string `json:"buffer_profile_id"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type LegacyCreateProductResponse struct {
	Product LegacyProductResponse `json:"product"`
}

type LegacyGetProductResponse struct {
	Product LegacyProductResponse `json:"product"`
}

type LegacyErrorResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
	Details   string `json:"details"`
}

func TestAuthCatalogFlow_CompleteUserJourney(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	organizationID := uuid.New().String()
	userEmail := fmt.Sprintf("test-%d@example.com", time.Now().Unix())
	userPassword := "SecurePassword123!"

	t.Run("1_RegisterUser_Success", func(t *testing.T) {
		registerReq := RegisterRequest{
			Email:          userEmail,
			Password:       userPassword,
			FirstName:      "Integration",
			LastName:       "Test",
			Phone:          "+1234567890",
			OrganizationID: organizationID,
		}

		resp := makeJSONRequest(t, "POST", authServiceURL+"/api/v1/auth/register", registerReq, "")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "User registration should succeed")
	})

	var accessToken string
	_ = accessToken // will be used

	t.Run("2_Login_Success", func(t *testing.T) {
		loginReq := LegacyLoginRequest{
			Email:    userEmail,
			Password: userPassword,
		}

		resp := makeJSONRequest(t, "POST", authServiceURL+"/api/v1/auth/login", loginReq, "")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "Login should succeed")

		var loginResp LegacyLoginResponse
		err := json.NewDecoder(resp.Body).Decode(&loginResp)
		require.NoError(t, err, "Should decode login response")

		assert.NotEmpty(t, loginResp.AccessToken, "Should receive access token")
		assert.Greater(t, loginResp.ExpiresIn, 0, "Token should have expiry time")
		assert.Equal(t, userEmail, loginResp.User.Email, "User email should match")
		assert.Equal(t, organizationID, loginResp.User.OrganizationID, "Organization ID should match")

		accessToken = loginResp.AccessToken
		_ = loginResp.User.ID // userID
	})

	var productID string
	productSKU := fmt.Sprintf("SKU-%d", time.Now().Unix())

	t.Run("3_CreateProduct_WithValidToken_Success", func(t *testing.T) {
		createProductReq := LegacyCreateProductRequest{
			OrganizationID: organizationID,
			SKU:            productSKU,
			Name:           "Integration Test Product",
			Description:    "Product created during integration testing",
			Category:       "Test Category",
			UnitOfMeasure:  "UNIT",
		}

		resp := makeJSONRequest(t, "POST", catalogServiceURL+"/api/v1/products", createProductReq, accessToken)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode, "Product creation should succeed")

		var createResp LegacyCreateProductResponse
		err := json.NewDecoder(resp.Body).Decode(&createResp)
		require.NoError(t, err, "Should decode create product response")

		assert.NotEmpty(t, createResp.Product.ID, "Product should have an ID")
		assert.Equal(t, productSKU, createResp.Product.SKU, "Product SKU should match")
		assert.Equal(t, organizationID, createResp.Product.OrganizationID, "Organization ID should match")
		assert.Equal(t, "active", createResp.Product.Status, "Product should be active by default")

		productID = createResp.Product.ID
	})

	t.Run("4_GetProduct_WithValidToken_Success", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/products/%s?organization_id=%s", catalogServiceURL, productID, organizationID)
		resp := makeRequest(t, "GET", url, nil, accessToken)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "Get product should succeed")

		var getResp LegacyGetProductResponse
		err := json.NewDecoder(resp.Body).Decode(&getResp)
		require.NoError(t, err, "Should decode get product response")

		assert.Equal(t, productID, getResp.Product.ID, "Product ID should match")
		assert.Equal(t, productSKU, getResp.Product.SKU, "Product SKU should match")
		assert.Equal(t, "Integration Test Product", getResp.Product.Name, "Product name should match")
		assert.Equal(t, organizationID, getResp.Product.OrganizationID, "Organization ID should match")
	})

	t.Run("5_CreateProduct_WithoutToken_Unauthorized", func(t *testing.T) {
		createProductReq := LegacyCreateProductRequest{
			OrganizationID: organizationID,
			SKU:            "UNAUTHORIZED-SKU",
			Name:           "Should Fail",
			Description:    "This should not be created",
			Category:       "Test",
			UnitOfMeasure:  "UNIT",
		}

		resp := makeJSONRequest(t, "POST", catalogServiceURL+"/api/v1/products", createProductReq, "")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should return unauthorized without token")

		var errResp LegacyErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		assert.Contains(t, errResp.Message, "unauthorized", "Error message should indicate unauthorized")
	})

	t.Run("6_CreateProduct_WithInvalidToken_Unauthorized", func(t *testing.T) {
		createProductReq := LegacyCreateProductRequest{
			OrganizationID: organizationID,
			SKU:            "INVALID-TOKEN-SKU",
			Name:           "Should Fail",
			Description:    "This should not be created",
			Category:       "Test",
			UnitOfMeasure:  "UNIT",
		}

		invalidToken := "invalid.jwt.token"
		resp := makeJSONRequest(t, "POST", catalogServiceURL+"/api/v1/products", createProductReq, invalidToken)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should return unauthorized with invalid token")

		var errResp LegacyErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		assert.Contains(t, errResp.Message, "unauthorized", "Error message should indicate unauthorized")
	})

	t.Run("7_GetProduct_WithoutToken_Unauthorized", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/products/%s?organization_id=%s", catalogServiceURL, productID, organizationID)
		resp := makeRequest(t, "GET", url, nil, "")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should return unauthorized without token")
	})

	t.Run("8_ListProducts_WithValidToken_Success", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/products?organization_id=%s&page=1&page_size=10", catalogServiceURL, organizationID)
		resp := makeRequest(t, "GET", url, nil, accessToken)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "List products should succeed")

		var listResp struct {
			Products []LegacyProductResponse `json:"products"`
			Total    int                     `json:"total"`
			Page     int                     `json:"page"`
			PageSize int                     `json:"page_size"`
		}
		err := json.NewDecoder(resp.Body).Decode(&listResp)
		require.NoError(t, err, "Should decode list products response")

		assert.GreaterOrEqual(t, len(listResp.Products), 1, "Should have at least one product")
		assert.GreaterOrEqual(t, listResp.Total, 1, "Total count should be at least 1")

		found := false
		for _, p := range listResp.Products {
			if p.ID == productID {
				found = true
				break
			}
		}
		assert.True(t, found, "Created product should be in the list")
	})

	t.Run("9_UpdateProduct_WithValidToken_Success", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"id":              productID,
			"organization_id": organizationID,
			"name":            "Updated Product Name",
			"description":     "Updated description",
			"category":        "Updated Category",
			"unit_of_measure": "UNIT",
			"status":          "active",
		}

		url := fmt.Sprintf("%s/api/v1/products/%s", catalogServiceURL, productID)
		resp := makeJSONRequest(t, "PUT", url, updateReq, accessToken)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "Update product should succeed")

		var updateResp struct {
			Product LegacyProductResponse `json:"product"`
		}
		err := json.NewDecoder(resp.Body).Decode(&updateResp)
		require.NoError(t, err, "Should decode update product response")

		assert.Equal(t, "Updated Product Name", updateResp.Product.Name, "Product name should be updated")
		assert.Equal(t, "Updated description", updateResp.Product.Description, "Product description should be updated")
	})

	t.Run("10_SearchProducts_WithValidToken_Success", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/products/search?organization_id=%s&query=%s", catalogServiceURL, organizationID, "Updated")
		resp := makeRequest(t, "GET", url, nil, accessToken)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "Search products should succeed")

		var searchResp struct {
			Products []LegacyProductResponse `json:"products"`
			Total    int                     `json:"total"`
		}
		err := json.NewDecoder(resp.Body).Decode(&searchResp)
		require.NoError(t, err, "Should decode search products response")

		assert.GreaterOrEqual(t, len(searchResp.Products), 1, "Should find the updated product")
	})

	t.Run("11_DeleteProduct_WithValidToken_Success", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/products/%s?organization_id=%s", catalogServiceURL, productID, organizationID)
		resp := makeRequest(t, "DELETE", url, nil, accessToken)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "Delete product should succeed")

		var deleteResp struct {
			Success bool `json:"success"`
		}
		err := json.NewDecoder(resp.Body).Decode(&deleteResp)
		require.NoError(t, err, "Should decode delete product response")

		assert.True(t, deleteResp.Success, "Delete operation should return success")
	})

	t.Run("12_GetDeletedProduct_ShouldFail", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/products/%s?organization_id=%s", catalogServiceURL, productID, organizationID)
		resp := makeRequest(t, "GET", url, nil, accessToken)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should not find deleted product")
	})
}

func makeJSONRequest(t *testing.T, method, url string, body interface{}, token string) *http.Response {
	t.Helper()

	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(t, err, "Should marshal request body")
	}

	return makeRequest(t, method, url, reqBody, token)
}

func makeRequest(t *testing.T, method, url string, body []byte, token string) *http.Response {
	t.Helper()

	var reqBody *bytes.Buffer
	if body != nil {
		reqBody = bytes.NewBuffer(body)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, reqBody)
	require.NoError(t, err, "Should create HTTP request")

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	require.NoError(t, err, "Should execute HTTP request")

	return resp
}
