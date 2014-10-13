# GoDDD [![wercker status](https://app.wercker.com/status/12b6c3ee4ea86efc7fd5184b589916b7/s "wercker status")](https://app.wercker.com/project/bykey/12b6c3ee4ea86efc7fd5184b589916b7)

This is an attempt to port the [DDD Sample App](http://dddsample.sourceforge.net/) to Go. The purpose is to explore how to write idiomatic Go applications using Domain-Driven Design.

I will update this README along the way, as I gain more insights. Therefore, the implementation may be different from the information in this document.

This project is __not__ intended as a tutorial or guide on how to do DDD in Go. It is meant to be a learning experience.

## Design

The purpose of this sample is because I wanted to learn Go and to see how well it can be used with Domain Driven Design. When faced with design challenges, I will lean towards the more idiomatic solution.

### Equality

In Go, two struct values are equal if their corresponding non-blank fields are equal. You cannot overload equality for structs and there is no standard interface for equality. For value objects, this seems nice at first but when comparing `Itinerary` structs you will get an error.

    type Itinerary struct {
	        Legs []Leg
    }

The reason is because `==` does a shallow comparison and that is probably not what you want in this case. The solution is to use [DeepEqual](http://golang.org/pkg/reflect/#DeepEqual) function, which will scan all the arrays and maps (if any) as well. One alternative is to implement a `ValueObject` interface:

	type ValueObject interface {
			SameValue(other ValueObject) bool
	}

    func (i Itinerary) SameValue(other ValueObject) bool {
		    return reflect.DeepEqual(i, other.(Itinerary))
    }

For entities we can then similarly implement a `Entity` interface*:

	type Entity interface {
			SameValue(other Entity) bool
	}

    func (c *Cargo) SameIdentity(other Entity) bool {
		    return c.TrackingId == other.(*Cargo).TrackingId
    }

*In both of cases though, it will still be very tempting to use the `==`.

__Read more__

- [Comparison operators](http://golang.org/ref/spec#Comparison_operators).
- [Equaler interface](https://golang.org/doc/faq#t_and_equal_interface)

### Immutability

Go does not support means of creating a immutable struct. All fields can be altered after creation. It is however possible to use interfaces to limit altering structs after creation.

    type Leg interface {
			LoadLocation() location.Location
    }

    type leg struct {
         loadLocation location.Location
    }

    func (l leg) LoadLocation() location.Location {
         return l.loadLocation
    }

    func NewLeg(load location.Location) Leg {
         return leg{loadLocation: load}
    }

Since the `leg` struct starts with a lowercase, it will not be exported outside the package. The jury is still out on this one though. On one hand, it makes it natural to use the `func New...` idiom to initialize the value object. On the other hand, it kind of feels non-idiomatic. A more idiomatic alternative would probably be to use _zero-initialization_ to construct a valid value object, and make sure the developers understand the concept of value objects.

Also, making sure that the _method receiver_ is of _non-pointer_ type feels like a good way to handle the temptation to modify a value object. Entities though, are more likely to have a pointer type receiver so that they may change during their life cycle.

__Read more__

- [Immutable objects](https://groups.google.com/forum/#!topic/golang-nuts/BnjG3N77Ico)

### Clean Architecture

__Read more__

- [The Clean Architecture](http://blog.8thlight.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Applying The Clean Architecture to Go applications](http://manuel.kiessling.net/2012/09/28/applying-the-clean-architecture-to-go-applications/)

### Other thoughts ...

- How can we use the zero-initialization idiom effectively in DDD? What does a zero-initialized Itinerary mean?
- Concurrency is one area where Go shines, but initial thought is to keep it out of the domain model. This might be interesting if concurrency is a explicit part of the model.

## REST API

The application exposes a REST API.

### Setup

1. Make sure you have your `$GOPATH` environment set.

    `export GOPATH=$HOME/go`

2. Get the latest version by running:

    `go get -u github.com/marcusolsson/goddd`

3. Run it!:

    `go run local.go`

_Note:_ The `appengine.go` is used to deploy it on Google App Engine. If you want to deploy it to your own project, edit the `app.yaml` to point to your own project. Read more about [deploying Go applications to App Engine](https://cloud.google.com/appengine/docs/go/).

### Cargos

#### GET /cargos
Returns a list of all currently booked cargos.

#### GET /cargos/:id
Returns a cargo with a given tracking ID.

__Example:__

    {
        "trackingId": "ABC123",
        "statusText": "In port Stockholm",
        "origin": "Stockholm",
        "destination": "Hongkong",
        "eta": "2009-03-12 12:00",
        "nextExpectedActivity": "Next expected activity is to load cargo onto voyage 0200T in New York",
        "events": [
          {
            "description": "Received in Hongkong, at 3/1/09 12:00 AM.",
            "expected": true
          },
          {
            "description": "Loaded onto voyage 0100S in Hongkong, at 3/2/09 12:00 AM.",
            "expected": false
          },
          {
            "description": "Unloaded off voyage 0100S in New York, at 3/5/09 12:00 AM.",
            "expected": false
          }
        ]
      }

#### POST /cargos
Books a cargo.

| URL Param | Description |
|:----------|:------------|
|origin=[string]|UN locode of the origin|
|destination=[string]|UN locode of the destination|
|arrivalDeadline=[timestamp]|Timestamp of the arrival deadline|

#### POST /cargos/:id/change_destination
Updates the route of a cargo with a new destination.

| URL Param | Description |
|:----------|:------------|
|destination=[string]|UN locode of the destination|

#### GET /cargos/:id/request_routes
Requests the possible routes for a booked cargo.

__Example:__

    [
      {
        "legs": [
          {
            "voyage": "S0001",
            "from": "SESTO",
            "to": "CNHKG",
            "loadTime": "2009-03-12 12:00",
            "unloadTime": "2009-03-23 14:50"
          },
		  {
            "voyage": "S0002",
            "from": "CNHKG",
            "to": "AUMEL",
            "loadTime": "2009-03-24 12:00",
            "unloadTime": "2009-03-30 11:20"
          }
        ]
      }
    ]

#### POST /cargos/:id/assign_to_route
Assigns the cargo to a route. Typically one of the routes returned by `/cargos/:id/request_routes`.

__Example:__

    {
      "legs": [
        {
          "voyage": "S0001",
          "from": "SESTO",
          "to": "CNHKG",
          "loadTime": "2009-03-12 12:00",
          "unloadTime": "2009-03-23 14:50"
        },
	    {
          "voyage": "S0002",
          "from": "CNHKG",
          "to": "AUMEL",
          "loadTime": "2009-03-24 12:00",
          "unloadTime": "2009-03-30 11:20"
        }
      ]
    }

### Locations

#### GET /locations
Returns a list of the registered locations.

__Example:__

    [
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
      }
    ]

## Copyright

Copyright © 2014 Marcus Olsson. See [LICENSE](LICENSE) for details.
