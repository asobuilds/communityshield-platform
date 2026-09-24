import pytest

from sqlalchemy import create_engine
from sqlalchemy.orm import Session

from app.database import Base
from app.services.location_service import (
    create_location,
    get_location,
    update_location,
    delete_location,
)


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


def test_update_location():
    engine = create_engine("sqlite:///:memory:")

    Base.metadata.create_all(engine)

    with Session(engine) as db:
        location = create_location(
            db=db,
            name="Original Location",
            address="Original Address",
            latitude=6.5244,
            longitude=3.3792,
        )

        updated_location = update_location(
            db=db,
            location_id=location.id,
            name="Updated Location",
            address="Updated Address",
            latitude=8.9783,
            longitude=7.3635,
        )

        assert updated_location is not None
        assert updated_location.id == location.id
        assert updated_location.name == "Updated Location"
        assert updated_location.address == "Updated Address"
        assert updated_location.latitude == 8.9783
        assert updated_location.longitude == 7.3635


def test_delete_location():
    engine = create_engine("sqlite:///:memory:")

    Base.metadata.create_all(engine)

    with Session(engine) as db:
        location = create_location(
            db=db,
            name="Location to Delete",
            address="Test Address",
            latitude=6.5244,
            longitude=3.3792,
        )

        deleted_location = delete_location(
            db=db,
            location_id=location.id,
        )

        assert deleted_location is not None
        assert deleted_location.id == location.id

        saved_location = get_location(db, location.id)

        assert saved_location is None


def test_create_location_rejects_invalid_coordinates():
    engine = create_engine("sqlite:///:memory:")

    Base.metadata.create_all(engine)

    with Session(engine) as db:
        with pytest.raises(ValueError):
            create_location(
                db=db,
                name="Invalid Location",
                address="Test Address",
                latitude=91,
                longitude=181,
            )