from math import radians, sin, cos, sqrt, atan2

from sqlalchemy.orm import Session

from app.models.emergency import EmergencyResource
from app.models.location import Location


def create_emergency_resource(
    db: Session,
    name: str,
    resource_type: str,
    location_id: int,
    organization_id: int | None = None,
    organization_name: str | None = None,
):
    location = db.get(Location, location_id)

    if not location:
        raise ValueError("Location not found")

    new_resource = EmergencyResource(
        name=name,
        resource_type=resource_type,
        location_id=location_id,
        organization_id=organization_id,
        organization_name=organization_name,
    )

    db.add(new_resource)
    db.commit()
    db.refresh(new_resource)

    return new_resource


def get_emergency_resource(
    db: Session,
    resource_id: int,
):
    return db.get(EmergencyResource, resource_id)


def get_emergency_resources(db: Session):
    return db.query(EmergencyResource).all()


def get_emergency_resources_by_type(
    db: Session,
    resource_type: str,
):
    return (
        db.query(EmergencyResource)
        .filter(EmergencyResource.resource_type == resource_type)
        .all()
    )


def find_nearby_emergency_resources(
    db: Session,
    latitude: float,
    longitude: float,
    radius_km: float,
):
    resources = db.query(EmergencyResource).all()

    nearby_resources = []

    earth_radius_km = 6371.0

    for resource in resources:
        location = resource.location

        lat1 = radians(latitude)
        lon1 = radians(longitude)
        lat2 = radians(location.latitude)
        lon2 = radians(location.longitude)

        dlat = lat2 - lat1
        dlon = lon2 - lon1

        a = (
            sin(dlat / 2) ** 2
            + cos(lat1)
            * cos(lat2)
            * sin(dlon / 2) ** 2
        )

        c = 2 * atan2(sqrt(a), sqrt(1 - a))

        distance_km = earth_radius_km * c

        if distance_km <= radius_km:
            nearby_resources.append(resource)

    return nearby_resources