package forge

import (
	"context"
	"fmt"
	"sort"
	"strings"

	v2 "github.com/mwantia/forge-sdk/pkg/api/v2"
	sdkrefs "github.com/mwantia/forge-sdk/pkg/api/v2/refs"
	sdkresources "github.com/mwantia/forge-sdk/pkg/api/v2/resources"
	sdksessions "github.com/mwantia/forge-sdk/pkg/api/v2/sessions"
	sdksystem "github.com/mwantia/forge-sdk/pkg/api/v2/system"
	"github.com/mwantia/forge-sdk/pkg/api/v2/transport"
)

// SDKClient implements Client using the forge-sdk v2 HTTP transport.
type SDKClient struct {
	api *v2.ForgeApi
}

// NewSDKClient returns a Client backed by the forge-sdk v2 API.
func NewSDKClient(addr, token string) *SDKClient {
	return &SDKClient{api: v2.NewApi(addr, token)}
}

// ─── mapping helpers ──────────────────────────────────────────────────────────

func sessionFromSDK(s sdksessions.SessionMetadata) Session {
	name := s.Name
	if name == "" {
		n := min(len(s.ID), 8)
		name = s.ID[:n]
	}

	plugins := "all"
	if len(s.Plugins) > 0 {
		plugins = strings.Join(s.Plugins, ",")
	}

	var tokIn, tokOut int
	var cost float64
	if s.Usage != nil {
		tokIn = s.Usage.InputTokens
		tokOut = s.Usage.OutputTokens
		cost = s.Usage.TotalCost
	}

	return Session{
		ID:        s.ID,
		Name:      name,
		Title:     s.Title,
		Desc:      s.Description,
		Parent:    s.Parent,
		Plugins:   plugins,
		Model:     s.Model,
		Created:   s.CreatedAt,
		Updated:   s.UpdatedAt,
		Archived:  s.ArchivedAt != nil,
		TokensIn:  tokIn,
		TokensOut: tokOut,
		Cost:      cost,
		Counts:    map[string]int{},
	}
}

func messageFromSDK(m sdksessions.Message) Message {
	lines := strings.Split(strings.TrimSpace(m.Content), "\n")
	preview := ""

	if len(lines) > 0 {
		preview = lines[0]
		r := []rune(preview)
		if len(r) > 80 {
			preview = string(r[:77]) + "…"
		}
	}

	var tokIn, tokOut int
	if m.Usage != nil {
		tokIn = m.Usage.InputTokens
		tokOut = m.Usage.OutputTokens
	}

	return Message{
		Hash:    shortH(m.Hash, 12),
		Role:    Role(m.Role),
		Date:    m.CreatedAt,
		Preview: preview,
		Body:    lines,
		TokIn:   tokIn,
		TokOut:  tokOut,
	}
}

func refsFromSDK(resp sdkrefs.RefsListResponse) []Ref {
	var out []Ref
	for sym, target := range resp.Symrefs {
		if sym == "HEAD" {
			hash := resp.Refs["HEAD"]
			out = append(out, Ref{Ref: "HEAD", Target: target, Hash: hash, IsHead: true})
		}
	}

	for name, hash := range resp.Refs {
		if _, isSymref := resp.Symrefs[name]; isSymref {
			continue
		}
		out = append(out, Ref{Ref: name, Hash: hash})
	}

	sort.Slice(out, func(i, j int) bool {
		return refSortKey(out[i]) < refSortKey(out[j])
	})

	return out
}

func resourceFromSDK(r *sdkresources.Resource) Resource {
	res := Resource{
		Path:    r.Path,
		Name:    r.ID,
		Scope:   scopeFromPath(r.Path),
		Updated: r.CreatedAt,
	}

	if r.Metadata != nil {
		if v, ok := r.Metadata["mime"].(string); ok && v != "" {
			res.MIME = v
		} else if v, ok := r.Metadata["content_type"].(string); ok && v != "" {
			res.MIME = v
		}
		if v, ok := r.Metadata["summary"].(string); ok {
			res.Summary = v
		}
	}

	if len(r.Content) > 0 {
		res.Size = fmtSize(len(r.Content))
	}

	return res
}

func fmtSize(n int) string {
	switch {
	case n >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))

	case n >= 1024:
		return fmt.Sprintf("%.1f kB", float64(n)/1024)

	case n > 0:
		return fmt.Sprintf("%d B", n)
	}

	return ""
}

