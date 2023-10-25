package fields

type computeOptions struct {
	maxComponentLen int
	maxFields       int
	maxDepth        int
	limitedToFields []string
}

func newComputeOptions(options []Option) *computeOptions {
	opts := &computeOptions{
		maxFields:       1000,
		maxDepth:        5,
		maxComponentLen: 128,
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

// WithMaxFieldComponentLength ...
func WithMaxFieldComponentLength(maxLength int) Option {
	return func(opts *computeOptions) {
		opts.maxComponentLen = maxLength
	}
}

// WithLimitedToFields ...
func WithLimitedToFields(limitedTo []string) Option {
	return func(opts *computeOptions) {
		opts.limitedToFields = limitedTo
	}
}
