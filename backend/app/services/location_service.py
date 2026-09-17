from sqlalchemy.orm import Session

from app.models.location import Location
from app.utils.location_validation import validate_coordinates


def create_location(
    db: Session,
    name: str,
    address: str | None,
    latitude: float,
    longitude: float,
) -> Location:
    validate_coordinates(latitude, longitude)

    location = Location(
        name=name,
        address=address,
        latitude=latitude,
        longitude=longitude,
    )

    db.add(location)
    db.commit()
    db.refresh(location)
    return location


def get_location(db: Session, location_id: int) -> Location | None:
    return db.get(Location, location_id)


def update_location(
    db: Session,
    location_id: int,
    name: str,
    address: str | None,
    latitude: float,
    longitude: float,
) -> Location | None:
    location = db.get(Location, location_id)

    if location is None:
        return None

    validate_coordinates(latitude, longitude)

    location.name = name
    location.address = address
    location.latitude = latitude
    location.longitude = longitude

    db.commit()
    db.refresh(location)

    return location


def delete_location(
    db: Session,
    location_id: int,
) -> Location | None:
    location = db.get(Location, location_id)

    if location is None:
        return None

    db.delete(location)
    db.commit()

    return location