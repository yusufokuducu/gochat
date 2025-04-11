import os
import pytest
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from fastapi.testclient import TestClient
from sqlalchemy.ext.declarative import declarative_base

from app.main import app
from app.db.base import Base, get_db
from app.core.config import settings
from app.core.security import create_access_token
from app.models.user import User

# Use an in-memory SQLite database for testing
SQLALCHEMY_DATABASE_URL = "sqlite:///./test.db"

engine = create_engine(
    SQLALCHEMY_DATABASE_URL, connect_args={"check_same_thread": False}
)
TestingSessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

@pytest.fixture(scope="function")
def db():
    # Create the database tables
    Base.metadata.create_all(bind=engine)
    
    # Create a new session for each test
    db = TestingSessionLocal()
    try:
        yield db
    finally:
        db.close()
        
    # Drop all tables after the test
    Base.metadata.drop_all(bind=engine)

@pytest.fixture(scope="function")
def client(db):
    # Override the get_db dependency to use the test database
    def override_get_db():
        try:
            yield db
        finally:
            db.close()
    
    app.dependency_overrides[get_db] = override_get_db
    
    with TestClient(app) as c:
        yield c
    
    # Remove the override after the test
    app.dependency_overrides = {}

@pytest.fixture(scope="function")
def test_user(db):
    # Create a test user
    user = User(
        username="testuser",
        email="test@example.com",
        password_hash="$2b$12$tVN1BzXJVS8rlZ6xnrYNWuQllADJY.Fg2r/GZiQX9Jb1ZzJg.EQdq"  # "password"
    )
    db.add(user)
    db.commit()
    db.refresh(user)
    return user

@pytest.fixture(scope="function")
def token(test_user):
    # Create a token for the test user
    return create_access_token(test_user.id)
