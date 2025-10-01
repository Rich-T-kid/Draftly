package handlers

import (
	"Draftly/CRUD/db"
	"Draftly/CRUD/models"
	"Draftly/CRUD/services"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type DocumentHandler struct {
	dbService *services.DatabaseService
	s3Service *services.S3Service
}

func NewDocumentHandler(dbService *services.DatabaseService, s3Service *services.S3Service) *DocumentHandler {
	return &DocumentHandler{
		dbService: dbService,
		s3Service: s3Service,
	}
}

// CreateDocument handles POST /v1/documents/{userId}
func (h *DocumentHandler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUG: CreateDocument function called")

	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil {
		fmt.Printf("DEBUG: Invalid user ID error: %v\n", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	fmt.Printf("DEBUG: Parsed userID: %d\n", userID)

	var docInput models.DocumentInput
	if err := json.NewDecoder(r.Body).Decode(&docInput); err != nil {
		fmt.Printf("DEBUG: JSON decode error: %v\n", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	fmt.Printf("DEBUG: Parsed docInput: %+v\n", docInput)

	fmt.Println("DEBUG: About to execute database query")
	var doc models.Document
	err = h.dbService.ExecuteQueryRow(db.CreateDocumentQuery,
		[]interface{}{&doc.ID, &doc.UserID, &doc.Title, &doc.CreatedAt, &doc.UpdatedAt},
		userID, docInput.Title)

	if err != nil {
		fmt.Printf("DEBUG: Database error: %v\n", err)
		http.Error(w, "Failed to create document", http.StatusInternalServerError)
		return
	}
	fmt.Printf("DEBUG: Created document: %+v\n", doc)

	// Handle permissions if provided
	if len(docInput.AllowedUsers) > 0 {
		fmt.Printf("DEBUG: Processing %d permissions\n", len(docInput.AllowedUsers))
		for _, permission := range docInput.AllowedUsers {
			_, err := h.dbService.ExecuteNonQuery(db.CreatePermissionQuery,
				doc.ID, permission.UserID, permission.Permission)
			if err != nil {
				fmt.Printf("DEBUG: Permission error: %v\n", err)
				continue
			}
		}
	}

	fmt.Println("DEBUG: Sending successful response")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(doc)
}

// GetUserDocuments handles GET /v1/documents/{userId}
func (h *DocumentHandler) GetUserDocuments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	results, err := h.dbService.ExecuteQuery(db.GetUserDocumentsQuery, userID)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GetDocument handles GET /v1/documents/{userId}/{documentId}
func (h *DocumentHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	documentID, err := strconv.Atoi(vars["documentId"])
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	results, err := h.dbService.ExecuteQuery(db.GetDocumentQuery, documentID, userID)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if len(results) == 0 {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	doc := results[0]

	// Convert database field names to API field names
	response := map[string]interface{}{
		"id":         doc["id"],
		"userId":     doc["user_id"], // Convert user_id to userId
		"title":      doc["title"],
		"created_at": doc["created_at"],
		"updated_at": doc["updated_at"],
	}

	// Get S3 content if available
	if s3Key, exists := doc["s3_key"]; exists && s3Key != nil {
		content, err := h.s3Service.DownloadDocument(s3Key.(string))
		if err == nil {
			response["content"] = string(content)
		}
	}

	// Get operations if they exist
	if operationsData, exists := doc["operations"]; exists && operationsData != nil {
		var operations []models.Operation
		if err := json.Unmarshal([]byte(operationsData.(string)), &operations); err == nil {
			response["operations"] = operations
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateDocument handles PUT /v1/documents/{userId}/{documentId}
func (h *DocumentHandler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	documentID, err := strconv.Atoi(vars["documentId"])
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	var docInput models.DocumentInput
	if err := json.NewDecoder(r.Body).Decode(&docInput); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var doc models.Document
	err = h.dbService.ExecuteQueryRow(db.UpdateDocumentQuery,
		[]interface{}{&doc.ID, &doc.UserID, &doc.Title, &doc.CreatedAt, &doc.UpdatedAt},
		docInput.Title, documentID, userID)

	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Update permissions if provided
	if len(docInput.AllowedUsers) > 0 {
		// Get existing permissions
		existingResults, err := h.dbService.ExecuteQuery(db.GetDocumentPermissionsQuery, documentID)
		if err != nil {
			// Log error but continue
		}

		// Create map of existing permissions
		existingPerms := make(map[int64]string)
		for _, result := range existingResults {
			userID := result["user_id"].(int64)
			permission := result["permission"].(string)
			existingPerms[userID] = permission
		}

		// Create map of new permissions
		newPerms := make(map[int64]string)
		for _, permission := range docInput.AllowedUsers {
			newPerms[int64(permission.UserID)] = permission.Permission
		}

		// Add or update permissions
		for userID, permission := range newPerms {
			if existingPerm, exists := existingPerms[userID]; exists {
				// Update if permission changed
				if existingPerm != permission {
					h.dbService.ExecuteNonQuery(db.UpdatePermissionQuery, permission, documentID, userID)
				}
			} else {
				// Add new permission
				h.dbService.ExecuteNonQuery(db.CreatePermissionQuery, documentID, userID, permission)
			}
		}

		// Remove permissions that are no longer in the new list
		for userID := range existingPerms {
			if _, exists := newPerms[userID]; !exists {
				h.dbService.ExecuteNonQuery(db.DeletePermissionQuery, documentID, userID)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

// DeleteDocument handles DELETE /v1/documents/{userId}/{documentId}
func (h *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	documentID, err := strconv.Atoi(vars["documentId"])
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Get S3 key before deleting
	results, err := h.dbService.ExecuteQuery(db.GetDocumentQuery, documentID, userID)
	if err == nil && len(results) > 0 {
		if s3Key, exists := results[0]["s3_key"]; exists && s3Key != nil {
			// Delete from S3
			h.s3Service.DeleteDocument(s3Key.(string))
		}
	}

	rowsAffected, err := h.dbService.ExecuteNonQuery(db.DeleteDocumentQuery, documentID, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateDocumentContent handles PUT /v1/documents/{documentId}
func (h *DocumentHandler) UpdateDocumentContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	documentID, err := strconv.Atoi(vars["documentId"])
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	fmt.Printf("DEBUG: UpdateDocumentContent called for document ID: %d\n", documentID)

	// Read request body (but we're not using it for operations processing)
	_, err = io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Check if document exists
	results, err := h.dbService.ExecuteQuery(db.GetDocumentByIDQuery, documentID)
	if err != nil || len(results) == 0 {
		fmt.Printf("DEBUG: Document not found or query error: %v\n", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	fmt.Printf("DEBUG: Document found in database\n")

	// Get current content from S3 (if exists)
	var currentContent string
	if s3Key, exists := results[0]["s3_key"]; exists && s3Key != nil {
		fmt.Printf("DEBUG: Found s3_key: %v\n", s3Key)
		s3Content, err := h.s3Service.DownloadDocument(s3Key.(string))
		if err != nil {
			fmt.Printf("DEBUG: Error downloading from S3: %v\n", err)
			currentContent = ""
		} else {
			currentContent = string(s3Content)
			fmt.Printf("DEBUG: Downloaded content from S3, length: %d\n", len(currentContent))
			fmt.Printf("DEBUG: Content: '%s'\n", currentContent)
		}
	} else {
		fmt.Println("DEBUG: No s3_key found, starting with empty content")
	}

	// Get operations from database
	var operations []models.Operation
	if operationsData, exists := results[0]["operations"]; exists && operationsData != nil {
		operationsStr := operationsData.(string)
		fmt.Printf("DEBUG: Operations from DB: %s\n", operationsStr)

		if err := json.Unmarshal([]byte(operationsStr), &operations); err != nil {
			fmt.Printf("DEBUG: Error parsing operations: %v\n", err)
			http.Error(w, "Failed to parse operations", http.StatusInternalServerError)
			return
		}
		fmt.Printf("DEBUG: Parsed %d operations\n", len(operations))
	} else {
		fmt.Println("DEBUG: No operations found in database")
	}

	// Apply operations to content (forward order)
	documentContent := currentContent
	fmt.Printf("DEBUG: Starting content: '%s'\n", documentContent)

	for i := 0; i < len(operations); i++ {
		operation := operations[i]
		fmt.Printf("DEBUG: Applying operation %d: type=%s, pos=%d, text='%s', len=%d\n",
			i, operation.Type, operation.Position, operation.Text, operation.Length)

		beforeContent := documentContent
		documentContent = h.applyOperation(documentContent, operation)
		fmt.Printf("DEBUG: Content changed from '%s' to '%s'\n", beforeContent, documentContent)
	}

	fmt.Printf("DEBUG: Final content: '%s'\n", documentContent)

	// Upload final content to S3
	s3Key, err := h.s3Service.UploadDocument(documentID, []byte(documentContent))
	if err != nil {
		fmt.Printf("DEBUG: Error uploading to S3: %v\n", err)
		http.Error(w, "Failed to upload to S3", http.StatusInternalServerError)
		return
	}
	fmt.Printf("DEBUG: Uploaded to S3 with key: %s\n", s3Key)

	// Update database with S3 key
	_, err = h.dbService.ExecuteNonQuery(db.UpdateDocumentS3KeyQuery, s3Key, documentID)
	if err != nil {
		fmt.Printf("DEBUG: Error updating S3 key in DB: %v\n", err)
		http.Error(w, "Failed to update document", http.StatusInternalServerError)
		return
	}

	// Clear operations from database after successful S3 upload
	_, err = h.dbService.ExecuteNonQuery(db.ClearDocumentOperationsQuery, documentID)
	if err != nil {
		fmt.Printf("DEBUG: Error clearing operations: %v\n", err)
		http.Error(w, "Failed to clear operations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Document updated successfully in S3"})
}

// applyOperation applies a single operation to the document content
func (h *DocumentHandler) applyOperation(content string, operation models.Operation) string {
	contentRunes := []rune(content)

	switch operation.Type {
	case "insert":
		if operation.Position >= 0 && operation.Position <= len(contentRunes) {
			result := string(contentRunes[:operation.Position]) + operation.Text + string(contentRunes[operation.Position:])
			return result
		}
	case "delete":
		start := operation.Position
		end := start + operation.Length
		if start >= 0 && start <= len(contentRunes) && end <= len(contentRunes) {
			return string(contentRunes[:start]) + string(contentRunes[end:])
		}
	}
	return content
}
