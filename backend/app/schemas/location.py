from pydantic import BaseModel, Field


class LocationCreate(BaseModel):
    name: str
    address: str | None = None
    latitude: float = Field(ge=-90, le=90)
    longitude: float = Field(ge=-180, le=180)


class LocationResponse(LocationCreate):
    id: int