import math


def validate_coordinates(latitude: float, longitude: float) -> bool:
    """
    Validate latitude and longitude coordinates.

    Returns True when the coordinates are valid.
    Raises ValueError when the coordinates are invalid.
    """

    if not isinstance(latitude, (int, float)):
        raise ValueError("Latitude must be a number.")

    if not isinstance(longitude, (int, float)):
        raise ValueError("Longitude must be a number.")

    if not math.isfinite(latitude):
        raise ValueError("Latitude must be finite.")

    if not math.isfinite(longitude):
        raise ValueError("Longitude must be finite.")

    if latitude < -90 or latitude > 90:
        raise ValueError("Latitude must be between -90 and 90.")

    if longitude < -180 or longitude > 180:
        raise ValueError("Longitude must be between -180 and 180.")

    return True