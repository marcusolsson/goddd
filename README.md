# GoDDD 

[![Build Status](https://travis-ci.org/marcusolsson/goddd.svg?branch=master)](https://travis-ci.org/marcusolsson/goddd)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/marcusolsson/goddd)
[![Go Report Card](https://goreportcard.com/badge/github.com/marcusolsson/goddd)](https://goreportcard.com/report/github.com/marcusolsson/goddd)
[![License MIT](https://img.shields.io/badge/license-MIT-lightgrey.svg?style=flat)](LICENSE)
![stability-unstable](https://img.shields.io/badge/stability-unstable-yellow.svg)

This is an attempt to port the [DDD Sample App](https://github.com/citerus/dddsample-core) to idiomatic Go. This project aims to:

- Demonstrate how the tactical design patterns from Domain Driven Design may be implemented in Go. 
- Serve as an example of a modern production-ready enterprise application.

### Important note

This project is intended for inspirational purposes and should **not** be considered a tutorial, guide or best-practice neither how to implement Domain Driven Design nor enterprise applications in Go. Make sure you adapt the code and ideas to the requirements of your own application.

## Porting from Java

The original application is written in Java and much thought has been given to the domain model, code organization and is intended to be an example of what you might find in an enterprise system.

I started out by first rewriting the original application, as is, in Go. The result was hardly idiomatic Go and I have since tried to refactor towards something that is true to the Go way. This means that you will still find oddities due to the application's Java heritage. If you do, please let me know so that we can weed out the remaining Java.

## Running the application

Start the application on port 8080 (or whatever the `PORT` variable is set to).

```
go run main.go -inmem
```

If you only want to try it out, this is enough. If you are looking for full functionality, you will need to have a [routing service](https://github.com/marcusolsson/pathfinder) running and start the application with `ROUTINGSERVICE_URL` (default: `http://localhost:7878`).

### Docker

You can also run the application using Docker.

```
# Start routing service
docker run --name some-pathfinder marcusolsson/pathfinder

# Start application
docker run --name some-goddd \
  --link some-pathfinder:pathfinder \
  -p 8080:8080 \
  -e ROUTINGSERVICE_URL=http://pathfinder:8080 \
  marcusolsson/goddd /goddd -inmem
```

... or if you're using Docker Compose:

```
docker-compose up
```

## Try it!

```
# Check out the sample cargos
curl localhost:8080/booking/v1/cargos

# Book new cargo
curl localhost:8080/booking/v1/cargos -d '{"origin": "SESTO", "destination": "FIHEL", "arrival_deadline": "2016-03-21T19:50:24Z"}'

# Request possible routes for sample cargo ABC123
curl localhost:8080/booking/v1/cargos/ABC123/request_routes
```

## Contributing

If you want to fork the repository, follow these step to avoid having to rewrite the import paths.

```shell
go get github.com/marcusolsson/goddd
cd $GOPATH/src/github.com/marcusolsson/goddd
git remote add fork git://github.com:<yourname>/goddd.git

# commit your changes

git push fork
```

For more information, read [this](http://blog.campoy.cat/2014/03/github-and-go-forking-pull-requests-and.html).

## Additional resources

### For watching

- [Building an Enterprise Service in Go](https://www.youtube.com/watch?v=twcDf_Y2gXY) at Golang UK Conference 2016

### For reading

- [Domain Driven Design in Go: Part 1](http://www.citerus.se/go-ddd)
- [Domain Driven Design in Go: Part 2](http://www.citerus.se/part-2-domain-driven-design-in-go)
- [Domain Driven Design in Go: Part 3](http://www.citerus.se/part-3-domain-driven-design-in-go)

### Related projects

The original application uses a external routing service to demonstrate the use of _bounded contexts_. For those who are interested, I have ported this service as well:

[pathfinder](https://github.com/marcusolsson/pathfinder)

To accompany this application, there is also an AngularJS-application to demonstrate the intended use-cases.

[dddelivery-angularjs](https://github.com/marcusolsson/dddelivery-angularjs)

Also, if you want to learn more about Domain Driven Design, I encourage you to take a look at the [Domain Driven Design](http://www.amazon.com/Domain-Driven-Design-Tackling-Complexity-Software/dp/0321125215) book by Eric Evans.

