package backups

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/backend/evidenceguard"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ObjectReader interface {
	Read(context.Context, string, int64) ([]byte, error)
}
type Dumper interface {
	Dump(context.Context, string, string) error
}
type PGDump struct{ Pool *pgxpool.Pool }

func (d PGDump) Dump(ctx context.Context, snapshot, path string) error {
	c := d.Pool.Config().ConnConfig
	// The application's database transport is the private Compose network.
	if c.TLSConfig != nil {
		return ErrConfig
	}
	cmd := exec.CommandContext(ctx, "pg_dump", "--format=custom", "--no-owner", "--no-acl", "--snapshot="+snapshot, "--file="+path)
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "PGHOST=" + c.Host, "PGPORT=" + strconv.Itoa(int(c.Port)), "PGUSER=" + c.User, "PGDATABASE=" + c.Database, "PGPASSWORD=" + c.Password, "PGSSLMODE=disable", "PGCONNECT_TIMEOUT=10"}
	if cmd.Run() != nil {
		return ErrBackup
	}
	return nil
}

type Object struct {
	Hash string `json:"sha256"`
	Key  string `json:"key"`
	Size int64  `json:"bytes"`
}
type Manifest struct {
	Format         int       `json:"format"`
	ID             string    `json:"id"`
	SnapshotAt     time.Time `json:"snapshot_at"`
	DatabaseSHA256 string    `json:"database_sha256"`
	DatabaseBytes  int64     `json:"database_bytes"`
	Objects        []Object  `json:"objects"`
}
type Bundle struct {
	Path, SHA256 string
	Size         int64
	Manifest     Manifest
}

// Build exports one database snapshot for pg_dump and the object inventory.
// Immutable object names allow collection to continue. Cleanup holds the
// exclusive evidence guard through both database removal and S3 deletion.
func Build(ctx context.Context, pool *pgxpool.Pool, objects ObjectReader, dump Dumper, scratch, id string) (Bundle, func(), error) {
	var b Bundle
	dir, err := os.MkdirTemp(scratch, "iwa-backup-")
	if err != nil {
		return b, func() {}, ErrBackup
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	fail := func() (Bundle, func(), error) { cleanup(); return Bundle{}, func() {}, ErrBackup }
	unlock, err := evidenceguard.Lock(ctx, pool, true)
	if err != nil {
		return fail()
	}
	defer unlock()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return fail()
	}
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(rollbackCtx)
	}()
	var snapshot string
	b.Manifest = Manifest{Format: 1, ID: id, Objects: []Object{}}
	if tx.QueryRow(ctx, `SELECT pg_export_snapshot(),clock_timestamp()`).Scan(&snapshot, &b.Manifest.SnapshotAt) != nil {
		return fail()
	}
	rows, err := tx.Query(ctx, `SELECT hash,object_key,byte_size FROM retained_objects ORDER BY hash`)
	if err != nil {
		return fail()
	}
	for rows.Next() {
		var o Object
		if rows.Scan(&o.Hash, &o.Key, &o.Size) != nil {
			rows.Close()
			return fail()
		}
		b.Manifest.Objects = append(b.Manifest.Objects, o)
	}
	rows.Close()
	if rows.Err() != nil {
		return fail()
	}
	dumpPath := filepath.Join(dir, "postgres.dump")
	if dump.Dump(ctx, snapshot, dumpPath) != nil {
		return fail()
	}
	f, err := os.Open(dumpPath)
	if err != nil {
		return fail()
	}
	defer f.Close()
	h := sha256.New()
	size, err := io.Copy(h, f)
	if err != nil || size == 0 {
		return fail()
	}
	b.Manifest.DatabaseSHA256 = hex.EncodeToString(h.Sum(nil))
	b.Manifest.DatabaseBytes = size
	if _, err = f.Seek(0, 0); err != nil {
		return fail()
	}
	b.Path = filepath.Join(dir, "bundle.tar")
	out, err := os.OpenFile(b.Path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fail()
	}
	h = sha256.New()
	tw := tar.NewWriter(io.MultiWriter(out, h))
	add := func(name string, size int64, r io.Reader) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0600, Size: size, ModTime: b.Manifest.SnapshotAt}); err != nil {
			return err
		}
		_, err := io.CopyN(tw, r, size)
		return err
	}
	writeErr := add("postgres.dump", size, f)
	for _, o := range b.Manifest.Objects {
		if writeErr != nil {
			break
		}
		// Validate the inventory as well as the bytes even for alternate readers.
		if len(o.Hash) != 64 || o.Key != "sha256/"+o.Hash[:2]+"/"+o.Hash {
			writeErr = ErrBackup
			break
		}
		raw, e := objects.Read(ctx, o.Hash, o.Size)
		sum := sha256.Sum256(raw)
		if e != nil || int64(len(raw)) != o.Size || hex.EncodeToString(sum[:]) != o.Hash {
			writeErr = ErrBackup
			break
		}
		writeErr = add("objects/"+o.Key, int64(len(raw)), bytes.NewReader(raw))
	}
	if writeErr == nil {
		raw, e := json.Marshal(b.Manifest)
		if e != nil {
			writeErr = e
		} else {
			writeErr = add("manifest.json", int64(len(raw)), bytes.NewReader(raw))
		}
	}
	closeErr := tw.Close()
	syncErr := out.Sync()
	info, statErr := out.Stat()
	fileErr := out.Close()
	if writeErr != nil || closeErr != nil || syncErr != nil || statErr != nil || fileErr != nil || ctx.Err() != nil {
		return fail()
	}
	b.Size = info.Size()
	b.SHA256 = hex.EncodeToString(h.Sum(nil))
	if tx.Commit(ctx) != nil {
		return fail()
	}
	return b, cleanup, nil
}
