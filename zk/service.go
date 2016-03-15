package zk

// service connection interface
type ServiceConn interface{}

// when got address deleted event call ServiceCli.OnDisable and set cli invalid, will skip the cli on GetConn
// when got address recovery event call ServiceCli.OnRecover and set cli valid
type ServiceCli interface {
	GetConn() ServiceConn
	Status() string
	Close()
	OnDisable()
	OnEnable()
}

// when got new address event call InitCli
type Service interface {
	InitCli(addr string, arg interface{}) ServiceCli
}