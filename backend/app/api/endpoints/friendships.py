from typing import List
from fastapi import APIRouter, Depends, HTTPException, status, Query, Body
from sqlalchemy.orm import Session
import logging

from app.api.deps import get_current_user
from app.db.base import get_db
from app.models.user import User
from app.models.friendship import Friendship, FriendshipStatus
from app.schemas.friendship import Friendship as FriendshipSchema, FriendshipCreate, FriendRequestResponse
from app.schemas.user import User as UserSchema
from app.services.friendship import (
    create_friendship, get_friendship, get_friendships_by_user,
    get_friend_requests, update_friendship_status, delete_friendship
)
from app.services.user import get_user_by_id

router = APIRouter()
logger = logging.getLogger(__name__)

@router.post("/", response_model=FriendshipSchema, status_code=status.HTTP_201_CREATED)
async def add_friend(
    friend_id: int = Query(..., description="ID of the user to add as friend"),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Send friend request
    """
    logger.info(f"Received friend request from user {current_user.id} to user {friend_id}")
    
    # Check if friend exists
    friend = get_user_by_id(db, friend_id)
    if not friend:
        logger.error(f"Friend with ID {friend_id} not found")
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="User not found"
        )
    
    # Check if trying to add self
    if friend_id == current_user.id:
        logger.error(f"User {current_user.id} tried to add themselves as a friend")
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Cannot add yourself as a friend"
        )
    
    # Check if friendship already exists
    existing_friendship = get_friendship(db, current_user.id, friend_id)
    if existing_friendship:
        logger.error(f"Friendship already exists between {current_user.id} and {friend_id} with status: {existing_friendship.status}")
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Friendship already exists with status: {existing_friendship.status}"
        )
    
    # Create friendship request
    try:
        friendship_in = FriendshipCreate(user_id=current_user.id, friend_id=friend_id)
        friendship = create_friendship(db, friendship_in)
        
        # Kullanıcı bilgilerini ekle
        friendship_response = FriendshipSchema.from_orm(friendship)
        friendship_response.user = UserSchema.from_orm(current_user)
        friendship_response.friend = UserSchema.from_orm(friend)
        
        logger.info(f"Friendship created successfully between {current_user.id} and {friend_id}")
        return friendship_response
    except Exception as e:
        logger.error(f"Error creating friendship: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to create friendship"
        )

@router.get("/", response_model=List[FriendshipSchema])
async def read_friends(
    status: FriendshipStatus = None,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Get list of friendships
    """
    friendships = get_friendships_by_user(db, current_user.id, status)
    
    # Kullanıcı bilgilerini ekle
    result = []
    for friendship in friendships:
        friendship_data = FriendshipSchema.from_orm(friendship)
        friendship_data.user = UserSchema.from_orm(current_user)
        
        friend = get_user_by_id(db, friendship.friend_id)
        if friend:
            friendship_data.friend = UserSchema.from_orm(friend)
        
        result.append(friendship_data)
    
    return result

@router.get("/requests", response_model=List[FriendshipSchema])
async def read_friend_requests(
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Get list of pending friend requests
    """
    requests = get_friend_requests(db, current_user.id)
    
    # Kullanıcı bilgilerini ekle
    result = []
    for request in requests:
        request_data = FriendshipSchema.from_orm(request)
        request_data.friend = UserSchema.from_orm(current_user)
        
        user = get_user_by_id(db, request.user_id)
        if user:
            request_data.user = UserSchema.from_orm(user)
        
        result.append(request_data)
    
    return result

@router.put("/requests/{request_id}", response_model=FriendRequestResponse)
async def respond_to_friend_request(
    request_id: int,
    status_data: dict = Body(...),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Accept or reject a friend request
    """
    # Status değerini al
    status_str = status_data.get("status")
    logger.info(f"User {current_user.id} responding to friend request {request_id} with status {status_str}")
    
    # Status değerini doğrula ve dönüştür
    try:
        status_value = FriendshipStatus(status_str)
    except ValueError:
        logger.error(f"Invalid status value: {status_str}")
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Invalid status value: {status_str}. Must be one of: {[s.value for s in FriendshipStatus]}"
        )
    
    # Find the friend request
    friendship = db.query(Friendship).filter(
        Friendship.id == request_id,
        Friendship.friend_id == current_user.id,
        Friendship.status == FriendshipStatus.PENDING
    ).first()
    
    if not friendship:
        logger.error(f"Friend request {request_id} not found for user {current_user.id}")
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Friend request not found"
        )
    
    # Update the status
    if status_value not in [FriendshipStatus.ACCEPTED, FriendshipStatus.REJECTED, FriendshipStatus.BLOCKED]:
        logger.error(f"Invalid status {status_value} for friend request response")
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Invalid status for friend request response"
        )
    
    try:
        friendship = update_friendship_status(db, friendship, status_value)
        
        # If accepted, create reverse friendship
        if status_value == FriendshipStatus.ACCEPTED:
            reverse_friendship = get_friendship(db, current_user.id, friendship.user_id)
            if not reverse_friendship:
                friendship_in = FriendshipCreate(user_id=current_user.id, friend_id=friendship.user_id)
                reverse_friendship = create_friendship(db, friendship_in)
                reverse_friendship = update_friendship_status(db, reverse_friendship, FriendshipStatus.ACCEPTED)
        
        logger.info(f"Friend request {request_id} updated to status {status_value}")
        return {"request_id": friendship.id, "status": friendship.status}
    except Exception as e:
        logger.error(f"Error updating friend request: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to update friend request"
        )

@router.delete("/{friend_id}", status_code=status.HTTP_204_NO_CONTENT)
async def remove_friend(
    friend_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Remove a friend
    """
    # Check if friendship exists
    friendship = get_friendship(db, current_user.id, friend_id)
    if not friendship:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Friendship not found"
        )
    
    # Delete the friendship
    delete_friendship(db, friendship)
    
    # Also delete reverse friendship if it exists
    reverse_friendship = get_friendship(db, friend_id, current_user.id)
    if reverse_friendship:
        delete_friendship(db, reverse_friendship)
    
    return None
