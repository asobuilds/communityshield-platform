from fastapi import FastAPI

from app.database import Base, engine
from app.routers.location import router as location_router


app = FastAPI(
    title="CommunityShield API",
)


Base.metadata.create_all(bind=engine)


app.include_router(location_router)