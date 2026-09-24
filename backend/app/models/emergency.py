from sqlalchemy import Column, Integer, String, ForeignKey
from sqlalchemy.orm import relationship

from app.database import Base


class EmergencyResource(Base):
    __tablename__ = "emergency_resources"

    id = Column(Integer, primary_key=True, index=True)

    name = Column(String, nullable=False)

    resource_type = Column(String, nullable=False)

    location_id = Column(
        Integer,
        ForeignKey("locations.id"),
        nullable=False,
    )

    organization_id = Column(
        Integer,
        ForeignKey("organizations.id"),
        nullable=True,
    )

    organization_name = Column(String, nullable=True)

    location = relationship(
        "Location",
        back_populates="emergency_resources"
    )

    organization = relationship(
        "Organization"
    )