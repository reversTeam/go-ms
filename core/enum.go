package core

// Define the GRPC server state in ENUM
type GoMsServerState int

const (
	Init      GoMsServerState = 0
	Boot      GoMsServerState = 1
	Listen    GoMsServerState = 2
	Ready     GoMsServerState = 3
	Error     GoMsServerState = 4
	Gracefull GoMsServerState = 5
	Stop      GoMsServerState = 6
)
