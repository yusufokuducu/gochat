import pytest
from fastapi import status

def test_read_users_me(client, token):
    """Test getting current user information"""
    response = client.get(
        "/api/users/me",
        headers={"Authorization": f"Bearer {token}"}
    )
    assert response.status_code == status.HTTP_200_OK
    assert response.json()["username"] == "testuser"
    assert response.json()["email"] == "test@example.com"

def test_read_users_me_unauthorized(client):
    """Test getting current user without token"""
    response = client.get("/api/users/me")
    assert response.status_code == status.HTTP_401_UNAUTHORIZED

def test_update_user_me(client, token):
    """Test updating current user information"""
    response = client.put(
        "/api/users/me",
        headers={"Authorization": f"Bearer {token}"},
        json={"username": "updateduser", "email": "updated@example.com"}
    )
    assert response.status_code == status.HTTP_200_OK
    assert response.json()["username"] == "updateduser"
    assert response.json()["email"] == "updated@example.com"

def test_update_user_password(client, token):
    """Test updating user password"""
    # Update password
    response = client.put(
        "/api/users/me",
        headers={"Authorization": f"Bearer {token}"},
        json={"password": "newpassword"}
    )
    assert response.status_code == status.HTTP_200_OK
    
    # Try logging in with new password
    response = client.post(
        "/api/auth/login",
        data={"username": "testuser", "password": "newpassword"}
    )
    assert response.status_code == status.HTTP_200_OK
    assert "access_token" in response.json()

def test_read_user_by_id(client, token, test_user):
    """Test getting user by ID"""
    response = client.get(
        f"/api/users/{test_user.id}",
        headers={"Authorization": f"Bearer {token}"}
    )
    assert response.status_code == status.HTTP_200_OK
    assert response.json()["username"] == "testuser"
    assert response.json()["email"] == "test@example.com"

def test_read_user_by_id_not_found(client, token):
    """Test getting nonexistent user by ID"""
    response = client.get(
        "/api/users/999",  # Nonexistent user ID
        headers={"Authorization": f"Bearer {token}"}
    )
    assert response.status_code == status.HTTP_404_NOT_FOUND

def test_read_users(client, token, test_user):
    """Test getting list of users"""
    response = client.get(
        "/api/users/",
        headers={"Authorization": f"Bearer {token}"}
    )
    assert response.status_code == status.HTTP_200_OK
    assert isinstance(response.json(), list)
    assert len(response.json()) >= 1  # At least the test user should be returned
