from typing import List, Optional
from sqlalchemy.orm import Session
from sqlalchemy import or_, and_

from app.models.message import Message
from app.schemas.message import MessageCreate, MessageUpdate

def get_message_by_id(db: Session, message_id: int) -> Optional[Message]:
    return db.query(Message).filter(Message.id == message_id).first()

def get_messages_between_users(
    db: Session, 
    user_id: int, 
    other_user_id: int, 
    skip: int = 0, 
    limit: int = 100
) -> List[Message]:
    return db.query(Message).filter(
        or_(
            and_(Message.sender_id == user_id, Message.receiver_id == other_user_id),
            and_(Message.sender_id == other_user_id, Message.receiver_id == user_id)
        )
    ).order_by(Message.timestamp.desc()).offset(skip).limit(limit).all()

def get_unread_messages(db: Session, user_id: int) -> List[Message]:
    return db.query(Message).filter(
        Message.receiver_id == user_id,
        Message.is_read == False
    ).all()

def create_message(db: Session, message_in: MessageCreate, sender_id: int) -> Message:
    db_message = Message(
        sender_id=sender_id,
        receiver_id=message_in.receiver_id,
        content=message_in.content
    )
    db.add(db_message)
    db.commit()
    db.refresh(db_message)
    return db_message

def mark_message_as_read(db: Session, message: Message) -> Message:
    message.is_read = True
    db.add(message)
    db.commit()
    db.refresh(message)
    return message

def mark_all_messages_as_read(db: Session, receiver_id: int, sender_id: int) -> None:
    db.query(Message).filter(
        Message.receiver_id == receiver_id,
        Message.sender_id == sender_id,
        Message.is_read == False
    ).update({"is_read": True})
    db.commit()
