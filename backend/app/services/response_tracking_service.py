from sqlalchemy.orm import Session

from app.models.response_assignment import ResponseAssignment
from app.models.live_location import LiveLocation
from app.models.incident import Incident


def get_assignment_latest_live_location(
    db: Session,
    assignment_id: int,
):
    assignment = db.get(
        ResponseAssignment,
        assignment_id,
    )

    if not assignment:
        raise ValueError("Response assignment not found")

    incident = db.get(
        Incident,
        assignment.incident_id,
    )

    if not incident:
        raise ValueError("Incident not found")

    if not assignment.organization_id:
        return None

    live_location = (
        db.query(LiveLocation)
        .filter(
            LiveLocation.organization_id
            == assignment.organization_id
        )
        .order_by(
            LiveLocation.recorded_at.desc()
        )
        .first()
    )

    return live_location