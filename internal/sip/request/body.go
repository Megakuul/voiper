package request

type Body interface {
	Type() string
	Length() int
	Content() string
}
