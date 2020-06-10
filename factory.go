package backoff

// A Factory creates and returns a new backoff policy
type Factory interface {
	NewBackOff() BackOff
}

// FactoryFunc is a function that returns a new backoff policy
type FactoryFunc func() BackOff

// NewBackOff implements Factory, allowing a FactoryFunc to be passed to anyywhere that accepts a Factory
func (ff FactoryFunc) NewBackOff() BackOff {
	return ff()
}
