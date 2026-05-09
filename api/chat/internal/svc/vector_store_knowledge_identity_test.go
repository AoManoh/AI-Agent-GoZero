package svc

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sashabaranov/go-openai"
)

func TestSaveKnowledgeBatchAssignsIncrementedSharedVersion(t *testing.T) {
	tx := &knowledgeVersionTx{t: t}
	store := &VectorStore{
		Pool:           &knowledgeVersionPool{tx: tx},
		OpenAIClient:   newKnowledgeEmbeddingClient(t),
		EmbeddingModel: "test-embedding",
	}

	err := store.SaveKnowledgeBatchForUserContextWithMeta(
		context.Background(),
		"Go 面试题",
		[]string{"第一块", "第二块"},
		7,
		"manual",
	)
	if err != nil {
		t.Fatalf("SaveKnowledgeBatchForUserContextWithMeta() error = %v", err)
	}
	if !tx.locked {
		t.Fatalf("expected document identity advisory lock")
	}
	if !tx.archived {
		t.Fatalf("expected old ready versions to be archived")
	}
	if !tx.committed {
		t.Fatalf("expected transaction commit")
	}
	if !reflect.DeepEqual(tx.insertVersions, []int64{2, 2}) {
		t.Fatalf("insert versions = %#v, want both chunks on version 2", tx.insertVersions)
	}
}

func TestKnowledgeDocumentIdentitySeparatesRepeatedUploads(t *testing.T) {
	ctx := context.Background()
	userID := int64(7)
	now := time.Date(2026, 5, 10, 9, 30, 0, 0, time.UTC)
	pool := &knowledgeIdentityPool{
		t:   t,
		now: now,
	}
	store := &VectorStore{Pool: pool}

	documents, err := store.ListKnowledgeDocuments(ctx, &userID, 10)
	if err != nil {
		t.Fatalf("ListKnowledgeDocuments() error = %v", err)
	}
	if len(documents) != 2 {
		t.Fatalf("ListKnowledgeDocuments() len = %d, want 2", len(documents))
	}
	if documents[0].Title != documents[1].Title {
		t.Fatalf("test fixture invalid: document titles should match")
	}
	if documents[0].Version == documents[1].Version {
		t.Fatalf("same-title repeated upload was collapsed into one version: %+v", documents)
	}
	if documents[0].Status != "ready" || documents[1].Status != "archived" {
		t.Fatalf("document statuses = %q/%q, want ready/archived", documents[0].Status, documents[1].Status)
	}

	document, chunks, err := store.LoadKnowledgeDocumentChunks(ctx, documents[0].DocumentID, &userID, 10)
	if err != nil {
		t.Fatalf("LoadKnowledgeDocumentChunks() error = %v", err)
	}
	if document.Version != 2 {
		t.Fatalf("loaded document version = %d, want 2", document.Version)
	}
	if len(chunks) != 2 {
		t.Fatalf("LoadKnowledgeDocumentChunks() len = %d, want 2", len(chunks))
	}
	for _, chunk := range chunks {
		if strings.Contains(chunk.Content, "v1") {
			t.Fatalf("loaded chunk from archived v1 batch: %+v", chunk)
		}
	}
}

type knowledgeVersionPool struct {
	tx *knowledgeVersionTx
}

func (p *knowledgeVersionPool) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return p.tx, nil
}

func (p *knowledgeVersionPool) Close() {}

func (p *knowledgeVersionPool) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("EXEC 0"), fmt.Errorf("unexpected pool Exec")
}

func (p *knowledgeVersionPool) Ping(context.Context) error {
	return nil
}

func (p *knowledgeVersionPool) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, fmt.Errorf("unexpected pool Query")
}

func (p *knowledgeVersionPool) QueryRow(context.Context, string, ...any) pgx.Row {
	return knowledgeRow{err: fmt.Errorf("unexpected pool QueryRow")}
}

type knowledgeVersionTx struct {
	t              *testing.T
	locked         bool
	archived       bool
	committed      bool
	insertVersions []int64
}

func (tx *knowledgeVersionTx) Begin(context.Context) (pgx.Tx, error) {
	return nil, fmt.Errorf("unexpected nested Begin")
}

func (tx *knowledgeVersionTx) Commit(context.Context) error {
	tx.committed = true
	return nil
}

func (tx *knowledgeVersionTx) Rollback(context.Context) error {
	return nil
}

func (tx *knowledgeVersionTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, fmt.Errorf("unexpected CopyFrom")
}

