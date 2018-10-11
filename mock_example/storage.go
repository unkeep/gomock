package mock_example

type storage interface {
	GetValue(key string) (int, error)
	SetValue(key string, value int) error
}
