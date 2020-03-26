package util

type Op struct {
	nameLike  string
	user      string
	namespace string
	group     string
	restriction string
	trafficcontrol string
}

func (o Op) NameLike() string  { return o.nameLike }
func (o Op) User() string      { return o.user }
func (o Op) Group() string     { return o.group }
func (o Op) Namespace() string { return o.namespace }
func (o Op) Restriction() string { return o.restriction }
func (o Op) Trafficcontrol() string { return o.trafficcontrol }

type OpOption func(*Op)

func WithNameLike(s string) OpOption  { return func(op *Op) { op.nameLike = s } }
func WithGroup(s string) OpOption     { return func(op *Op) { op.group = s } }
func WithUser(s string) OpOption      { return func(op *Op) { op.user = s } }
func WithNamespace(s string) OpOption { return func(op *Op) { op.namespace = s } }
func WithRestriction(s string) OpOption { return func(op *Op) { op.restriction = s } }
func WithTrafficcontrol(s string) OpOption { return func(op *Op) { op.trafficcontrol = s } }

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