func scopeFromPath(path string) Scope {
	switch {
	case strings.HasPrefix(path, "/sessions/"):
		return ScopeSession

	case strings.HasPrefix(path, "/archives/"):
		return ScopeArchive

	case strings.HasPrefix(path, "/global/"):
		return ScopeGlobal
	}

	return ScopeGlobal
}

func shortH(h string, n int) string {
	if len(h) <= n {
		return h
	}

	return h[:n]
}

func refSortKey(r Ref) string {
	switch {
	case r.IsHead:
		return "0"

	case r.Ref == "main":
		return "1"

	case strings.HasPrefix(r.Ref, "edit-"):
		return "2" + r.Ref

	case strings.HasPrefix(r.Ref, "fork-"):
		return "3" + r.Ref
	}

	return "4" + r.Ref
}

// ─── Client implementation ────────────────────────────────────────────────────

func (c *SDKClient) Sessions(ctx context.Context) ([]Session, error) {
	resp, err := c.api.Sessions.List(ctx, sdksessions.SessionsListRequest{
		Pagination: transport.Pagination{Limit: 200},
	})

	if err != nil {
		return nil, err
	}

	out := make([]Session, len(resp.Sessions))
	for i, s := range resp.Sessions {
		out[i] = sessionFromSDK(s)
	}

	return out, nil
}

func (c *SDKClient) Session(ctx context.Context, idOrName string) (Session, error) {
	resp, err := c.api.Sessions.Get(ctx, sdksessions.SessionsGetRequest{ID: idOrName})
	if err != nil {
		return Session{}, err
	}

	return sessionFromSDK(resp.Session), nil
}

func (c *SDKClient) Log(ctx context.Context, sessionID string) ([]Message, error) {
	resp, err := c.api.Sessions.ListMessages(ctx, sdksessions.SessionsListMessagesRequest{
		SessionID:  sessionID,
		Pagination: transport.Pagination{Limit: 500},
	})

	if err != nil {
		return nil, err
	}

	msgs := make([]Message, len(resp.Messages))
	for i, m := range resp.Messages {
		msgs[i] = messageFromSDK(m)
	}

	// Correlate ref names onto messages via hash lookup
	refsResp, _ := c.api.Refs.List(ctx, sdkrefs.RefsListRequest{SessionID: sessionID})
	hashToRefs := map[string][]string{}

	for name, hash := range refsResp.Refs {
		sh := shortH(hash, 12)
		hashToRefs[sh] = append(hashToRefs[sh], name)
	}

	for i := range msgs {
		if rs, ok := hashToRefs[msgs[i].Hash]; ok {
			msgs[i].Refs = rs
		}
	}

	return msgs, nil
}

func (c *SDKClient) Refs(ctx context.Context, sessionID string) ([]Ref, error) {
	resp, err := c.api.Refs.List(ctx, sdkrefs.RefsListRequest{SessionID: sessionID})
	if err != nil {
		return nil, err
	}

	return refsFromSDK(resp), nil
}

func (c *SDKClient) Resources(ctx context.Context, scope string) ([]Resource, error) {
	path := "/"
	switch scope {
	case "session":
		path = "/sessions/"

	case "archive":
		path = "/archives/"

	case "global":
		path = "/global/"
	}

	resources, err := c.listRecursive(ctx, path, 3, 0)
	if err != nil {
		return nil, err
	}

	if scope == "" || scope == "all" {
		return resources, nil
	}

	var out []Resource
	for _, r := range resources {
		if string(r.Scope) == scope {
			out = append(out, r)
		}
	}

	return out, nil
}

func (c *SDKClient) listRecursive(ctx context.Context, path string, maxDepth, depth int) ([]Resource, error) {
	resp, err := c.api.Resources.List(ctx, sdkresources.ResourcesListRequest{Path: path})
	if err != nil {
		return nil, err
	}

	var out []Resource
	for _, r := range resp.Resources {
		if r.Type == "dir" && depth < maxDepth {
			children, _ := c.listRecursive(ctx, r.Path, maxDepth, depth+1)
			out = append(out, children...)
		} else if r.Type != "dir" {
			out = append(out, resourceFromSDK(r))
		}
	}

	return out, nil
}

