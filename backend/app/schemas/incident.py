from datetime import datetime

from pydantic import BaseModel, Field


class IncidentCreate(BaseModel):
    title: str
    description: str | None = None
    incident_type: str
    severity: str
    status: str = "OPEN"
    location_id: int
    reporter_id: int | None = None
    reported_at: datetime


class IncidentResponse(BaseModel):
    id: int
    title: str
    description: str | None
    incident_type: str
    severity: str
    status: str
    location_id: int
    reporter_id: int | None
    reported_at: datetime

    model_config = {
        "from_attributes": True
    }