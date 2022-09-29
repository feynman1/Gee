package algorithm

type Common interface {
	Get(key string) string
	Add(key string, value string) string
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}
