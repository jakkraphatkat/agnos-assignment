package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Record struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Store struct {
	client       *minio.Client
	bucket       string
	recordPrefix string
}

func openStore(cfg config) (*Store, error) {
	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
		Region: cfg.MinIORegion,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	store := &Store{
		client:       client,
		bucket:       cfg.MinIOBucket,
		recordPrefix: normalizePrefix(cfg.MinIORecordPrefix),
	}

	if err := store.ensureBucket(context.Background(), cfg.MinIORegion); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) Close() error {
	return nil
}

func (s *Store) TouchTodayRecords(now time.Time) (int, error) {
	ctx := context.Background()
	now = now.UTC()
	updated := 0

	for object := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    s.objectPrefixForDate(now),
		Recursive: true,
	}) {
		if object.Err != nil {
			return 0, fmt.Errorf("list objects: %w", object.Err)
		}

		record, err := s.getRecord(ctx, object.Key)
		if err != nil {
			return 0, err
		}

		record.UpdatedAt = now
		if err := s.putRecord(ctx, object.Key, record); err != nil {
			return 0, err
		}

		updated++
	}

	return updated, nil
}

func (s *Store) ensureBucket(ctx context.Context, region string) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket %q: %w", s.bucket, err)
	}

	if exists {
		return nil
	}

	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{Region: region}); err != nil {
		return fmt.Errorf("create bucket %q: %w", s.bucket, err)
	}

	return nil
}

func (s *Store) getRecord(ctx context.Context, objectName string) (Record, error) {
	reader, err := s.client.GetObject(ctx, s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return Record{}, fmt.Errorf("get object %q: %w", objectName, err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			logJSON("error", "close object reader failed", map[string]any{
				"bucket": s.bucket,
				"object": objectName,
				"error":  closeErr.Error(),
			})
		}
	}()

	payload, err := io.ReadAll(reader)
	if err != nil {
		return Record{}, fmt.Errorf("read object %q: %w", objectName, err)
	}

	var record Record
	if err := json.Unmarshal(payload, &record); err != nil {
		return Record{}, fmt.Errorf("decode object %q: %w", objectName, err)
	}

	return record, nil
}

func (s *Store) putRecord(ctx context.Context, objectName string, record Record) error {
	payload, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal record: %w", err)
	}

	_, err = s.client.PutObject(
		ctx,
		s.bucket,
		objectName,
		bytes.NewReader(payload),
		int64(len(payload)),
		minio.PutObjectOptions{ContentType: "text/plain; charset=utf-8"},
	)
	if err != nil {
		return fmt.Errorf("put object %q: %w", objectName, err)
	}

	return nil
}

func (s *Store) objectPrefixForDate(date time.Time) string {
	date = date.UTC()
	return fmt.Sprintf(
		"%s%04d/%02d/%02d/",
		s.recordPrefix,
		date.Year(),
		date.Month(),
		date.Day(),
	)
}

func normalizePrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	prefix = strings.Trim(prefix, "/")
	if prefix == "" {
		return "records/"
	}

	return prefix + "/"
}
