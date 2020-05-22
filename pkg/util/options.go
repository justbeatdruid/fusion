package util

type Op struct {
	nameLike       string
	user           string
	namespace      string
	group          string
	restriction    string
	trafficcontrol string
	topic          string
	available      string
	typpe          string
	stype          string
}

func (o Op) NameLike() string       { return o.nameLike }
func (o Op) User() string           { return o.user }
func (o Op) Group() string          { return o.group }
func (o Op) Namespace() string      { return o.namespace }
func (o Op) Restriction() string    { return o.restriction }
func (o Op) Trafficcontrol() string { return o.trafficcontrol }
func (o Op) Topic() string          { return o.topic }
func (o Op) Available() string      { return o.available }
func (o Op) Type() string           { return o.typpe }
func (o Op) Stype() string          { return o.stype }

type OpOption func(*Op)

func WithNameLike(s string) OpOption       { return func(op *Op) { op.nameLike = s } }
func WithGroup(s string) OpOption          { return func(op *Op) { op.group = s } }
func WithUser(s string) OpOption           { return func(op *Op) { op.user = s } }
func WithNamespace(s string) OpOption      { return func(op *Op) { op.namespace = s } }
func WithRestriction(s string) OpOption    { return func(op *Op) { op.restriction = s } }
func WithTrafficcontrol(s string) OpOption { return func(op *Op) { op.trafficcontrol = s } }
func WithTopic(s string) OpOption          { return func(op *Op) { op.topic = s } }
func WithAvailable(s string) OpOption      { return func(op *Op) { op.available = s } }
func WithType(s string) OpOption           { return func(op *Op) { op.typpe = s } }
func WithStype(s string) OpOption          { return func(op *Op) { op.stype = s } }

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
