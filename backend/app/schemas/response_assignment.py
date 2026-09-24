from datetime import datetime

from pydantic import BaseModel, ConfigDict


class ResponseAssignmentCreate(BaseModel):
    incident_id: int
    organization_id: int
    emergency_resource_id: int | None = None
    responder_user_id: int | None = None


class ResponseAssignmentUpdate(BaseModel):
    status: str


class ResponseAssignmentResponse(BaseModel):
    id: int
    incident_id: int
    organization_id: int
    emergency_resource_id: int | None
    responder_user_id: int | None
    status: str
    assigned_at: datetime
    updated_at: datetime

    model_config = ConfigDict(from_attributes=True)