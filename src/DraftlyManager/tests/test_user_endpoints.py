import requests
import json
import pytest
from typing import Dict, Any

# Configuration
BASE_URL = "http://localhost:6060/v1"
HEADERS = {"Content-Type": "application/json"}

class TestUserEndpoints:
    """Test suite for User API endpoints"""
    
    def setup_method(self):
        """Setup method run before each test"""
        self.base_url = BASE_URL
        self.headers = HEADERS
        self.created_user_id = None
    
    def teardown_method(self):
        """Cleanup method run after each test"""
        # Clean up created user if exists
        if self.created_user_id:
            try:
                requests.delete(f"{self.base_url}/users/{self.created_user_id}")
            except:
                pass  # Ignore cleanup errors
    
    def test_create_user_success(self):
        """Test successful user creation"""
        user_data = {
            "name": "John Doe",
            "email": "john.doe@example.com"
        }
        
        response = requests.post(
            f"{self.base_url}/users",
            headers=self.headers,
            json=user_data
        )
        
        assert response.status_code == 201
        
        response_data = response.json()
        assert "id" in response_data
        assert response_data["name"] == user_data["name"]
        assert response_data["email"] == user_data["email"]
        assert "created_at" in response_data
        assert "updated_at" in response_data
        
        # Store for cleanup
        self.created_user_id = response_data["id"]
    
    def test_create_user_invalid_json(self):
        """Test user creation with invalid JSON"""
        response = requests.post(
            f"{self.base_url}/users",
            headers=self.headers,
            data="invalid json"
        )
        
        assert response.status_code == 400
    
    def test_create_user_missing_fields(self):
        """Test user creation with missing required fields"""
        user_data = {
            "name": "John Doe"
            # Missing email
        }
        
        response = requests.post(
            f"{self.base_url}/users",
            headers=self.headers,
            json=user_data
        )
        
        # Should fail due to missing email
        assert response.status_code in [400, 500]
    
    def test_get_user_success(self):
        """Test successful user retrieval"""
        # First create a user
        user_data = {
            "name": "Jane Smith",
            "email": "jane.smith@example.com"
        }
        
        create_response = requests.post(
            f"{self.base_url}/users",
            headers=self.headers,
            json=user_data
        )
        
        assert create_response.status_code == 201
        user_id = create_response.json()["id"]
        self.created_user_id = user_id
        
        # Now get the user
        response = requests.get(f"{self.base_url}/users/{user_id}")
        
        assert response.status_code == 200
        
        response_data = response.json()
        assert response_data["id"] == user_id
        assert response_data["name"] == user_data["name"]
        assert response_data["email"] == user_data["email"]
    
    def test_get_user_not_found(self):
        """Test getting non-existent user"""
        response = requests.get(f"{self.base_url}/users/99999")
        
        assert response.status_code == 404
    
    def test_get_user_invalid_id(self):
        """Test getting user with invalid ID"""
        response = requests.get(f"{self.base_url}/users/invalid")
        
        assert response.status_code == 400
    
    def test_update_user_success(self):
        """Test successful user update"""
        # First create a user
        user_data = {
            "name": "Bob Johnson",
            "email": "bob.johnson@example.com"
        }
        
        create_response = requests.post(
            f"{self.base_url}/users",
            headers=self.headers,
            json=user_data
        )
        
        assert create_response.status_code == 201
        user_id = create_response.json()["id"]
        self.created_user_id = user_id
        
        # Update the user
        update_data = {
            "name": "Bob Johnson Updated",
            "email": "bob.updated@example.com"
        }
        
        response = requests.put(
            f"{self.base_url}/users/{user_id}",
            headers=self.headers,
            json=update_data
        )
        
        assert response.status_code == 200
        
        response_data = response.json()
        assert response_data["id"] == user_id
        assert response_data["name"] == update_data["name"]
        assert response_data["email"] == update_data["email"]
    
    def test_update_user_not_found(self):
        """Test updating non-existent user"""
        update_data = {
            "name": "Non Existent",
            "email": "nonexistent@example.com"
        }
        
        response = requests.put(
            f"{self.base_url}/users/99999",
            headers=self.headers,
            json=update_data
        )
        
        assert response.status_code == 404
    
    def test_update_user_invalid_json(self):
        """Test updating user with invalid JSON"""
        response = requests.put(
            f"{self.base_url}/users/1",
            headers=self.headers,
            data="invalid json"
        )
        
        assert response.status_code == 400
    
    def test_delete_user_success(self):
        """Test successful user deletion"""
        # First create a user
        user_data = {
            "name": "Delete Me",
            "email": "delete.me@example.com"
        }
        
        create_response = requests.post(
            f"{self.base_url}/users",
            headers=self.headers,
            json=user_data
        )
        
        assert create_response.status_code == 201
        user_id = create_response.json()["id"]
        
        # Delete the user
        response = requests.delete(f"{self.base_url}/users/{user_id}")
        
        assert response.status_code == 204
        
        # Verify user is deleted
        get_response = requests.get(f"{self.base_url}/users/{user_id}")
        assert get_response.status_code == 404
        
        # Don't need cleanup since user is deleted
        self.created_user_id = None
    
    def test_delete_user_not_found(self):
        """Test deleting non-existent user"""
        response = requests.delete(f"{self.base_url}/users/99999")
        
        assert response.status_code == 404
    
    def test_delete_user_invalid_id(self):
        """Test deleting user with invalid ID"""
        response = requests.delete(f"{self.base_url}/users/invalid")
        
        assert response.status_code == 400
    
    def test_user_crud_workflow(self):
        """Test complete CRUD workflow"""
        # Create
        user_data = {
            "name": "CRUD Test User",
            "email": "crud.test@example.com"
        }
        
        create_response = requests.post(
            f"{self.base_url}/users",
            headers=self.headers,
            json=user_data
        )
        
        assert create_response.status_code == 201
        user_id = create_response.json()["id"]
        self.created_user_id = user_id
        
        # Read
        get_response = requests.get(f"{self.base_url}/users/{user_id}")
        assert get_response.status_code == 200
        assert get_response.json()["name"] == user_data["name"]
        
        # Update
        update_data = {
            "name": "CRUD Test User Updated",
            "email": "crud.updated@example.com"
        }
        
        update_response = requests.put(
            f"{self.base_url}/users/{user_id}",
            headers=self.headers,
            json=update_data
        )
        
        assert update_response.status_code == 200
        assert update_response.json()["name"] == update_data["name"]
        
        # Delete
        delete_response = requests.delete(f"{self.base_url}/users/{user_id}")
        assert delete_response.status_code == 204
        
        # Verify deletion
        final_get_response = requests.get(f"{self.base_url}/users/{user_id}")
        assert final_get_response.status_code == 404
        
        self.created_user_id = None


def test_api_health():
    """Test if the API is running"""
    try:
        response = requests.get(f"{BASE_URL.replace('/v1', '')}/health", timeout=5)
        assert response.status_code == 200
    except requests.exceptions.RequestException:
        pytest.skip("API server is not running")


if __name__ == "__main__":
    # Run tests when script is executed directly
    pytest.main([__file__, "-v"])