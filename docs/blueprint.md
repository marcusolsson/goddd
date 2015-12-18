# Group Booking 
Used by persons with administrative permissions to book an order.

# Cargo [/cargos]

## book cargo [POST]
Books a cargo.

+ Response 200

## list booked cargos [GET]
Lists all booked cargos.

+ Response 200 (application/json)

			{
			  "cargos": [
				{
				  "arrivalDeadline": "0001-01-01T00:00:00Z",
				  "destination": "CNHKG",
				  "misrouted": false,
				  "origin": "SESTO",
				  "routed": false,
				  "trackingId": "ABC123"
				},
				{
				  "arrivalDeadline": "0001-01-01T00:00:00Z",
				  "destination": "SESTO",
				  "misrouted": false,
				  "origin": "AUMEL",
				  "routed": false,
				  "trackingId": "FTL456"
				}
			  ]
			}

## assign to route [POST /cargos/{id}/assign_to_route]
Assigns given route to the cargo.

+ Parameters
	+ id - ID of a cargo.

+ Response 200

## change destination [POST /cargos/{id}/change_destination]
Changes destination of the cargo. This might result in a misrouted cargo.

+ Parameters
	+ id - ID of a cargo.

+ Response 200

## request routes [GET /cargos/{id}/request_routes]
Requests routes based on current specification. Uses an external routing service provided by the routing package.

+ Parameters
	+ id - ID of a cargo.

+ Response 200 (application/json)

	+ Body

			{
			  "routes": [
				{
				  "legs": [
					{
					  "voyageNumber": "0301S",
					  "from": "SESTO",
					  "to": "FIHEL",
					  "loadTime": "2015-11-14T14:10:29.173391809Z",
					  "unloadTime": "2015-11-15T21:55:29.173391809Z"
					},
					{
					  "voyageNumber": "0100S",
					  "from": "FIHEL",
					  "to": "CNHKG",
					  "loadTime": "2015-11-18T02:19:29.173391809Z",
					  "unloadTime": "2015-11-19T04:11:29.173391809Z"
					}
				  ]
				},
				{
				  "legs": [
					{
					  "voyageNumber": "0400S",
					  "from": "SESTO",
					  "to": "JNTKO",
					  "loadTime": "2015-11-14T06:22:29.173415471Z",
					  "unloadTime": "2015-11-15T10:22:29.173415471Z"
					},
					{
					  "voyageNumber": "0200T",
					  "from": "JNTKO",
					  "to": "CNHKG",
					  "loadTime": "2015-11-17T10:45:29.173415471Z",
					  "unloadTime": "2015-11-18T11:48:29.173415471Z"
					}
				  ]
				}
			  ]
			}

# Locations [/locations]

## list registered locations [GET]
Lists all registered locations.

+ Response 200 (application/json)

			{
			  "locations": [
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
				},
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
				}
			  ]
			}

# Group Tracking
Provided to our customer to see the status of their cargo.

# track cargo [GET /cargos/{id}]
Returns the cargo's tracking information.

+ Parameters
	+ id - ID of a cargo.

+ Response 200 (application/json)

			{
			  "cargo": {
				"trackingId": "ABC123",
				"statusText": "Not received",
				"origin": "SESTO",
				"destination": "CNHKG",
				"eta": "0001-01-01T00:00:00Z",
				"nextExpectedActivity": "There are currently no expected activities for this cargo.",
				"misrouted": false,
				"routed": false,
				"arrivalDeadline": "0001-01-01T00:00:00Z",
				"events": null
			  }
			}

+ Response 404 (application/json)

# Group Handling
Allows the staff at each location to register handling events along the route.

# Incidents [/incidents]

## register incident [POST]
Registers handling events along the route.

+ Response 200
