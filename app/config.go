package app

type Config struct {
	Name      string
	Version   string
	Env       string
	Host      string
	HostAdmin string
	Listener  ConfigListener
}

type ConfigListener struct {
	Host          string
	HttpPort      int32
	HttpAdminPort int32
	GrpcPort      int32
}
