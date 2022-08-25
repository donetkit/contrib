package discovery

type Discovery interface {
	Register() error
	Deregister() error
}
