# Community Shield Location Specification

## Purpose

The Location model represents a real-world geographic location
that can be used by Community Shield to identify incidents,
emergency services, public places, and other geographically
relevant entities.

## 1. Location Identity

The identity section describes what the location is and provides
a unique identifier for the location.

### Fields

| Field | Type | Required | Description |
|---|---|---|---|
| location_id | string/UUID | Yes | Unique identifier for the location |
| name | string | No | Human-readable name of the location |
| location_type | enum | Yes | Defines what kind of location this is |

### Location Types

Possible location types include:

- public_place
- road
- neighborhood
- hospital
- police_station
- fire_station
- school
- shelter
- incident_site
- other


## 2. Geographic Position

The geographic position defines the precise or approximate
physical position of the location on Earth.

### Fields

| Field | Type | Required | Description |
|---|---|---|---|
| latitude | decimal | Yes | Latitude in decimal degrees |
| longitude | decimal | Yes | Longitude in decimal degrees |
| accuracy_meters | decimal | No | Estimated accuracy of the coordinates in meters |

### Coordinate Rules

- Latitude must be between -90 and +90.
- Longitude must be between -180 and +180.
- Coordinates must be stored as decimal degrees.
- Latitude and longitude should not be stored as formatted text.
- accuracy_meters should be positive when provided.

## 3. Address Information

Address information provides a human-readable representation
of the location.

### Fields

| Field | Type | Required | Description |
|---|---|---|---|
| address | string | No | Street address or physical address |
| city | string | No | City or locality |
| state | string | No | State or administrative region |
| country | string | No | Country |
| postal_code | string | No | Postal or ZIP code |

## 4. Description

Additional information that helps humans understand the location.

### Fields

| Field | Type | Required | Description |
|---|---|---|---|
| description | string | No | General description of the location |
| landmark | string | No | Nearby landmark or recognizable reference point |

## 5. Location Source and Reliability

Location information may come from different sources.
The system should record how the geographic information was obtained.

### Fields

| Field | Type | Required | Description |
|---|---|---|---|
| location_source | enum | No | Source of the geographic coordinates |

### Location Sources

Possible values include:

- gps
- user_input
- map_selection
- geocoding
- official_database
- external_api
- unknown

## 6. Status

The status indicates whether the location is currently considered
active or usable by the system.

### Fields

| Field | Type | Required | Description |
|---|---|---|---|
| status | enum | Yes | Current status of the location |

## 7. Metadata

Metadata helps the system track the origin and lifecycle
of location records.

### Fields

| Field | Type | Required | Description |
|---|---|---|---|
| created_at | datetime | Yes | When the location record was created |
| updated_at | datetime | Yes | When the location record was last updated |
| source | string | No | External or internal source of the record |
| external_id | string | No | Identifier used by an external system |

## 8. Spatial Representation

Community Shield locations represent geographic points on Earth.

The primary geographic coordinates are represented using:

- latitude
- longitude

Coordinates are stored using decimal degrees.

The eventual spatial database representation should support
geographic queries such as:

- finding locations within a radius
- calculating distance between locations
- finding the nearest emergency service
- determining whether an incident is inside a geographic area

The database may use a spatial POINT geometry representation,
such as PostgreSQL with PostGIS.

The API should expose latitude and longitude explicitly for
clients such as the Community Shield frontend.