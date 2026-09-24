from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from app.database import get_db

from app.schemas.response_assignment import (
    ResponseAssignmentCreate,
    ResponseAssignmentResponse,
    ResponseAssignmentUpdate,
)

from app.services.response_assignment_service import (
    create_response_assignment,
    get_response_assignment,
    get_response_assignments,
    get_incident_response_assignments,
    update_response_assignment_status,
    delete_response_assignment,
    get_responder_latest_live_location,
    get_response_assignment_events,
    get_incident_response_summary,
)


router = APIRouter(
    prefix="/response-assignments",
    tags=["Response Assignments"],
)


@router.post(
    "/",
    response_model=ResponseAssignmentResponse,
)
def create_assignment(
    assignment: ResponseAssignmentCreate,
    db: Session = Depends(get_db),
):
    try:
        return create_response_assignment(
            db=db,
            incident_id=assignment.incident_id,
            organization_id=assignment.organization_id,
            emergency_resource_id=(
                assignment.emergency_resource_id
            ),
            responder_user_id=(
                assignment.responder_user_id
            ),
        )

    except ValueError as e:
        raise HTTPException(
            status_code=404,
            detail=str(e),
        )


@router.get(
    "/",
    response_model=list[ResponseAssignmentResponse],
)
def get_assignments(
    db: Session = Depends(get_db),
):
    return get_response_assignments(db)


@router.get(
    "/incident/{incident_id}",
    response_model=list[ResponseAssignmentResponse],
)
def get_assignments_for_incident(
    incident_id: int,
    db: Session = Depends(get_db),
):
    return get_incident_response_assignments(
        db,
        incident_id,
    )


@router.get(
    "/incident/{incident_id}/summary",
)
def get_incident_summary(
    incident_id: int,
    db: Session = Depends(get_db),
):
    summary = get_incident_response_summary(
        db,
        incident_id,
    )

    if not summary:
        raise HTTPException(
            status_code=404,
            detail="Incident not found",
        )

    return summary


@router.get(
    "/{assignment_id}/history",
)
def get_assignment_history(
    assignment_id: int,
    db: Session = Depends(get_db),
):
    assignment = get_response_assignment(
        db,
        assignment_id,
    )

    if not assignment:
        raise HTTPException(
            status_code=404,
            detail="Response assignment not found",
        )

    events = get_response_assignment_events(
        db,
        assignment_id,
    )

    return events


@router.get(
    "/{assignment_id}/live-location",
)
def get_assignment_live_location(
    assignment_id: int,
    db: Session = Depends(get_db),
):
    try:
        live_location = (
            get_responder_latest_live_location(
                db,
                assignment_id,
            )
        )

    except ValueError as e:
        raise HTTPException(
            status_code=404,
            detail=str(e),
        )

    if not live_location:
        raise HTTPException(
            status_code=404,
            detail="No live location found for responder",
        )

    return {
        "assignment_id": assignment_id,
        "user_id": live_location.user_id,
        "latitude": live_location.latitude,
        "longitude": live_location.longitude,
        "accuracy_meters": (
            live_location.accuracy_meters
        ),
        "marker_type": live_location.marker_type,
        "recorded_at": live_location.recorded_at,
    }


@router.get(
    "/{assignment_id}",
    response_model=ResponseAssignmentResponse,
)
def get_assignment(
    assignment_id: int,
    db: Session = Depends(get_db),
):
    assignment = get_response_assignment(
        db,
        assignment_id,
    )

    if not assignment:
        raise HTTPException(
            status_code=404,
            detail="Response assignment not found",
        )

    return assignment


@router.put(
    "/{assignment_id}/status",
    response_model=ResponseAssignmentResponse,
)
def update_assignment_status(
    assignment_id: int,
    assignment: ResponseAssignmentUpdate,
    db: Session = Depends(get_db),
):
    updated_assignment = (
        update_response_assignment_status(
            db=db,
            assignment_id=assignment_id,
            status=assignment.status,
        )
    )

    if not updated_assignment:
        raise HTTPException(
            status_code=404,
            detail="Response assignment not found",
        )

    return updated_assignment


@router.delete(
    "/{assignment_id}",
)
def delete_assignment(
    assignment_id: int,
    db: Session = Depends(get_db),
):
    deleted = delete_response_assignment(
        db,
        assignment_id,
    )

    if not deleted:
        raise HTTPException(
            status_code=404,
            detail="Response assignment not found",
        )

    return {
        "message": "Response assignment deleted successfully"
    }