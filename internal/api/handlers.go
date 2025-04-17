package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"kubernetes-api/internal/auth"
	"kubernetes-api/internal/database"
	"kubernetes-api/internal/metrics"
	"kubernetes-api/internal/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// healthHandler is the handler for the /api/health endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := models.ApiResponse{
		Status:  "success",
		Message: "Service is healthy",
		Data: map[string]interface{}{
			"version": "1.0.0",
			"uptime":  time.Now().Unix(), // This should be actual uptime in a real app
		},
	}
	json.NewEncoder(w).Encode(resp)
}

// registerHandler handles user registration
func registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" || req.Email == "" {
		http.Error(w, "Username, password, and email are required", http.StatusBadRequest)
		return
	}

	// Hash the password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		logrus.WithError(err).Error("Failed to hash password")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Insert user into database
	var userID int
	err = metrics.TrackDatabaseOperation("create_user", func() error {
		return database.DB.QueryRow(
			"INSERT INTO users (username, password_hash, email) VALUES ($1, $2, $3) RETURNING id",
			req.Username, passwordHash, req.Email,
		).Scan(&userID)
	})

	if err != nil {
		logrus.WithError(err).Error("Failed to create user")
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Create user object
	user := models.User{
		ID:           userID,
		Username:     req.Username,
		PasswordHash: passwordHash,
		Email:        req.Email,
		CreatedAt:    time.Now(),
	}

	// Generate JWT
	token, err := auth.GenerateJWT(user)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate JWT")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return response
	resp := models.AuthResponse{
		Token:   token,
		User:    user,
		Message: "User registered successfully",
	}

	json.NewEncoder(w).Encode(resp)
}

// loginHandler handles user login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Query user from database
	var user models.User
	var passwordHash string
	err := metrics.TrackDatabaseOperation("get_user", func() error {
		return database.DB.QueryRow(
			"SELECT id, username, password_hash, email, created_at FROM users WHERE username = $1",
			req.Username,
		).Scan(&user.ID, &user.Username, &passwordHash, &user.Email, &user.CreatedAt)
	})

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		} else {
			logrus.WithError(err).Error("Failed to query user")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Verify password
	if !auth.CheckPasswordHash(req.Password, passwordHash) {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	token, err := auth.GenerateJWT(user)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate JWT")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return response
	resp := models.AuthResponse{
		Token:   token,
		User:    user,
		Message: "Login successful",
	}

	json.NewEncoder(w).Encode(resp)
}

// itemsHandler handles CRUD operations for items
func itemsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		getItemsHandler(w, r)
	case http.MethodPost:
		createItemHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getItemsHandler handles GET /api/v1/items
func getItemsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (set by auth middleware)
	_, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Query items from database
	var items []models.Item
	err := metrics.TrackDatabaseOperation("get_items", func() error {
		rows, err := database.DB.Query("SELECT id, name, description, created_at, updated_at FROM items")
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var item models.Item
			if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
				return err
			}
			items = append(items, item)
		}

		return rows.Err()
	})

	if err != nil {
		logrus.WithError(err).Error("Failed to query items")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return response
	resp := models.ApiResponse{
		Status: "success",
		Data:   items,
	}

	json.NewEncoder(w).Encode(resp)
}

// createItemHandler handles POST /api/v1/items
func createItemHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (set by auth middleware)
	_, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.ItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Insert item into database
	var itemID int
	err := metrics.TrackDatabaseOperation("create_item", func() error {
		return database.DB.QueryRow(
			"INSERT INTO items (name, description) VALUES ($1, $2) RETURNING id",
			req.Name, req.Description,
		).Scan(&itemID)
	})

	if err != nil {
		logrus.WithError(err).Error("Failed to create item")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Query the created item
	var item models.Item
	err = metrics.TrackDatabaseOperation("get_item", func() error {
		return database.DB.QueryRow(
			"SELECT id, name, description, created_at, updated_at FROM items WHERE id = $1",
			itemID,
		).Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt)
	})

	if err != nil {
		logrus.WithError(err).Error("Failed to query created item")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return response
	resp := models.ApiResponse{
		Status:  "success",
		Message: "Item created successfully",
		Data:    item,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// itemHandler handles operations on a single item
func itemHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract item ID from URL
	vars := mux.Vars(r)
	itemID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getItemHandler(w, r, itemID)
	case http.MethodPut:
		updateItemHandler(w, r, itemID)
	case http.MethodDelete:
		deleteItemHandler(w, r, itemID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getItemHandler handles GET /api/v1/items/{id}
func getItemHandler(w http.ResponseWriter, r *http.Request, itemID int) {
	// Query item from database
	var item models.Item
	err := metrics.TrackDatabaseOperation("get_item", func() error {
		return database.DB.QueryRow(
			"SELECT id, name, description, created_at, updated_at FROM items WHERE id = $1",
			itemID,
		).Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt)
	})

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Item not found", http.StatusNotFound)
		} else {
			logrus.WithError(err).Error("Failed to query item")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return response
	resp := models.ApiResponse{
		Status: "success",
		Data:   item,
	}

	json.NewEncoder(w).Encode(resp)
}

// updateItemHandler handles PUT /api/v1/items/{id}
func updateItemHandler(w http.ResponseWriter, r *http.Request, itemID int) {
	var req models.ItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Update item in database
	err := metrics.TrackDatabaseOperation("update_item", func() error {
		result, err := database.DB.Exec(
			"UPDATE items SET name = $1, description = $2, updated_at = NOW() WHERE id = $3",
			req.Name, req.Description, itemID,
		)
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Item not found", http.StatusNotFound)
		} else {
			logrus.WithError(err).Error("Failed to update item")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Query the updated item
	var item models.Item
	err = metrics.TrackDatabaseOperation("get_item", func() error {
		return database.DB.QueryRow(
			"SELECT id, name, description, created_at, updated_at FROM items WHERE id = $1",
			itemID,
		).Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt)
	})

	if err != nil {
		logrus.WithError(err).Error("Failed to query updated item")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return response
	resp := models.ApiResponse{
		Status:  "success",
		Message: "Item updated successfully",
		Data:    item,
	}

	json.NewEncoder(w).Encode(resp)
}

// deleteItemHandler handles DELETE /api/v1/items/{id}
func deleteItemHandler(w http.ResponseWriter, r *http.Request, itemID int) {
	err := metrics.TrackDatabaseOperation("delete_item", func() error {
		result, err := database.DB.Exec("DELETE FROM items WHERE id = $1", itemID)
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Item not found", http.StatusNotFound)
		} else {
			logrus.WithError(err).Error("Failed to delete item")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return response
	resp := models.ApiResponse{
		Status:  "success",
		Message: "Item deleted successfully",
	}

	json.NewEncoder(w).Encode(resp)
}
