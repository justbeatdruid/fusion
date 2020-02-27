package util

type Op struct {
	nameLike string
	group    string
}

func (o Op) NameLike() string { return o.nameLike }
func (o Op) Group() string    { return o.group }

type OpOption func(*Op)

func WithNameLike(s string) OpOption { return func(op *Op) { op.nameLike = s } }
func WithGroup(s string) OpOption    { return func(op *Op) { op.group = s } }

func OpList(opts ...OpOption) Op {
	ret := Op{}
	ret.applyOpts(opts)
	return ret
}

func (op *Op) applyOpts(opts []OpOption) {
	for _, opt := range opts {
		opt(op)
	}
}
