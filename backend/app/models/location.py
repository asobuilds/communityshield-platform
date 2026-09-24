from enum import Enum

from sqlalchemy import CheckConstraint, Enum as SQLEnum
from sqlalchemy import Float, Integer, String
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.database import Base


class LocationType(str, Enum):
    PUBLIC_PLACE = "PUBLIC_PLACE"
    ROAD = "ROAD"
    NEIGHBORHOOD = "NEIGHBORHOOD"
    LANDMARK = "LANDMARK"
    INCIDENT_ZONE = "INCIDENT_ZONE"


class CoordinateSource(str, Enum):
    GPS = "GPS"
    OFFICIAL_RECORD = "OFFICIAL_RECORD"
    GEOCODED_ADDRESS = "GEOCODED_ADDRESS"
    MAP_SELECTED = "MAP_SELECTED"
    MANUAL_ENTRY = "MANUAL_ENTRY"


class Location(Base):
    __tablename__ = "locations"

    __table_args__ = (
        CheckConstraint(
            "accuracy_meters IS NULL OR accuracy_meters >= 0",
            name="ck_locations_accuracy_nonnegative"
        ),
    )

    id: Mapped[int] = mapped_column(
        Integer,
        primary_key=True,
        index=True
    )

    name: Mapped[str] = mapped_column(
        String(255),
        nullable=False
    )

    address: Mapped[str | None] = mapped_column(
        String(500),
        nullable=True
    )

    location_type: Mapped[LocationType | None] = mapped_column(
        SQLEnum(LocationType),
        nullable=True
    )

    latitude: Mapped[float] = mapped_column(
        Float,
        nullable=False
    )

    longitude: Mapped[float] = mapped_column(
        Float,
        nullable=False
    )

    accuracy_meters: Mapped[float | None] = mapped_column(
        Float,
        nullable=True
    )

    coordinate_source: Mapped[CoordinateSource | None] = mapped_column(
        SQLEnum(CoordinateSource),
        nullable=True
    )

    organizations = relationship(
        "Organization",
        back_populates="location"
    )

    emergency_resources = relationship(
        "EmergencyResource",
        back_populates="location"
    )
