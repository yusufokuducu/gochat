from typing import Optional
from pydantic import BaseModel

from app.models.friendship import FriendshipStatus

# Shared properties
class FriendshipBase(BaseModel):
    user_id: int
    friend_id: int

# Properties to receive via API on creation
class FriendshipCreate(FriendshipBase):
    pass

# Properties to receive via API on update
class FriendshipUpdate(BaseModel):
    status: FriendshipStatus

# User basic info for friendship
class UserInfo(BaseModel):
    id: int
    username: str
    email: str
    
    class Config:
        orm_mode = True

# Properties to return via API
class Friendship(FriendshipBase):
    id: int
    status: FriendshipStatus
    user: Optional[UserInfo] = None
    friend: Optional[UserInfo] = None

    class Config:
        orm_mode = True

# Friend request response
class FriendRequestResponse(BaseModel):
    request_id: int
    status: FriendshipStatus
