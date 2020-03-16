package util

type Op struct {
	nameLike  string
	user      string
	namespace string
	group     string
}

func (o Op) NameLike() string  { return o.nameLike }
func (o Op) User() string      { return o.user }
func (o Op) Group() string     { return o.group }
func (o Op) Namespace() string { return o.namespace }

type OpOption func(*Op)

func WithNameLike(s string) OpOption  { return func(op *Op) { op.nameLike = s } }
func WithGroup(s string) OpOption     { return func(op *Op) { op.group = s } }
func WithUser(s string) OpOption      { return func(op *Op) { op.user = s } }
func WithNamespace(s string) OpOption { return func(op *Op) { op.namespace = s } }

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
