package shared

// Entity is used to compare entities by identity.
type Entity interface {
	SameIdentity(Entity) bool
}

// ValueObject is used to compare value object by value.
type ValueObject interface {
	SameValue(ValueObject) bool
}
