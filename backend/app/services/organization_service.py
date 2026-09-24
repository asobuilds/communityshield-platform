from sqlalchemy.orm import Session

from app.models.organization import Organization
from app.schemas.organization import OrganizationCreate


def create_organization(
    db: Session,
    organization: OrganizationCreate,
):
    new_organization = Organization(
        name=organization.name,
        organization_type=organization.organization_type,
        description=organization.description,
        contact_phone=organization.contact_phone,
        contact_email=organization.contact_email,
    )

    db.add(new_organization)
    db.commit()
    db.refresh(new_organization)

    return new_organization


def get_organization(
    db: Session,
    organization_id: int,
):
    return db.get(Organization, organization_id)


def get_organizations(db: Session):
    return db.query(Organization).all()