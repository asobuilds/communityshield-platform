from math import asin, cos, radians, sin, sqrt

from sqlalchemy.orm import Session

from app.models.emergency import EmergencyResource
from app.models.incident import Incident
from app.models.live_location import LiveLocation
from app.schemas.incident import IncidentCreate


def create_incident(
    db: Session,
    incident_data: IncidentCreate,
) -> Incident:
    incident = Incident(
        title=incident_data.title,
        description=incident_data.description,
        incident_type=incident_data.incident_type,
        severity=incident_data.severity,
        status=incident_data.status,
        location_id=incident_data.location_id,
        reporter_id=incident_data.reporter_id,
        reported_at=incident_data.reported_at,
    )

    db.add(incident)
    db.commit()
    db.refresh(incident)

    return incident


def get_incident(
    db: Session,
    incident_id: int,
) -> Incident | None:
    return db.get(Incident, incident_id)


def get_incidents(
    db: Session,
) -> list[Incident]:
    return db.query(Incident).all()


def update_incident(
    db: Session,
    incident_id: int,
    incident_data: IncidentCreate,
) -> Incident | None:
    incident = db.get(Incident, incident_id)

    if incident is None:
        return None

    incident.title = incident_data.title
    incident.description = incident_data.description
    incident.incident_type = incident_data.incident_type
    incident.severity = incident_data.severity
    incident.status = incident_data.status
    incident.location_id = incident_data.location_id
    incident.reporter_id = incident_data.reporter_id
    incident.reported_at = incident_data.reported_at

    db.commit()
    db.refresh(incident)

    return incident


def update_incident_status(
    db: Session,
    incident_id: int,
    status: str,
) -> Incident | None:
    incident = db.get(Incident, incident_id)

    if incident is None:
        return None

    incident.status = status

    db.commit()
    db.refresh(incident)

    return incident


def delete_incident(
    db: Session,
    incident_id: int,
) -> bool:
    incident = db.get(Incident, incident_id)

    if incident is None:
        return False

    db.delete(incident)
    db.commit()

    return True


def find_nearby_resources_for_incident(
    db: Session,
    incident_id: int,
    radius_km: float,
) -> list[EmergencyResource]:
    incident = db.get(Incident, incident_id)

    if incident is None:
        return []

    location = incident.location

    if location is None:
        return []

    resources = db.query(EmergencyResource).all()

    nearby_resources = []

    earth_radius_km = 6371.0

    for resource in resources:
        resource_location = resource.location

        if resource_location is None:
            continue

        lat1 = radians(location.latitude)
        lat2 = radians(resource_location.latitude)

        delta_lat = radians(
            resource_location.latitude - location.latitude
        )

        delta_lon = radians(
            resource_location.longitude - location.longitude
        )

        a = (
            sin(delta_lat / 2) ** 2
            + cos(lat1)
            * cos(lat2)
            * sin(delta_lon / 2) ** 2
        )

        distance = (
            2
            * earth_radius_km
            * asin(sqrt(a))
        )

        if distance <= radius_km:
            nearby_resources.append(resource)

    return nearby_resources


def find_nearby_live_locations_for_incident(
    db: Session,
    incident_id: int,
    radius_km: float,
) -> list[LiveLocation]:
    incident = db.get(Incident, incident_id)

    if incident is None:
        return []

    location = incident.location

    if location is None:
        return []

    live_locations = db.query(LiveLocation).all()

    nearby_locations = []

    earth_radius_km = 6371.0

    for live_location in live_locations:
        lat1 = radians(location.latitude)
        lat2 = radians(live_location.latitude)

        delta_lat = radians(
            live_location.latitude - location.latitude
        )

        delta_lon = radians(
            live_location.longitude - location.longitude
        )

        a = (
            sin(delta_lat / 2) ** 2
            + cos(lat1)
            * cos(lat2)
            * sin(delta_lon / 2) ** 2
        )

        distance = (
            2
            * earth_radius_km
            * asin(sqrt(a))
        )

        if distance <= radius_km:
            nearby_locations.append(live_location)

    return nearby_locations