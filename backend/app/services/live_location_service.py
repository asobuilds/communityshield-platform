from math import asin, cos, radians, sin, sqrt

from sqlalchemy.orm import Session

from app.models.live_location import LiveLocation
from app.schemas.live_location import LiveLocationCreate


def create_live_location(
    db: Session,
    location_data: LiveLocationCreate,
) -> LiveLocation:
    live_location = LiveLocation(
        user_id=location_data.user_id,
        latitude=location_data.latitude,
        longitude=location_data.longitude,
        accuracy_meters=location_data.accuracy_meters,
        marker_type=location_data.marker_type,
        recorded_at=location_data.recorded_at,
    )

    db.add(live_location)
    db.commit()
    db.refresh(live_location)

    return live_location


def get_live_location(
    db: Session,
    location_id: int,
) -> LiveLocation | None:
    return db.get(LiveLocation, location_id)


def get_latest_live_location(
    db: Session,
    user_id: int,
) -> LiveLocation | None:
    return (
        db.query(LiveLocation)
        .filter(LiveLocation.user_id == user_id)
        .order_by(LiveLocation.recorded_at.desc())
        .first()
    )


def find_nearby_live_locations(
    db: Session,
    latitude: float,
    longitude: float,
    radius_km: float,
) -> list[LiveLocation]:
    locations = db.query(LiveLocation).all()

    nearby_locations = []

    earth_radius_km = 6371.0

    for location in locations:
        lat1 = radians(latitude)
        lat2 = radians(location.latitude)

        delta_lat = radians(location.latitude - latitude)
        delta_lon = radians(location.longitude - longitude)

        a = (
            sin(delta_lat / 2) ** 2
            + cos(lat1)
            * cos(lat2)
            * sin(delta_lon / 2) ** 2
        )

        distance = 2 * earth_radius_km * asin(sqrt(a))

        if distance <= radius_km:
            nearby_locations.append(location)

    return nearby_locations