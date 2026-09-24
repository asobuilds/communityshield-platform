from datetime import datetime

from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from app.database import get_db
from app.schemas.live_location import (
    LiveLocationCreate,
    LiveLocationResponse,
)
from app.services.live_location_service import (
    create_live_location,
    find_nearby_live_locations,
    get_latest_live_location,
)

router = APIRouter(
    prefix="/live-locations",
    tags=["Live Locations"],
)


@router.post(
    "/",
    response_model=LiveLocationResponse,
)
def create_new_live_location(
    location: LiveLocationCreate,
    db: Session = Depends(get_db),
):
    return create_live_location(db, location)


@router.get(
    "/user/{user_id}",
    response_model=LiveLocationResponse,
)
def get_user_latest_live_location(
    user_id: int,
    db: Session = Depends(get_db),
):
    location = get_latest_live_location(db, user_id)

    if location is None:
        raise HTTPException(
            status_code=404,
            detail="Live location not found",
        )

    return location


@router.get(
    "/nearby",
    response_model=list[LiveLocationResponse],
)
def get_nearby_live_locations(
    latitude: float = Query(..., ge=-90, le=90),
    longitude: float = Query(..., ge=-180, le=180),
    radius_km: float = Query(..., gt=0),
    db: Session = Depends(get_db),
):
    return find_nearby_live_locations(
        db=db,
        latitude=latitude,
        longitude=longitude,
        radius_km=radius_km,
    )