package discovery

type KV interface {
	Get(key string) ([]byte, error)
	Set(key string, value string) error
	Delete(key string) error
	List(key string) (map[string][]byte, error)
}
