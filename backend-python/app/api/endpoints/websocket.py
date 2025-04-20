import json
from typing import Dict, List
from fastapi import APIRouter, Depends, WebSocket, WebSocketDisconnect, HTTPException, status
from sqlalchemy.orm import Session

from app.db.base import get_db
from app.models.user import User
from app.schemas.message import WebSocketMessage, MessageCreate
from app.services.message import create_message
from app.services.user import get_user_by_id

router = APIRouter()

# Store active connections
class ConnectionManager:
    def __init__(self):
        # user_id -> WebSocket
        self.active_connections: Dict[int, WebSocket] = {}
    
    async def connect(self, websocket: WebSocket, user_id: int):
        await websocket.accept()
        self.active_connections[user_id] = websocket
    
    def disconnect(self, user_id: int):
        if user_id in self.active_connections:
            del self.active_connections[user_id]
    
    async def send_personal_message(self, message: str, user_id: int):
        if user_id in self.active_connections:
            await self.active_connections[user_id].send_text(message)
    
    async def broadcast(self, message: str, exclude_user_id: int = None):
        for user_id, connection in self.active_connections.items():
            if exclude_user_id is None or user_id != exclude_user_id:
                await connection.send_text(message)

manager = ConnectionManager()

@router.websocket("/chat/{token}")
async def websocket_endpoint(
    websocket: WebSocket,
    token: str,
    db: Session = Depends(get_db)
):
    # Authenticate user from token
    try:
        from app.api.deps import get_current_user
        user = get_current_user(db=db, token=token)
    except HTTPException:
        await websocket.close(code=status.WS_1008_POLICY_VIOLATION)
        return
    
    await manager.connect(websocket, user.id)
    
    try:
        while True:
            data = await websocket.receive_text()
            message_data = json.loads(data)
            
            # Validate message format
            try:
                ws_message = WebSocketMessage(**message_data)
            except Exception as e:
                await websocket.send_text(json.dumps({"error": "Invalid message format"}))
                continue
            
            # Check if sender is the authenticated user
            if ws_message.sender_id != user.id:
                await websocket.send_text(json.dumps({"error": "Sender ID does not match authenticated user"}))
                continue
            
            # Check if receiver exists
            receiver = get_user_by_id(db, ws_message.receiver_id)
            if not receiver:
                await websocket.send_text(json.dumps({"error": "Receiver not found"}))
                continue
            
            # Create message in database
            message_in = MessageCreate(
                content=ws_message.content,
                receiver_id=ws_message.receiver_id
            )
            db_message = create_message(db, message_in, user.id)
            
            # Format message for sending
            message_json = json.dumps({
                "id": db_message.id,
                "sender_id": db_message.sender_id,
                "receiver_id": db_message.receiver_id,
                "content": db_message.content,
                "timestamp": db_message.timestamp.isoformat(),
                "is_read": db_message.is_read
            })
            
            # Send to receiver if online
            await manager.send_personal_message(message_json, ws_message.receiver_id)
            
            # Send confirmation to sender
            await websocket.send_text(json.dumps({"status": "sent", "message_id": db_message.id}))
            
    except WebSocketDisconnect:
        manager.disconnect(user.id)
    except Exception as e:
        manager.disconnect(user.id)
        await websocket.close(code=status.WS_1011_INTERNAL_ERROR)
