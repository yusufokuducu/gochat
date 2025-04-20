from fastapi import APIRouter

from app.api.endpoints import auth, users, friendships, messages, websocket

api_router = APIRouter()

api_router.include_router(auth.router, prefix="/auth", tags=["authentication"])
api_router.include_router(users.router, prefix="/users", tags=["users"])
api_router.include_router(friendships.router, prefix="/friendships", tags=["friendships"])
api_router.include_router(messages.router, prefix="/messages", tags=["messages"])
api_router.include_router(websocket.router, prefix="/ws", tags=["websocket"])
