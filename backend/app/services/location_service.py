from sqlalchemy.orm import Session

from app.models.location import Location


def create_location(
    db: Session,
    name: str,
    address: str | None,
    latitude: float,
    longitude: float,
) -> Location:
    location = Location(
        name=name,
        address=address,
        latitude=latitude,
        longitude=longitude,
    )

    db.add(location)
    db.commit()
    return location