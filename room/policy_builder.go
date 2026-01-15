package room

type PoliciesBuilder struct {
	p PolicySet
}

func NewPolicies() *PoliciesBuilder {
	return &PoliciesBuilder{p: PolicySet{}}
}

func (p PolicySet) Clone() PolicySet {
	out := PolicySet{}
	if len(p.Join) > 0 {
		out.Join = append([]JoinPolicy(nil), p.Join...)
	}
	if len(p.Say) > 0 {
		out.Say = append([]SayPolicy(nil), p.Say...)
	}
	return out
}

func (base PolicySet) Merge(override PolicySet) PolicySet {
	out := base.Clone()
	out.Join = append(out.Join, override.Join...)
	out.Say = append(out.Say, override.Say...)
	return out
}

func FromPolicies(base PolicySet) *PoliciesBuilder {
	return &PoliciesBuilder{p: base.Clone()}
}

// Adders
func (b *PoliciesBuilder) WithJoin(p ...JoinPolicy) *PoliciesBuilder {
	b.p.Join = append(b.p.Join, p...)
	return b
}

func (b *PoliciesBuilder) WithSay(p ...SayPolicy) *PoliciesBuilder {
	b.p.Say = append(b.p.Say, p...)
	return b
}

func (b *PoliciesBuilder) Build() PolicySet {
	return b.p.Clone()
}
