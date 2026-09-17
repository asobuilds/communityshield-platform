import pytest

from app.utils.location_validation import validate_coordinates


def test_valid_coordinates():
    validate_coordinates(8.9773, 7.3579)


def test_invalid_latitude():
    with pytest.raises(ValueError):
        validate_coordinates(100, 7.3579)


def test_invalid_longitude():
    with pytest.raises(ValueError):
        validate_coordinates(8.9773, 200)