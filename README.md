# GoDDD #
This is an attempt to port the [DDD Sample App](http://dddsample.sourceforge.net/) to Go. The purpose is to explore how to write idiomatic Go applications using Domain-Driven Design.

### Equality

In Go, two struct values are equal if their corresponding non-blank fields are equal. You cannot overload equality for structs and there is no standard interface for equality. This makes it more difficult to implement entities. You can create your own interface but it will still be tempting to use the == operator.

The current implementation uses the Equaler interface from the [golang FAQ](https://golang.org/doc/faq#t_and_equal_interface).

Read more about [comparison operators](http://golang.org/ref/spec#Comparison_operators).

### Value Objects

Two structs will be equal if the fields are equal. Structs in Go however, cannot be made immutable. This can be solved with having a NewLocation() function return an interface that hides a package-private location struct, but it is not really idiomatic Go code.

### Bounded contexts

Packages in Go seem to map pretty nicely into bounded contexts.