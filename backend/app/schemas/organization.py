from pydantic import BaseModel


class OrganizationCreate(BaseModel):
    name: str
    organization_type: str | None = None
    description: str | None = None
    contact_phone: str | None = None
    contact_email: str | None = None


class OrganizationResponse(BaseModel):
    id: int
    name: str
    organization_type: str | None = None
    description: str | None = None
    contact_phone: str | None = None
    contact_email: str | None = None

    class Config:
        from_attributes = True