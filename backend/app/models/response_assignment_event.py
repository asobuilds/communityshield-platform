from datetime import datetime

from sqlalchemy import Column, Integer, String, ForeignKey, DateTime
from sqlalchemy.orm import relationship

from app.database import Base


class ResponseAssignmentEvent(Base):
    __tablename__ = "response_assignment_events"

    id = Column(
        Integer,
        primary_key=True,
        index=True
    )

    response_assignment_id = Column(
        Integer,
        ForeignKey("response_assignments.id"),
        nullable=False,
        index=True
    )

    status = Column(
        String,
        nullable=False
    )

    recorded_at = Column(
        DateTime,
        default=datetime.utcnow,
        nullable=False
    )

    response_assignment = relationship(
        "ResponseAssignment"
    )