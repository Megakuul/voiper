package empty

type Body struct{}

func NewBody() *Body {
	return &Body{}
}

func (b *Body) Type() string {
	return ""
}

func (b *Body) Length() int {
	return 0
}

func (b *Body) Content() string {
	return ""
}