func (c *SDKClient) ResourceDetail(ctx context.Context, path, name string) (Resource, error) {
	getResp, err := c.api.Resources.Get(ctx, sdkresources.ResourcesGetRequest{Path: path, Name: name})
	if err != nil {
		return Resource{}, err
	}

	res := resourceFromSDK(&getResp.Resource)
	if len(getResp.Resource.Content) > 0 {
		res.Size = fmtSize(len(getResp.Resource.Content))
	}

	histResp, err := c.api.Resources.History(ctx, sdkresources.ResourcesHistoryRequest{Path: path, Name: name})
	if err != nil {
		return res, nil // return partial result on history error
	}

	res.Versions = len(histResp.History)
	if len(histResp.History) > 0 {
		res.HEAD = shortH(histResp.History[0].Hash, 8)

		for i, rev := range histResp.History {
			delta := ""
			if rev.Metadata != nil {
				if d, ok := rev.Metadata["delta"].(string); ok {
					delta = d
				}
			}

			res.History = append(res.History, ResourceRev{
				Version: res.Versions - i,
				Hash:    shortH(rev.Hash, 8),
				Date:    rev.CreatedAt,
				Delta:   delta,
			})
		}
	}

	return res, nil
}

func (c *SDKClient) System(ctx context.Context) (System, error) {
	status, _ := c.api.Resources.Status(ctx)
	objCount, _ := c.api.System.DagObjectsCount(ctx, sdksystem.DagObjectsCountRequest{})

	var sys System
	sys.Storage.Backend = status.Backend
	sys.Storage.Objects = objCount.Count
	sys.Agent.HTTP = c.api.GetAddress()

	return sys, nil
}

func (c *SDKClient) NewSession(ctx context.Context, name, model string) (Session, error) {
	resp, err := c.api.Sessions.Create(ctx, sdksessions.SessionsCreateRequest{
		Name:  name,
		Model: model,
	})

	if err != nil {
		return Session{}, err
	}

	return sessionFromSDK(resp.Session), nil
}

func (c *SDKClient) CloneSession(ctx context.Context, from, _ string) (Session, error) {
	id, err := c.resolveSessionID(ctx, from)
	if err != nil {
		return Session{}, err
	}

	resp, err := c.api.Sessions.Clone(ctx, sdksessions.SessionsCloneRequest{ID: id})
	if err != nil {
		return Session{}, err
	}

	return sessionFromSDK(resp.Session), nil
}

func (c *SDKClient) ArchiveSession(ctx context.Context, name string) error {
	id, err := c.resolveSessionID(ctx, name)
	if err != nil {
		return err
	}

	return c.api.Sessions.Archive(ctx, sdksessions.SessionsArchiveRequest{ID: id})
}

func (c *SDKClient) DeleteSession(ctx context.Context, name string) error {
	id, err := c.resolveSessionID(ctx, name)
	if err != nil {
		return err
	}

	return c.api.Sessions.Delete(ctx, sdksessions.SessionsDeleteRequest{ID: id})
}

func (c *SDKClient) Checkout(ctx context.Context, sessionID, branch string) error {
	_, err := c.api.Refs.Checkout(ctx, sdkrefs.RefsCheckoutRequest{
		SessionID: sessionID,
		Branch:    branch,
	})

	return err
}

func (c *SDKClient) Branch(ctx context.Context, sessionID, name, fromHash string) error {
	_, err := c.api.Refs.Create(ctx, sdkrefs.RefsCreateRequest{
		SessionID: sessionID,
		Name:      name,
		Hash:      fromHash,
	})

	return err
}

func (c *SDKClient) DeleteRef(ctx context.Context, sessionID, ref string) error {
	return c.api.Refs.Delete(ctx, sdkrefs.RefsDeleteRequest{
		SessionID: sessionID,
		Ref:       ref,
	})
}

func (c *SDKClient) Merge(_ context.Context, _, _ string) error {
	return nil // not directly supported
}

func (c *SDKClient) EditFork(ctx context.Context, sessionID, fromHash, _ string) (Message, error) {
	name := "edit-" + shortH(fromHash, 8)
	_, err := c.api.Refs.Create(ctx, sdkrefs.RefsCreateRequest{
		SessionID: sessionID,
		Name:      name,
		Hash:      fromHash,
	})

	if err != nil {
		return Message{}, err
	}

	return Message{Hash: shortH(fromHash, 12)}, nil
}

func (c *SDKClient) resolveSessionID(ctx context.Context, nameOrID string) (string, error) {
	if len(nameOrID) == 32 {
		return nameOrID, nil
	}

	ss, err := c.Sessions(ctx)
	if err != nil {
		return "", err
	}

	for _, s := range ss {
		if s.Name == nameOrID || s.ID == nameOrID {
			return s.ID, nil
		}
	}

	return "", fmt.Errorf("session not found: %s", nameOrID)
}