func (tx *knowledgeVersionTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults {
	return nil
}

func (tx *knowledgeVersionTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (tx *knowledgeVersionTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, fmt.Errorf("unexpected Prepare")
}

func (tx *knowledgeVersionTx) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	tx.t.Helper()
	compactSQL := compactKnowledgeSQL(sql)
	switch {
	case strings.Contains(compactSQL, "pg_advisory_xact_lock"):
		assertKnowledgeArgs(tx.t, args, "7:Go 面试题:manual")
		tx.locked = true
	case strings.HasPrefix(compactSQL, "UPDATE knowledge_base SET status = 'archived'"):
		assertKnowledgeArgs(tx.t, args, int64(7), "Go 面试题", "manual")
		tx.archived = true
	case strings.HasPrefix(compactSQL, "INSERT INTO knowledge_base"):
		if len(args) != 8 {
			tx.t.Fatalf("insert args len = %d, want 8", len(args))
		}
		if args[0] != "Go 面试题" || args[3] != int64(7) || args[4] != "manual" || args[5] != "private" {
			tx.t.Fatalf("unexpected insert args: %#v", args)
		}
		version, ok := args[6].(int64)
		if !ok {
			tx.t.Fatalf("insert version arg type = %T, want int64", args[6])
		}
		hash, ok := args[7].(string)
		if !ok || len(hash) != 64 {
			tx.t.Fatalf("insert content hash = %#v, want sha256 hex", args[7])
		}
		tx.insertVersions = append(tx.insertVersions, version)
	default:
		return pgconn.NewCommandTag("EXEC 0"), fmt.Errorf("unexpected tx Exec: %s", compactSQL)
	}

	return pgconn.NewCommandTag("EXEC 1"), nil
}

func (tx *knowledgeVersionTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, fmt.Errorf("unexpected tx Query")
}

func (tx *knowledgeVersionTx) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	tx.t.Helper()
	compactSQL := compactKnowledgeSQL(sql)
	if !strings.Contains(compactSQL, "SELECT coalesce(max(version), 0) + 1 FROM knowledge_base") {
		return knowledgeRow{err: fmt.Errorf("unexpected tx QueryRow: %s", compactSQL)}
	}
	assertKnowledgeArgs(tx.t, args, int64(7), "Go 面试题", "manual")
	return knowledgeRow{values: []any{int64(2)}}
}

func (tx *knowledgeVersionTx) Conn() *pgx.Conn {
	return nil
}

type knowledgeIdentityPool struct {
	t   *testing.T
	now time.Time
}

func (p *knowledgeIdentityPool) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, fmt.Errorf("unexpected BeginTx")
}

func (p *knowledgeIdentityPool) Close() {}

func (p *knowledgeIdentityPool) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("EXEC 0"), fmt.Errorf("unexpected Exec")
}

func (p *knowledgeIdentityPool) Ping(context.Context) error {
	return nil
}

