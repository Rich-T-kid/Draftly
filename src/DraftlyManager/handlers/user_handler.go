package handlers

import (
	"Draftly/CRUD/db"
	"Draftly/CRUD/models"
	"Draftly/CRUD/services"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type UserHandler struct {
	dbService *services.DatabaseService
}

func NewUserHandler(dbService *services.DatabaseService) *UserHandler {
	return &UserHandler{
		dbService: dbService,
	}
}

// CreateUser handles POST /v1/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUG: CreateUser function called")

	var userInput models.UserInput
	if err := json.NewDecoder(r.Body).Decode(&userInput); err != nil {
		fmt.Printf("DEBUG: CreateUser JSON decode error: %v\n", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	fmt.Printf("DEBUG: CreateUser parsed input: %+v\n", userInput)

	// Add validation for required fields
	if userInput.Name == "" || userInput.Email == "" {
		fmt.Println("DEBUG: CreateUser validation failed - missing fields")
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	fmt.Println("DEBUG: CreateUser about to execute database query")
	var user models.User
	err := h.dbService.ExecuteQueryRow(db.CreateUserQuery,
		[]interface{}{&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt},
		userInput.Name, userInput.Email)

	if err != nil {
		fmt.Printf("DEBUG: CreateUser database error: %v\n", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}
	fmt.Printf("DEBUG: CreateUser success: %+v\n", user)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// GetUser handles GET /v1/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	results, err := h.dbService.ExecuteQuery(db.GetUserByIDQuery, id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if len(results) == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results[0])
}

// UpdateUser handles PUT /v1/users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var userInput models.UserInput
	if err := json.NewDecoder(r.Body).Decode(&userInput); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Add validation for required fields
	if userInput.Name == "" || userInput.Email == "" {
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	var user models.User
	err = h.dbService.ExecuteQueryRow(db.UpdateUserQuery,
		[]interface{}{&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt},
		userInput.Name, userInput.Email, id)

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// DeleteUser handles DELETE /v1/users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	rowsAffected, err := h.dbService.ExecuteNonQuery(db.DeleteUserQuery, id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
