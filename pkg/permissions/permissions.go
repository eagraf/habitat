package permissions

type AppPermissionsManager interface {
	CheckApp(token string, communityID string) bool
}

type SourcesPermissionsManager interface {
	CheckCanRead(token string, sourceID string) bool
	CheckCanWrite(token string, sourceID string) bool

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

func (B *BasicPermissionsManager) CheckCanRead(token string, sourceID string) bool {
	return true
}

func (B *BasicPermissionsManager) CheckCanWrite(token string, sourceID string) bool {
	return true
}

func (B *BasicPermissionsManager) CheckApp(token string, communityId string) bool {
	return true
}

func NewBasicPermissionsManager() *BasicPermissionsManager {
	return &BasicPermissionsManager{}
}
