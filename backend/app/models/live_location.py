from datetime import datetime

from sqlalchemy import DateTime, Float, Integer, String
from sqlalchemy.orm import Mapped, mapped_column

from app.database import Base


class LiveLocation(Base):
    __tablename__ = "live_locations"

    id: Mapped[int] = mapped_column(
        Integer,
        primary_key=True,
        index=True
    )

    user_id: Mapped[int] = mapped_column(
        Integer,
        nullable=False,
        index=True
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

    marker_type: Mapped[str | None] = mapped_column(
        String(100),
        nullable=True
    )

    recorded_at: Mapped[datetime] = mapped_column(
        DateTime,
        nullable=False
    )
