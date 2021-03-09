package loadgenerator

type LoadGenerator interface {
	Start() error
	Stop() error
	SetUsers(users int) error
}
