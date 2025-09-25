# Draftly Database Schema

## Users
- **id**: INT, Primary Key, Auto Increment  
- **name**: VARCHAR(255), NOT NULL  
- **email**: VARCHAR(255), UNIQUE, NOT NULL  

---

## Documents
- **id**: INT, Primary Key, Auto Increment  
- **user_id**: INT, Foreign Key → Users(id), NOT NULL, ON DELETE CASCADE  
- **title**: VARCHAR(255), NOT NULL  

---

## Operations
- **id**: INT, Primary Key, Auto Increment  
- **document_id**: INT, Foreign Key → Documents(id), NOT NULL, ON DELETE CASCADE  
- **type**: ENUM('insert', 'delete'), NOT NULL  
- **position**: INT, NOT NULL  
- **text**: TEXT, NOT NULL  
- **length**: INT, NOT NULL  

**Constraint:**  
- If `type = 'insert'` → `text` must be non-empty and `length = 0`  
- If `type = 'delete'` → `text` must be empty and `length > 0`  

---

## DocumentPermissions
- **id**: INT, Primary Key, Auto Increment  
- **document_id**: INT, Foreign Key → Documents(id), NOT NULL, ON DELETE CASCADE  
- **user_id**: INT, Foreign Key → Users(id), NOT NULL, ON DELETE CASCADE  
- **permission**: ENUM('edit', 'view-only'), NOT NULL  

**Unique Constraint:** `(document_id, user_id)`  

**Trigger:** Document owner always has `'edit'` permission.  
