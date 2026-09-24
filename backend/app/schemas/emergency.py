from pydantic import BaseModel


class EmergencyResourceCreate(BaseModel):
    name: str
    resource_type: str
    location_id: int
    organization_id: int | None = None
    organization_name: str | None = None


class EmergencyResourceResponse(BaseModel):
    id: int
    name: str
    resource_type: str
    location_id: int
    organization_id: int | None = None
    organization_name: str | None = None

    class Config:
        from_attributes = True