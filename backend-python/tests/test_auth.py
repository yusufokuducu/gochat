import pytest
from fastapi import status

def test_login_success(client, test_user):
    """Test successful login with correct credentials"""
    response = client.post(
        "/api/auth/login",
        data={"username": "testuser", "password": "password"}
    )
    assert response.status_code == status.HTTP_200_OK
    assert "access_token" in response.json()
    assert response.json()["token_type"] == "bearer"

def test_login_wrong_password(client, test_user):
    """Test login with wrong password"""
    response = client.post(
        "/api/auth/login",
        data={"username": "testuser", "password": "wrongpassword"}
    )
    assert response.status_code == status.HTTP_401_UNAUTHORIZED

def test_login_nonexistent_user(client):
    """Test login with nonexistent user"""
    response = client.post(
        "/api/auth/login",
        data={"username": "nonexistentuser", "password": "password"}
    )
    assert response.status_code == status.HTTP_401_UNAUTHORIZED

def test_register_success(client):
    """Test successful user registration"""
    response = client.post(
        "/api/auth/register",
        json={
            "username": "newuser",
            "email": "new@example.com",
            "password": "password123"
        }
    )
    assert response.status_code == status.HTTP_201_CREATED
    assert response.json()["username"] == "newuser"
    assert response.json()["email"] == "new@example.com"
    assert "id" in response.json()

def test_register_duplicate_username(client, test_user):
    """Test registration with duplicate username"""
    response = client.post(
        "/api/auth/register",
        json={
            "username": "testuser",  # Same as test_user
            "email": "another@example.com",
            "password": "password123"
        }
    )
    assert response.status_code == status.HTTP_400_BAD_REQUEST

def test_register_duplicate_email(client, test_user):
    """Test registration with duplicate email"""
    response = client.post(
        "/api/auth/register",
        json={
            "username": "anotheruser",
            "email": "test@example.com",  # Same as test_user
            "password": "password123"
        }
    )
    assert response.status_code == status.HTTP_400_BAD_REQUEST
