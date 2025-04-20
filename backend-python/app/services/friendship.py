from typing import List, Optional
from sqlalchemy.orm import Session, joinedload

from app.models.friendship import Friendship, FriendshipStatus
from app.models.user import User
from app.schemas.friendship import FriendshipCreate, FriendshipUpdate

def get_friendship(db: Session, user_id: int, friend_id: int) -> Optional[Friendship]:
    return db.query(Friendship).filter(
        Friendship.user_id == user_id,
        Friendship.friend_id == friend_id
    ).first()

def get_friendships_by_user(db: Session, user_id: int, status: Optional[FriendshipStatus] = None) -> List[Friendship]:
    query = db.query(Friendship).filter(Friendship.user_id == user_id)
    if status:
        query = query.filter(Friendship.status == status)
    
    # Kullanıcı bilgilerini de yükle
    query = query.join(User, Friendship.friend_id == User.id)
    
    return query.all()

def get_friend_requests(db: Session, user_id: int) -> List[Friendship]:
    # Arkadaşlık isteklerini ve gönderen kullanıcı bilgilerini al
    return db.query(Friendship).filter(
        Friendship.friend_id == user_id,
        Friendship.status == FriendshipStatus.PENDING
    ).join(User, Friendship.user_id == User.id).all()

def create_friendship(db: Session, friendship_in: FriendshipCreate) -> Friendship:
    db_friendship = Friendship(
        user_id=friendship_in.user_id,
        friend_id=friendship_in.friend_id,
        status=FriendshipStatus.PENDING
    )
    db.add(db_friendship)
    db.commit()
    db.refresh(db_friendship)
    return db_friendship

def update_friendship_status(db: Session, friendship: Friendship, status: FriendshipStatus) -> Friendship:
    friendship.status = status
    db.add(friendship)
    db.commit()
    db.refresh(friendship)
    return friendship

def delete_friendship(db: Session, friendship: Friendship) -> None:
    db.delete(friendship)
    db.commit()
