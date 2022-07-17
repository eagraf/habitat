package sources

type PermissionsManager interface {
	CheckCanRead(ReadRequest) bool
	CheckCanWrite(WriteRequest) bool

	/* TODO: refine these
	ExtendCanRead()
	ExtendCanWrite()

	RevokeCanRead()
	RevokeCanWrite()
	*/
}

/*
	TODO: evolve this. for now the main focus is on data storage/schema, so allow everything.
*/

type BasicPermissionsManager struct{}

func (B *BasicPermissionsManager) CheckCanRead(ReadRequest) bool {
	return true
}

func (B *BasicPermissionsManager) CheckCanWrite(WriteRequest) bool {
	return true
}

func NewBasicPermissionsManager() *BasicPermissionsManager {
	return &BasicPermissionsManager{}
}
