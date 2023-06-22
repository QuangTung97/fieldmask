package fields

type computeOptions struct {
	maxFields int
	maxDepth  int
}

func newComputeOptions(options []Option) *computeOptions {
	opts := &computeOptions{
		maxFields: 1000,
		maxDepth:  5,
	}
	for _, fn := range options {
		fn(opts)
	}
	return opts
}

// Option ...
type Option func(opts *computeOptions)

// WithMaxFields ...
func WithMaxFields(max int) Option {
	return func(opts *computeOptions) {
		opts.maxFields = max
	}
}

// WithMaxFieldDepth ...
func WithMaxFieldDepth(depth int) Option {
	return func(opts *computeOptions) {
		opts.maxDepth = depth
	}
}
