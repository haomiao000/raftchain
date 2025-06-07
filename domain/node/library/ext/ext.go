package ext

type RaftConfig struct {
	NodeID               int
	GrpcListenAddress    string
	Peers                map[int]string
	ElectionTimeoutMsMin int
	ElectionTimeoutMsMax int
	HeartbeatIntervalMs  int
}
