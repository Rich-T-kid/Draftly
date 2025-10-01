#!/usr/bin/env python3
"""
Test suite for Document API endpoints with operations
"""

import requests
import json
import os
import pytest
import time
import psycopg2
from typing import Dict, Any, List
from dotenv import load_dotenv

load_dotenv()

BASE_URL = "http://localhost:6060/v1"
HEADERS = {"Content-Type": "application/json"}

DB_CONFIG = {
    "host": os.getenv("POSTGRESS_HOST"),
    "port": os.getenv("POSTGRESS_PORT"),
    "user": os.getenv("POSTGRESS_USER"),
    "password": os.getenv("POSTGRESS_PASSWORD"),
    "dbname": os.getenv("POSTGRESS_DB_NAME")
}


class TestDocumentEndpoints:
    """Test suite for Document API endpoints"""
    
    def setup_method(self):
        self.base_url = BASE_URL
        self.headers = HEADERS
        self.created_user_ids = []
        self.created_document_id = None
        self.db_conn = None
        self.operations_dir = "operations"
        self.timestamp = str(int(time.time() * 1000))
        
        try:
            self.db_conn = psycopg2.connect(**DB_CONFIG)
        except Exception:
            self.db_conn = None
        
    def teardown_method(self):
        if self.created_document_id and self.created_user_ids:
            try:
                requests.delete(f"{self.base_url}/documents/{self.created_user_ids[0]}/{self.created_document_id}")
            except Exception:
                pass
                
        for user_id in self.created_user_ids:
            try:
                requests.delete(f"{self.base_url}/users/{user_id}")
            except Exception:
                pass
        
        if self.db_conn:
            self.db_conn.close()
        
        self.created_user_ids = []
        self.created_document_id = None
    
    def create_test_user(self, name="Test User", email_prefix="test"):
        unique_email = f"{email_prefix}.{self.timestamp}.{len(self.created_user_ids)}@example.com"
        user_data = {"name": name, "email": unique_email}
        
        response = requests.post(f"{self.base_url}/users", headers=self.headers, json=user_data)
        assert response.status_code == 201
        user = response.json()
        
        self.created_user_ids.append(user["id"])
        return user
    
    def load_operations(self, filename):
        try:
            with open(os.path.join(self.operations_dir, filename), 'r') as f:
                return json.load(f)
        except FileNotFoundError:
            return {"operations": []}
    
    def add_operations_to_db(self, document_id: int, operations: List[Dict[str, Any]]):
        if not self.db_conn:
            return False
            
        try:
            cursor = self.db_conn.cursor()
            operations_json = json.dumps(operations)
            cursor.execute(
                'UPDATE "Documents" SET operations = %s WHERE id = %s',
                (operations_json, document_id)
            )
            self.db_conn.commit()
            cursor.close()
            return True
        except Exception:
            if self.db_conn:
                self.db_conn.rollback()
            return False
    
    def get_document_from_db(self, document_id: int) -> Dict[str, Any]:
        if not self.db_conn:
            return None
            
        try:
            cursor = self.db_conn.cursor()
            cursor.execute(
                'SELECT id, user_id, title, operations, s3_key FROM "Documents" WHERE id = %s',
                (document_id,)
            )
            result = cursor.fetchone()
            cursor.close()
            
            if result:
                return {
                    "id": result[0],
                    "user_id": result[1],
                    "title": result[2],
                    "operations": result[3],
                    "s3_key": result[4]
                }
            return None
        except Exception:
            return None
    
    def test_create_document_success(self):
        user = self.create_test_user("Doc Creator", "doc.creator")
        
        doc_data = {
            "title": "Test Document",
            "userId": user["id"]
        }
        
        response = requests.post(
            f"{self.base_url}/documents/{user['id']}",
            headers=self.headers,
            json=doc_data
        )
        
        assert response.status_code == 201
        self.created_document_id = response.json()["id"]
    
    def test_create_document_with_permissions(self):
        owner = self.create_test_user("Owner", "owner")
        collaborator = self.create_test_user("Collaborator", "collaborator")
        
        doc_data = {
            "title": "Shared Document",
            "userId": owner["id"],
            "allowedUsers": [
                {"userId": collaborator["id"], "permission": "edit"}
            ]
        }
        
        response = requests.post(
            f"{self.base_url}/documents/{owner['id']}",
            headers=self.headers,
            json=doc_data
        )
        
        assert response.status_code == 201
        self.created_document_id = response.json()["id"]
    
    def test_get_user_documents(self):
        user = self.create_test_user("Multi Doc User", "multi")
        
        doc1_data = {"title": "First Document", "userId": user["id"]}
        doc2_data = {"title": "Second Document", "userId": user["id"]}
        
        doc1_response = requests.post(f"{self.base_url}/documents/{user['id']}", headers=self.headers, json=doc1_data)
        doc2_response = requests.post(f"{self.base_url}/documents/{user['id']}", headers=self.headers, json=doc2_data)
        
        assert doc1_response.status_code == 201
        assert doc2_response.status_code == 201
        
        self.created_document_id = doc1_response.json()["id"]
        doc2_id = doc2_response.json()["id"]
        
        response = requests.get(f"{self.base_url}/documents/{user['id']}")
        
        assert response.status_code == 200
        docs = response.json()
        assert len(docs) >= 2
        
        try:
            requests.delete(f"{self.base_url}/documents/{user['id']}/{doc2_id}")
        except Exception:
            pass
    
    def test_get_document_success(self):
        user = self.create_test_user("Doc Reader", "reader")
        
        doc_data = {"title": "Readable Document", "userId": user["id"]}
        create_response = requests.post(f"{self.base_url}/documents/{user['id']}", headers=self.headers, json=doc_data)
        
        assert create_response.status_code == 201
        doc = create_response.json()
        self.created_document_id = doc["id"]
        
        time.sleep(0.1)
        
        response = requests.get(f"{self.base_url}/documents/{user['id']}/{doc['id']}")
        
        assert response.status_code == 200
    
    def test_update_document_metadata(self):
        user = self.create_test_user("Doc Updater", "updater")
        collaborator = self.create_test_user("Collaborator", "collab.update")
        
        doc_data = {"title": "Original Title", "userId": user["id"]}
        create_response = requests.post(f"{self.base_url}/documents/{user['id']}", headers=self.headers, json=doc_data)
        
        assert create_response.status_code == 201
        doc = create_response.json()
        self.created_document_id = doc["id"]
        
        time.sleep(0.1)
        
        update_data = {
            "title": "Updated Title", 
            "userId": user["id"],
            "allowedUsers": [
                {"userId": collaborator["id"], "permission": "view-only"}
            ]
        }
        
        response = requests.put(
            f"{self.base_url}/documents/{user['id']}/{doc['id']}",
            headers=self.headers,
            json=update_data
        )
        
        assert response.status_code == 200
    
    def test_delete_document_success(self):
        user = self.create_test_user("Doc Deleter", "deleter")
        
        doc_data = {"title": "To Be Deleted", "userId": user["id"]}
        create_response = requests.post(f"{self.base_url}/documents/{user['id']}", headers=self.headers, json=doc_data)
        
        assert create_response.status_code == 201
        doc = create_response.json()
        
        time.sleep(0.1)
        
        response = requests.delete(f"{self.base_url}/documents/{user['id']}/{doc['id']}")
        
        assert response.status_code == 204
        self.created_document_id = None
    
    def test_document_not_found(self):
        user = self.create_test_user("Not Found User", "notfound")
        response = requests.get(f"{self.base_url}/documents/{user['id']}/99999")
        
        assert response.status_code == 404
    
    def test_operations_from_json_files(self):
        if not self.db_conn:
            pytest.skip("Database connection required for operations tests")
        
        user = self.create_test_user("JSON Ops User", "jsonops")
        
        doc_data = {"title": "JSON Operations Document", "userId": user["id"]}
        create_response = requests.post(f"{self.base_url}/documents/{user['id']}", headers=self.headers, json=doc_data)
        assert create_response.status_code == 201
        doc = create_response.json()
        self.created_document_id = doc["id"]
        
        operation_files = ["empty.json", "insert.json", "delete.json"]
        
        for op_file in operation_files:
            ops_data = self.load_operations(op_file)
            
            if "operations" in ops_data:
                operations = ops_data["operations"]
            elif isinstance(ops_data, list):
                operations = ops_data
            else:
                operations = []
            
            if operations:
                self.add_operations_to_db(doc["id"], operations)
                requests.put(f"{self.base_url}/documents/{doc['id']}")
    
    def test_operations_cleared_after_processing(self):
        if not self.db_conn:
            pytest.skip("Database connection required for operations tests")
        
        user = self.create_test_user("Clear Ops User", "clearops")
        
        doc_data = {"title": "Clear Operations Document", "userId": user["id"]}
        create_response = requests.post(f"{self.base_url}/documents/{user['id']}", headers=self.headers, json=doc_data)
        assert create_response.status_code == 201
        doc = create_response.json()
        self.created_document_id = doc["id"]
        
        operations = [{"type": "insert", "position": 0, "text": "test", "length": 0}]
        self.add_operations_to_db(doc["id"], operations)
        
        doc_before = self.get_document_from_db(doc["id"])
        assert doc_before["operations"] is not None
        
        process_response = requests.put(f"{self.base_url}/documents/{doc['id']}")
        assert process_response.status_code == 200
        
        time.sleep(0.2)
        doc_after = self.get_document_from_db(doc["id"])
        operations_after = json.loads(doc_after["operations"]) if doc_after["operations"] else []
        
        assert len(operations_after) == 0


def test_api_health():
    try:
        response = requests.get(f"{BASE_URL.replace('/v1', '')}/health", timeout=5)
        assert response.status_code == 200
    except requests.exceptions.RequestException:
        pytest.skip("API server is not running")


if __name__ == "__main__":
    pytest.main([__file__, "-v", "-s"])