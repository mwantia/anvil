package forge

import "context"

// Client is the surface anvil uses to talk to a Forge daemon.
type Client interface {
	// reads
	Sessions(ctx context.Context) ([]Session, error)
	Session(ctx context.Context, idOrName string) (Session, error)
	// Log returns messages for sessionID. ref selects the branch to walk;
	// empty string walks from HEAD (default behaviour).
	Log(ctx context.Context, sessionID, ref string) ([]Message, error)
	Refs(ctx context.Context, sessionID string) ([]Ref, error)
	Resources(ctx context.Context, scope string) ([]Resource, error)
	ResourceDetail(ctx context.Context, path, name string) (Resource, error)
	System(ctx context.Context) (System, error)

	// mutations
	NewSession(ctx context.Context, name, model string) (Session, error)
	CloneSession(ctx context.Context, from, to string) (Session, error)
	ArchiveSession(ctx context.Context, name string) error
	DeleteSession(ctx context.Context, name string) error

	Checkout(ctx context.Context, sessionID, ref string) error
	Branch(ctx context.Context, sessionID, name, fromHash string) error
	DeleteRef(ctx context.Context, sessionID, ref string) error
	Merge(ctx context.Context, sessionID, ref string) error

	// EditFork creates an edit-… ref off fromHash and returns the message.
	EditFork(ctx context.Context, sessionID, fromHash, newBody string) (Message, error)
}
