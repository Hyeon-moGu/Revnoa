package collector

type Tailer interface {
	Start() error
	Stop()
}
