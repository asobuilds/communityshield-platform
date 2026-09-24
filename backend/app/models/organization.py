from sqlalchemy import ForeignKey
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.database import Base


class Organization(Base):
    __tablename__ = "organizations"

    id: Mapped[int] = mapped_column(
        primary_key=True,
        index=True
    )

    name: Mapped[str] = mapped_column(
        nullable=False
    )

    organization_type: Mapped[str | None] = mapped_column(
        nullable=True
    )

    description: Mapped[str | None] = mapped_column(
        nullable=True
    )

    contact_phone: Mapped[str | None] = mapped_column(
        nullable=True
    )

    contact_email: Mapped[str | None] = mapped_column(
        nullable=True
    )

    location_id: Mapped[int | None] = mapped_column(
        ForeignKey("locations.id"),
        nullable=True,
        index=True
    )

    location = relationship(
        "Location",
        back_populates="organizations"
    )