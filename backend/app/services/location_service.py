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
    db.refresh(location)
    return location


def get_location(db: Session, location_id: int) -> Location | None:
    return db.get(Location, location_id)