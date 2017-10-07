package iface

const (
	// DefaultPort is the default port used by sarga clients.
	DefaultPort = 6778
)

type CommonArgs struct {
	Type string

	Port  int
	IP    string
	Proto Proto
}
