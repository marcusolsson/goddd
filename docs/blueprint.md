# dddsample API

# Group Booking 
Used by persons with administrative permissions to book an order.

## book cargo [/cargos{?origin,destination,arrival_deadline}]

+ Parameters
    + origin (required, string) - name of origin
    + destination (required, string) - name of destination
    + arrival_deadline (required, string) - deadline of arrival

### POST
Books a cargo.

+ Response 200 (application/json)

        {
            "tracking_id": "ABC123"
        }

## assign to route [/cargos/{id}/assign_to_route]

+ Parameters
    + id (required, string) - tracking ID of the cargo

### POST
Assigns given route to the cargo.

+ Request (application/json)

        {
            "legs": [
                {
                    "voyage_number": "0301S",
                    "from": "SESTO",
                    "to": "FIHEL",
                    "load_time": "2015-11-14T14:10:29.173391809Z",
                    "unload_time": "2015-11-15T21:55:29.173391809Z"
                },
                {
                    "voyage_number": "0100S",
                    "from": "FIHEL",
                    "to": "CNHKG",
                    "load_time": "2015-11-18T02:19:29.173391809Z",
                    "unload_time": "2015-11-19T04:11:29.173391809Z"
                }
            ]
        },

+ Response 200 

## change destination [/cargos/{id}/change_destination{?destination}]

+ Parameters
    + id (required, string) - tracking ID of the cargo
    + destination (required, string) - UN locode of the destination

### POST
Changes destination of the cargo. This might result in a misrouted cargo.

+ Response 200

## request routes [/cargos/{id}/request_routes]

+ Parameters
    + id (required, string) - tracking ID of the cargo

### GET
Requests routes based on current specification. Uses an external routing service provided by the routing package.

+ Response 200 (application/json)

        {
            "routes": [
                {
                    "legs": [
                        {
                            "voyage_number": "0301S",
                            "from": "SESTO",
                            "to": "FIHEL",
                            "load_time": "2015-11-14T14:10:29.173391809Z",
                            "unload_time": "2015-11-15T21:55:29.173391809Z"
                        },
                        {
                            "voyage_number": "0100S",
                            "from": "FIHEL",
                            "to": "CNHKG",
                            "load_time": "2015-11-18T02:19:29.173391809Z",
                            "unload_time": "2015-11-19T04:11:29.173391809Z"
                        }
                    ]
                },
                {
                    "legs": [
                        {
                            "voyage_number": "0400S",
                            "from": "SESTO",
                            "to": "JNTKO",
                            "load_time": "2015-11-14T06:22:29.173415471Z",
                            "unload_time": "2015-11-15T10:22:29.173415471Z"
                        },
                        {
                            "voyage_number": "0200T",
                            "from": "JNTKO",
                            "to": "CNHKG",
                            "load_time": "2015-11-17T10:45:29.173415471Z",
                            "unload_time": "2015-11-18T11:48:29.173415471Z"
                        }
                    ]
                }
            ]
        }

## list cargos [/cargos]

### GET
Lists all booked cargos.

+ Response 200 (application/json)

        {
            "cargos": [
                {
                    "arrival_deadline": "0001-01-01T00:00:00Z",
                    "destination": "CNHKG",
                    "misrouted": false,
                    "origin": "SESTO",
                    "routed": false,
                    "tracking_id": "ABC123"
                },
                {
                    "arrival_deadline": "0001-01-01T00:00:00Z",
                    "destination": "SESTO",
                    "misrouted": false,
                    "origin": "AUMEL",
                    "routed": false,
                    "tracking_id": "FTL456"
                }
            ]
        }

## list locations [/locations]

### GET
Lists all registered locations.

+ Response 200 (application/json)

        {
            "locations": [
                {
                    "locode": "DEHAM",
                    "name": "Hamburg"
                },
                {
                    "locode": "SESTO",
                    "name": "Stockholm"
                },
                {
                    "locode": "AUMEL",
                    "name": "Melbourne"
                },
                {
                    "locode": "CNHKG",
                    "name": "Hongkong"
                },
                {
                    "locode": "JNTKO",
                    "name": "Tokyo"
                },
                {
                    "locode": "NLRTM",
                    "name": "Rotterdam"
                }
            ]
        }

# Group Tracking
Provided to our customer to see the status of their cargo.

## track cargo [/cargos/{id}]

+ Parameters
    + id (required, string) - tracking ID of the cargo

### GET
Returns the cargo's tracking information.

+ Response 200 (application/json)

        {
            "cargo": {
                "tracking_id": "ABC123",
                "status_text": "Not received",
                "origin": "SESTO",
                "destination": "CNHKG",
                "eta": "0001-01-01T00:00:00Z",
                "next_expected_activity": "There are currently no expected activities for this cargo.",
                "arrival_deadline": "0001-01-01T00:00:00Z",
                "events": null,
                "misrouted": false,
                "routed": false
            }
        }

+ Response 404

# Group Handling
Allows the staff at each location to register handling events along the route.

## register incident [/incidents{?completion_time,tracking_id,voyage,location,event_type}]

+ Parameters
    + completion_time (required, string) - time when incident was completed
    + tracking_id (required, string) - tracking ID of the cargo
    + voyage (required, string) - voyage number
    + location (required, string) - UN locode of where the incident occurred
    + event_type (required, string) - type of handling event

### POST
Registers handling events along the route.

+ Response 200
