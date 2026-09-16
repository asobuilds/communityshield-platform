from sqlalchemy import create_engine
from sqlalchemy.orm import Session

from app.database import Base
from app.models.location import Location
from app.services.location_service import create_location, get_location


def test_create_and_get_location():
    engine = create_engine("sqlite:///:memory:")

    Base.metadata.create_all(engine)

    with Session(engine) as db:
        location = create_location(
            db=db,
            name="Test Location",
            address="Test Address",
            latitude=6.5244,
            longitude=3.3792,
        )

        assert location.id is not None
        assert location.name == "Test Location"

        saved_location = get_location(db, location.id)

        assert saved_location is not None
        assert saved_location.id == location.id
        assert saved_location.latitude == 6.5244
        assert saved_location.longitude == 3.3792