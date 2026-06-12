package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"new-forstitch-site/backend/internal/models"
)

type MinIOStorage struct {
	bucket string
	client *minio.Client
}

func NewMinIO(endpoint string, accessKey string, secretKey string, bucket string, useSSL bool) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &MinIOStorage{
		bucket: bucket,
		client: client,
	}, nil
}

func (s *MinIOStorage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
}

func (s *MinIOStorage) PutProductImage(ctx context.Context, productID string, filename string, contentType string, reader io.Reader, size int64) (string, error) {
	objectName := imageObjectName("products", productID, filename)
	_, err := s.client.PutObject(ctx, s.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return objectName, nil
}

func (s *MinIOStorage) PutProductFile(ctx context.Context, productID string, filename string, contentType string, reader io.Reader, size int64) (string, error) {
	objectName := imageObjectName("product-files", productID, filename)
	_, err := s.client.PutObject(ctx, s.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return objectName, nil
}

func (s *MinIOStorage) PutTestimonialImage(ctx context.Context, testimonialID string, filename string, contentType string, reader io.Reader, size int64) (string, error) {
	objectName := imageObjectName("testimonials", testimonialID, filename)
	_, err := s.client.PutObject(ctx, s.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return objectName, nil
}

func (s *MinIOStorage) PutBlogImage(ctx context.Context, postID string, filename string, contentType string, reader io.Reader, size int64) (string, error) {
	objectName := imageObjectName("blog", postID, filename)
	_, err := s.client.PutObject(ctx, s.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return objectName, nil
}

func (s *MinIOStorage) PutGalleryImage(ctx context.Context, itemID string, filename string, contentType string, reader io.Reader, size int64) (string, error) {
	objectName := imageObjectName("gallery", itemID, filename)
	_, err := s.client.PutObject(ctx, s.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return objectName, nil
}

func (s *MinIOStorage) Get(ctx context.Context, objectName string) (io.ReadCloser, models.FileObject, error) {
	info, err := s.client.StatObject(ctx, s.bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, models.FileObject{}, err
	}

	object, err := s.client.GetObject(ctx, s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, models.FileObject{}, err
	}

	return object, models.FileObject{
		ContentType: info.ContentType,
		Size:        info.Size,
	}, nil
}

func imageObjectName(prefix string, ownerID string, filename string) string {
	extension := strings.ToLower(path.Ext(filename))
	if extension == "" || len(extension) > 12 {
		extension = ".bin"
	}
	return fmt.Sprintf("%s/%s/%d%s", prefix, cleanPathSegment(ownerID), time.Now().UnixNano(), extension)
}

func cleanPathSegment(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-' || r == '_':
			builder.WriteRune(r)
		default:
			builder.WriteByte('-')
		}
	}
	out := strings.Trim(builder.String(), "-_")
	if out == "" {
		return "product"
	}
	return out
}
