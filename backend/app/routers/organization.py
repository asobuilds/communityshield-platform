from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from app.database import get_db
from app.schemas.organization import (
    OrganizationCreate,
    OrganizationResponse,
)
from app.services.organization_service import (
    create_organization,
    get_organization,
    get_organizations,
)


router = APIRouter(
    prefix="/organizations",
    tags=["Organizations"],
)


@router.post("/", response_model=OrganizationResponse)
def create_new_organization(
    organization: OrganizationCreate,
    db: Session = Depends(get_db),
):
    return create_organization(db, organization)


@router.get("/", response_model=list[OrganizationResponse])
def list_organizations(
    db: Session = Depends(get_db),
):
    return get_organizations(db)


@router.get("/{organization_id}", response_model=OrganizationResponse)
def get_organization_by_id(
    organization_id: int,
    db: Session = Depends(get_db),
):
    organization = get_organization(db, organization_id)

    if not organization:
        raise HTTPException(
            status_code=404,
            detail="Organization not found",
        )

    return organization