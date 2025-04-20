from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from app.api.deps import get_current_user
from app.db.base import get_db
from app.models.user import User
from app.models.friendship import Friendship, FriendshipStatus
from app.schemas.message import Message as MessageSchema, MessageCreate
from app.services.message import (
    create_message, get_message_by_id, get_messages_between_users,
    get_unread_messages, mark_message_as_read, mark_all_messages_as_read
)
from app.services.friendship import get_friendship
from app.services.user import get_user_by_id

router = APIRouter()

@router.post("/", response_model=MessageSchema, status_code=status.HTTP_201_CREATED)
async def send_message(
    message_in: MessageCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Send a message to another user
    """
    # Check if receiver exists
    receiver = get_user_by_id(db, message_in.receiver_id)
    if not receiver:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Receiver not found"
        )
    
    # Check if they are friends
    friendship = get_friendship(db, current_user.id, message_in.receiver_id)
    if not friendship or friendship.status != FriendshipStatus.ACCEPTED:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="You can only send messages to friends"
        )
    
    # Create the message
    message = create_message(db, message_in, current_user.id)
    return message

@router.get("/with/{user_id}", response_model=List[MessageSchema])
async def read_messages_with_user(
    user_id: int,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Get messages between current user and another user
    """
    # Check if other user exists
    other_user = get_user_by_id(db, user_id)
    if not other_user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="User not found"
        )
    
    # Get messages
    messages = get_messages_between_users(db, current_user.id, user_id, skip, limit)
    
    # Mark messages as read
    mark_all_messages_as_read(db, current_user.id, user_id)
    
    return messages

@router.get("/unread", response_model=List[MessageSchema])
async def read_unread_messages(
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Get all unread messages for the current user
    """
    messages = get_unread_messages(db, current_user.id)
    return messages

@router.put("/{message_id}/read", response_model=MessageSchema)
async def mark_message_read(
    message_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Mark a message as read
    """
    # Get the message
    message = get_message_by_id(db, message_id)
    if not message:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Message not found"
        )
    
    # Check if current user is the receiver
    if message.receiver_id != current_user.id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="You can only mark your own messages as read"
        )
    
    # Mark as read
    message = mark_message_as_read(db, message)
    return message
