from sqlalchemy import Float, Integer, String
from sqlalchemy.orm import Mapped, mapped_column

from app.database import Base


class Location(Base):
    __tablename__ = "locations"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)

    name: Mapped[str] = mapped_column(String(255), nullable=False)

    address: Mapped[str | None] = mapped_column(String(500), nullable=True)

    latitude: Mapped[float] = mapped_column(Float, nullable=False)

    longitude: Mapped[float] = mapped_column(Float, nullable=False)