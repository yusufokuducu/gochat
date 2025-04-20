import logging
from sqlalchemy.orm import Session

from app.db.base import Base, engine
from app.core.config import settings
from app.models.user import User
from app.core.security import get_password_hash

logger = logging.getLogger(__name__)

# Create tables in the database
def init_db() -> None:
    Base.metadata.create_all(bind=engine)
    logger.info("Database tables created")

# Create initial admin user if it doesn't exist
def create_initial_user(db: Session) -> None:
    user = db.query(User).filter(User.username == "admin1").first()
    if not user:
        user = User(
            username="admin1",
            email="admin1@example.com",
            password_hash=get_password_hash("admin1234")
        )
        db.add(user)
        db.commit()
        logger.info("Initial admin user created")
    else:
        logger.info("Admin user already exists")

# Main function to initialize database
def main() -> None:
    logger.info("Creating initial data")
    init_db()
    
    from app.db.base import SessionLocal
    db = SessionLocal()
    try:
        create_initial_user(db)
    finally:
        db.close()
    
    logger.info("Initial data created")

if __name__ == "__main__":
    main()
