from math import radians, sin, cos, sqrt, atan2

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


def find_nearby_locations(
    db: Session,
    latitude: float,
    longitude: float,
    radius_km: float,
) -> list[Location]:
    validate_coordinates(latitude, longitude)

    locations = db.query(Location).all()

    nearby_locations = []

    earth_radius_km = 6371.0

    for location in locations:
        lat1 = radians(latitude)
        lon1 = radians(longitude)

        lat2 = radians(location.latitude)
        lon2 = radians(location.longitude)

        dlat = lat2 - lat1
        dlon = lon2 - lon1

        a = (
            sin(dlat / 2) ** 2
            + cos(lat1) * cos(lat2) * sin(dlon / 2) ** 2
        )

        c = 2 * atan2(sqrt(a), sqrt(1 - a))

        distance_km = earth_radius_km * c

        if distance_km <= radius_km:
            nearby_locations.append(location)

    return nearby_locations