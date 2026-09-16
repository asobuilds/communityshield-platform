from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from app.database import get_db
from app.schemas.location import LocationCreate, LocationResponse
from app.services.location_service import create_location, get_location


router = APIRouter(
    prefix="/locations",
    tags=["Locations"],
)


@router.post("/", response_model=LocationResponse)
def create_new_location(
    location: LocationCreate,
    db: Session = Depends(get_db),
):
    return create_location(
        db=db,
        name=location.name,
        address=location.address,
        latitude=location.latitude,
        longitude=location.longitude,
    )


@router.get("/{location_id}", response_model=LocationResponse)
def read_location(
    location_id: int,
    db: Session = Depends(get_db),
):
    location = get_location(db, location_id)

    if location is None:
        raise HTTPException(status_code=404, detail="Location not found")

    return location