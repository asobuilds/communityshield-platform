from datetime import datetime

from pydantic import BaseModel, Field


class LiveLocationCreate(BaseModel):
    user_id: int
    latitude: float = Field(..., ge=-90, le=90)
    longitude: float = Field(..., ge=-180, le=180)
    accuracy_meters: float | None = Field(default=None, ge=0)
    marker_type: str | None = None
    recorded_at: datetime


class LiveLocationResponse(BaseModel):
    id: int
    user_id: int
    latitude: float
    longitude: float
    accuracy_meters: float | None
    marker_type: str | None
    recorded_at: datetime

    model_config = {
        "from_attributes": True
    }