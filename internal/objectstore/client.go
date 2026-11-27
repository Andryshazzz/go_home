package objectstore

import (
	"context"
	"fmt"
	"go_home/pkg/logger"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	Timeout   time.Duration
}

type Client struct {
	mc      *minio.Client
	bucket  string
	timeout time.Duration
	log     *logger.Logger
}

func New(ctx context.Context, cfg Config, log *logger.Logger) (*Client, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	c := &Client{mc: mc, bucket: cfg.Bucket, timeout: cfg.Timeout, log: log}

	if err := c.ensureBucket(ctx); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) ensureBucket(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	exists, err := c.mc.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("check bucket %q (is the storage running at the configured endpoint?): %w", c.bucket, err)
	}
	if exists {
		return nil
	}

	if err := c.mc.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{}); err != nil {
		if minio.ToErrorResponse(err).Code == "BucketAlreadyOwnedByYou" {
			return nil
		}

		return fmt.Errorf("create bucket %q: %w", c.bucket, err)
	}

	c.log.Info("object storage bucket created", zap.String("bucket", c.bucket))

	return nil
}

func (c *Client) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	if size <= 0 {
		return fmt.Errorf("put object %q: size must be known and positive, got %d", key, size)
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if _, err := c.mc.PutObject(ctx, c.bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	}); err != nil {
		return fmt.Errorf("put object %q: %w", key, err)
	}

	return nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if err := c.mc.RemoveObject(ctx, c.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete object %q: %w", key, err)
	}

	return nil
}

func (c *Client) PresignedGetURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	u, err := c.mc.PresignedGetObject(ctx, c.bucket, key, ttl, nil)
	if err != nil {
		return "", fmt.Errorf("presign object %q: %w", key, err)
	}

	return u.String(), nil
}
