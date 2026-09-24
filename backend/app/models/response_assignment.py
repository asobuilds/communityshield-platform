from sqlalchemy import Column, Integer, String, ForeignKey, DateTime
from sqlalchemy.orm import relationship
from datetime import datetime

from app.database import Base


class ResponseAssignment(Base):
    __tablename__ = "response_assignments"

    id = Column(Integer, primary_key=True, index=True)

    incident_id = Column(
        Integer,
        ForeignKey("incidents.id"),
        nullable=False
    )

    organization_id = Column(
        Integer,
        ForeignKey("organizations.id"),
        nullable=False
    )

    emergency_resource_id = Column(
        Integer,
        ForeignKey("emergency_resources.id"),
        nullable=True
    )

    responder_user_id = Column(
        Integer,
        nullable=True,
        index=True
    )

    status = Column(
        String,
        nullable=False,
        default="ASSIGNED"
    )

    assigned_at = Column(
        DateTime,
        default=datetime.utcnow,
        nullable=False
    )

    updated_at = Column(
        DateTime,
        default=datetime.utcnow,
        onupdate=datetime.utcnow,
        nullable=False
    )

    incident = relationship("Incident")
    organization = relationship("Organization")
    emergency_resource = relationship("EmergencyResource")