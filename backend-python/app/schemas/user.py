from typing import Optional
from datetime import datetime
from pydantic import BaseModel, EmailStr, Field

# Shared properties
class UserBase(BaseModel):
    username: str
    email: EmailStr

# Properties to receive via API on creation
class UserCreate(UserBase):
    password: str

# Properties to receive via API on update
class UserUpdate(BaseModel):
    username: Optional[str] = None
    email: Optional[EmailStr] = None
    password: Optional[str] = None

# Properties to return via API
class User(UserBase):
    id: int
    created_at: datetime

    class Config:
        orm_mode = True

# Properties for token
class Token(BaseModel):
    access_token: str
    token_type: str

class TokenPayload(BaseModel):
    sub: Optional[int] = None

# Login request
class UserLogin(BaseModel):
    username: str
    password: str
