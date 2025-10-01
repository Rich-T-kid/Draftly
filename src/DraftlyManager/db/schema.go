package db

// User table queries
const (
	CreateUserQuery = `
		INSERT INTO "Users" (name, email, created_at, updated_at) 
		VALUES ($1, $2, NOW(), NOW()) 
		RETURNING id, name, email, created_at, updated_at`

	GetUserByIDQuery = `
		SELECT id, name, email, created_at, updated_at 
		FROM "Users" 
		WHERE id = $1`

	UpdateUserQuery = `
		UPDATE "Users" 
		SET name = $1, email = $2, updated_at = NOW() 
		WHERE id = $3 
		RETURNING id, name, email, created_at, updated_at`

	DeleteUserQuery = `
		DELETE FROM "Users" 
		WHERE id = $1`

	GetAllUsersQuery = `
		SELECT id, name, email, created_at, updated_at 
		FROM "Users" 
		ORDER BY created_at DESC`

	GetUserByEmailQuery = `
		SELECT id, name, email, created_at, updated_at 
		FROM "Users" 
		WHERE email = $1`
)

// Document table queries
const (
	CreateDocumentQuery = `
		INSERT INTO "Documents" (user_id, title, created_at, updated_at) 
		VALUES ($1, $2, NOW(), NOW()) 
		RETURNING id, user_id, title, created_at, updated_at`

	GetUserDocumentsQuery = `
		SELECT id, user_id, title, created_at, updated_at 
		FROM "Documents" 
		WHERE user_id = $1 
		ORDER BY updated_at DESC`

	GetDocumentQuery = `
		SELECT id, user_id, title, operations, s3_key, created_at, updated_at 
		FROM "Documents" 
		WHERE id = $1 AND user_id = $2`

	GetDocumentByIDQuery = `
		SELECT id, user_id, title, operations, s3_key, created_at, updated_at 
		FROM "Documents" 
		WHERE id = $1`

	UpdateDocumentQuery = `
		UPDATE "Documents" 
		SET title = $1, updated_at = NOW() 
		WHERE id = $2 AND user_id = $3 
		RETURNING id, user_id, title, created_at, updated_at`

	DeleteDocumentQuery = `
		DELETE FROM "Documents" 
		WHERE id = $1 AND user_id = $2`

	UpdateDocumentS3KeyQuery = `
		UPDATE "Documents" 
		SET s3_key = $1, updated_at = NOW() 
		WHERE id = $2`

	GetDocumentOperationsQuery = `
		SELECT operations 
		FROM "Documents" 
		WHERE id = $1`

	UpdateDocumentOperationsQuery = `
		UPDATE "Documents" 
		SET operations = $1, updated_at = NOW() 
		WHERE id = $2`

	ClearDocumentOperationsQuery = `
		UPDATE "Documents" 
		SET operations = '[]'::json, updated_at = NOW() 
		WHERE id = $1`
)

// Operation table queries
const (
	CreateOperationQuery = `
		INSERT INTO "Operations" (document_id, type, position, text, length, created_at) 
		VALUES ($1, $2, $3, $4, $5, NOW()) 
		RETURNING id, document_id, type, position, text, length, created_at`

	GetDocumentOperationsListQuery = `
		SELECT id, document_id, type, position, text, length, created_at 
		FROM "Operations" 
		WHERE document_id = $1 
		ORDER BY created_at ASC`

	DeleteOperationQuery = `
		DELETE FROM "Operations" 
		WHERE id = $1`

	DeleteAllDocumentOperationsQuery = `
		DELETE FROM "Operations" 
		WHERE document_id = $1`

	GetOperationByIDQuery = `
		SELECT id, document_id, type, position, text, length, created_at 
		FROM "Operations" 
		WHERE id = $1`
)

// Permission table queries
const (
	CreatePermissionQuery = `
		INSERT INTO "DocumentPermissions" (document_id, user_id, permission, created_at, updated_at) 
		VALUES ($1, $2, $3, NOW(), NOW())`

	GetDocumentPermissionsQuery = `
		SELECT user_id, permission, created_at, updated_at 
		FROM "DocumentPermissions" 
		WHERE document_id = $1`

	GetUserPermissionsQuery = `
		SELECT document_id, permission, created_at, updated_at 
		FROM "DocumentPermissions" 
		WHERE user_id = $1`

	UpdatePermissionQuery = `
		UPDATE "DocumentPermissions" 
		SET permission = $1, updated_at = NOW() 
		WHERE document_id = $2 AND user_id = $3`

	DeletePermissionQuery = `
		DELETE FROM "DocumentPermissions" 
		WHERE document_id = $1 AND user_id = $2`

	DeleteAllDocumentPermissionsQuery = `
		DELETE FROM "DocumentPermissions" 
		WHERE document_id = $1`

	CheckUserPermissionQuery = `
		SELECT permission 
		FROM "DocumentPermissions" 
		WHERE document_id = $1 AND user_id = $2`

	GetUsersWithDocumentAccessQuery = `
		SELECT u.id, u.name, u.email, dp.permission 
		FROM "Users" u 
		JOIN "DocumentPermissions" dp ON u.id = dp.user_id 
		WHERE dp.document_id = $1`
)
