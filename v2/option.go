package extmap

type Option struct {
	Split bool
}

func defaultOption() *Option {
	return &Option{Split: true}
}

type OptionFunc func(op *Option)
