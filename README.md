# GoDDD [![wercker status](https://app.wercker.com/status/12b6c3ee4ea86efc7fd5184b589916b7/s "wercker status")](https://app.wercker.com/project/bykey/12b6c3ee4ea86efc7fd5184b589916b7)

This is an attempt to port the [DDD Sample App](http://dddsample.sourceforge.net/) to Go. The purpose is to explore how to write idiomatic Go applications using Domain-Driven Design.

I will update this README along the way, as I gain more insights.

## Design

The purpose of this sample is because I wanted to learn Go and to see how well it can be used with Domain Driven Design. When faced with design challenges, I will lean towards the more idiomatic solution.

### Equality

In Go, two struct values are equal if their corresponding non-blank fields are equal. You cannot overload equality for structs and there is no standard interface for equality. The question is rather how to implement entities. You can create your own interface but it will still be tempting to use the == operator.

The current implementation uses the _Equaler_ interface from the [golang FAQ](https://golang.org/doc/faq#t_and_equal_interface).

Read more about [comparison operators](http://golang.org/ref/spec#Comparison_operators).

### Immutability

Go does not support means of creating a immutable struct. All exported fields can be altered after creation. It is however possible to use interfaces to handle modification of structs.

    type ValueObject interface {
         Name() string
    }

    type valueObject struct {
         name string
    }

    func (v *valueObject) Name() {
         return v.name
    }

    func NewValueObject(s string) ValueObject {
         return valueObject {name: s}
    }

Since the struct starts with a lowercase, it will not be exported outside the package. This however, does not prevent internal functions to modify the state of the value object after it has been created.

[Read more](https://groups.google.com/forum/#!topic/golang-nuts/BnjG3N77Ico) about immutable objects in this forum thread.

### Other thoughts ...

- How can we use the zero-initialization idiom effectively in DDD? What does a zero-initialized Itinerary mean?
- Concurrency is one area where Go shines, but initial thought is to keep it out of the domain model. This might be interesting if concurrency is a explicit part of the model.

## REST API

The application exposes a REST API using [Martini](https://github.com/go-martini/martini).

### Setup

1. Make sure you have your `$GOPATH` environment set.

    `export GOPATH=$HOME/go`

2. Get the latest version by running:

    `go get -u github.com/marcusolsson/goddd`

3. Navigate into the __server__ directory and run it by typing:

    `go run cmd/local.go`

_Note:_ The `cmd/appengine.go` is used to deploy it on Google App Engine. If you want to deploy it to your own project, edit the `app.yaml` to point to your own project. Read more about [deploying Go applications to App Engine](https://cloud.google.com/appengine/docs/go/).

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
