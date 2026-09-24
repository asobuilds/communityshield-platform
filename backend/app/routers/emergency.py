from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from app.database import get_db
from app.schemas.emergency import (
    EmergencyResourceCreate,
    EmergencyResourceResponse,
)
from app.services.emergency_service import (
    create_emergency_resource,
    get_emergency_resource,
    get_emergency_resources,
    get_emergency_resources_by_type,
    find_nearby_emergency_resources,
)


router = APIRouter(
    prefix="/emergency-resources",
    tags=["Emergency Resources"],
)


@router.post("/", response_model=EmergencyResourceResponse)
def create_new_emergency_resource(
    resource: EmergencyResourceCreate,
    db: Session = Depends(get_db),
):
    try:
        return create_emergency_resource(
            db=db,
            name=resource.name,
            resource_type=resource.resource_type,
            location_id=resource.location_id,
            organization_id=resource.organization_id,
            organization_name=resource.organization_name,
        )
    except ValueError as exc:
        raise HTTPException(
            status_code=404,
            detail=str(exc),
        )


@router.get("/", response_model=list[EmergencyResourceResponse])
def list_emergency_resources(
    db: Session = Depends(get_db),
):
    return get_emergency_resources(db)


@router.get("/nearby", response_model=list[EmergencyResourceResponse])
def nearby_emergency_resources(
    latitude: float,
    longitude: float,
    radius_km: float = 10,
    db: Session = Depends(get_db),
):
    return find_nearby_emergency_resources(
        db=db,
        latitude=latitude,
        longitude=longitude,
        radius_km=radius_km,
    )


@router.get("/type/{resource_type}", response_model=list[EmergencyResourceResponse])
def emergency_resources_by_type(
    resource_type: str,
    db: Session = Depends(get_db),
):
    return get_emergency_resources_by_type(
        db=db,
        resource_type=resource_type,
    )


@router.get("/{resource_id}", response_model=EmergencyResourceResponse)
def get_emergency_resource_by_id(
    resource_id: int,
    db: Session = Depends(get_db),
):
    resource = get_emergency_resource(db, resource_id)

    if not resource:
        raise HTTPException(
            status_code=404,
            detail="Emergency resource not found",
        )

    return resource