from typing import Optional
from datetime import datetime
from pydantic import BaseModel

# Shared properties
class MessageBase(BaseModel):
    content: str

# Properties to receive via API on creation
class MessageCreate(MessageBase):
    receiver_id: int

# Properties to receive via API on update
class MessageUpdate(BaseModel):
    is_read: bool = True

# Properties to return via API
class Message(MessageBase):
    id: int
    sender_id: int
    receiver_id: int
    timestamp: datetime
    is_read: bool

    class Config:
        orm_mode = True

# WebSocket message format
class WebSocketMessage(BaseModel):
    sender_id: int
    receiver_id: int
    content: str
    timestamp: Optional[datetime] = None