func (p *knowledgeIdentityPool) Query(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
	p.t.Helper()
	compactSQL := compactKnowledgeSQL(sql)
	switch {
	case strings.Contains(compactSQL, "FROM knowledge_base WHERE") && strings.Contains(compactSQL, "GROUP BY user_id, title, source, version"):
		if strings.Contains(compactSQL, "status = 'ready'") {
			p.t.Fatalf("document list query should include archived versions, got SQL: %s", compactSQL)
		}
		assertKnowledgeArgs(p.t, args, int64(7), 10)
		return &knowledgeRows{
			rows: [][]any{
				{int64(201), int64(7), "Go 面试题", "manual", "private", "ready", int64(2), int64(2), p.now.Add(time.Minute), p.now.Add(2 * time.Minute), "v2 第一块"},
				{int64(101), int64(7), "Go 面试题", "manual", "private", "archived", int64(1), int64(2), p.now.Add(-time.Hour), p.now, "v1 第一块"},
			},
		}, nil
	case strings.Contains(compactSQL, "SELECT id, title, content, created_at FROM knowledge_base"):
		if !strings.Contains(compactSQL, "source = $3 AND version = $4") {
			p.t.Fatalf("chunk query does not constrain source/version: %s", compactSQL)
		}
		assertKnowledgeArgs(p.t, args, int64(7), "Go 面试题", "manual", int64(2), 10)
		return &knowledgeRows{
			rows: [][]any{
				{int64(201), "Go 面试题", "v2 第一块", p.now.Add(time.Minute)},
				{int64(202), "Go 面试题", "v2 第二块", p.now.Add(2 * time.Minute)},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unexpected query: %s", compactSQL)
	}
}

func (p *knowledgeIdentityPool) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	p.t.Helper()
	compactSQL := compactKnowledgeSQL(sql)
	switch {
	case strings.Contains(compactSQL, "SELECT user_id, title, coalesce(source, ''), coalesce(version, 1) FROM knowledge_base"):
		if strings.Contains(compactSQL, "status = 'ready'") {
			p.t.Fatalf("document identity lookup should allow archived rows, got SQL: %s", compactSQL)
		}
		assertKnowledgeArgs(p.t, args, int64(201), int64(7))
		return knowledgeRow{values: []any{int64(7), "Go 面试题", "manual", int64(2)}}
	case strings.Contains(compactSQL, "FROM knowledge_base WHERE user_id = $1 AND title = $2 AND source = $3 AND version = $4"):
		assertKnowledgeArgs(p.t, args, int64(7), "Go 面试题", "manual", int64(2))
		return knowledgeRow{values: []any{
			int64(201), int64(7), "Go 面试题", "manual", "private", "ready", int64(2), int64(2),
			p.now.Add(time.Minute), p.now.Add(2 * time.Minute), "v2 第一块",
		}}
	default:
		return knowledgeRow{err: fmt.Errorf("unexpected query row: %s", compactSQL)}
	}
}

type knowledgeRow struct {
	values []any
	err    error
}

func (r knowledgeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	return scanKnowledgeValues(r.values, dest...)
}

type knowledgeRows struct {
	rows   [][]any
	idx    int
	closed bool
	err    error
}

func (r *knowledgeRows) Close() {
	r.closed = true
}

func (r *knowledgeRows) Err() error {
	return r.err
}

func (r *knowledgeRows) CommandTag() pgconn.CommandTag {
	return pgconn.NewCommandTag("SELECT 0")
}

func (r *knowledgeRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (r *knowledgeRows) Next() bool {
	if r.closed {
		return false
	}
	if r.idx >= len(r.rows) {
		r.Close()
		return false
	}
	r.idx++
	return true
}

func (r *knowledgeRows) Scan(dest ...any) error {
	if r.idx == 0 || r.idx > len(r.rows) {
		return fmt.Errorf("Scan called without current row")
	}
	return scanKnowledgeValues(r.rows[r.idx-1], dest...)
}

func (r *knowledgeRows) Values() ([]any, error) {
	if r.idx == 0 || r.idx > len(r.rows) {
		return nil, fmt.Errorf("Values called without current row")
	}
	return r.rows[r.idx-1], nil
}

func (r *knowledgeRows) RawValues() [][]byte {
	return nil
}

func (r *knowledgeRows) Conn() *pgx.Conn {
	return nil
}

func scanKnowledgeValues(values []any, dest ...any) error {
	if len(values) != len(dest) {
		return fmt.Errorf("scan value count = %d, dest count = %d", len(values), len(dest))
	}

	for i, value := range values {
		if err := assignKnowledgeValue(dest[i], value); err != nil {
			return fmt.Errorf("scan column %d: %w", i, err)
		}
	}
	return nil
}

func assignKnowledgeValue(dest any, value any) error {
	switch d := dest.(type) {
	case *int64:
		v, ok := value.(int64)
		if !ok {
			return fmt.Errorf("cannot assign %T to *int64", value)
		}
		*d = v
		return nil
	case *string:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("cannot assign %T to *string", value)
		}
		*d = v
		return nil
	case *time.Time:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("cannot assign %T to *time.Time", value)
		}
		*d = v
		return nil
	default:
		target := reflect.ValueOf(dest)
		if target.Kind() != reflect.Pointer || target.IsNil() {
			return fmt.Errorf("destination %T is not a non-nil pointer", dest)
		}
		source := reflect.ValueOf(value)
		if !source.Type().AssignableTo(target.Elem().Type()) {
			return fmt.Errorf("cannot assign %T to %T", value, dest)
		}
		target.Elem().Set(source)
		return nil
	}
}

func assertKnowledgeArgs(t *testing.T, got []any, want ...any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("query args = %#v, want %#v", got, want)
	}
}

func compactKnowledgeSQL(sql string) string {
	return strings.Join(strings.Fields(sql), " ")
}

func newKnowledgeEmbeddingClient(t *testing.T) *openai.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/embeddings") {
			t.Fatalf("unexpected embedding path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"object":"list","data":[{"object":"embedding","index":0,"embedding":[0.1,0.2,0.3]}],"model":"test-embedding","usage":{"prompt_tokens":1,"total_tokens":1}}`)
	}))
	t.Cleanup(server.Close)

	cfg := openai.DefaultConfig("test-key")
	cfg.BaseURL = server.URL
	return openai.NewClientWithConfig(cfg)
}
