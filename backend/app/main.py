from fastapi import FastAPI

from app.database import Base, engine

from app.routers.location import router as location_router
from app.routers.emergency import router as emergency_router
from app.routers.organization import router as organization_router
from app.routers.live_location import router as live_location_router
from app.routers.incident import router as incident_router
from app.routers.response_assignment import router as response_assignment_router


app = FastAPI(
    title="Nativity Shield",
)


Base.metadata.create_all(bind=engine)


app.include_router(location_router)
app.include_router(emergency_router)
app.include_router(organization_router)
app.include_router(live_location_router)
app.include_router(incident_router)
app.include_router(response_assignment_router)