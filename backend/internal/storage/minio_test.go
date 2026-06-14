package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCleanPathSegment(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: " Product_123 ", want: "product_123"},
		{input: "hello/world", want: "hello-world"},
		{input: " Ёжик ", want: "product"},
		{input: "---valid---", want: "valid"},
		{input: "", want: "product"},
	}

	for _, test := range tests {
		if got := cleanPathSegment(test.input); got != test.want {
			t.Errorf("cleanPathSegment(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestImageObjectName(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		ext      string
	}{
		{name: "lowercases extension", filename: "Cover.JPEG", ext: ".jpeg"},
		{name: "missing extension", filename: "cover", ext: ".bin"},
		{name: "long extension", filename: "cover.abcdefghijklmnop", ext: ".bin"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			objectName := imageObjectName("products", " Product ID ", test.filename)
			if !strings.HasPrefix(objectName, "products/product-id/") {
				t.Fatalf("unexpected object prefix: %q", objectName)
			}
			if path.Ext(objectName) != test.ext {
				t.Fatalf("expected extension %q, got %q", test.ext, path.Ext(objectName))
			}
		})
	}
}

func TestNewMinIO(t *testing.T) {
	storage, err := NewMinIO("localhost:9000", "access", "secret", "bucket", false)
	if err != nil {
		t.Fatalf("create MinIO client: %v", err)
	}
	if storage.client == nil || storage.bucket != "bucket" {
		t.Fatalf("unexpected storage: %+v", storage)
	}

	if _, err := NewMinIO("://invalid", "access", "secret", "bucket", false); err == nil {
		t.Fatal("expected invalid endpoint error")
	}
}

func TestMinIOStorageLifecycle(t *testing.T) {
	serverState := &s3TestServer{objects: make(map[string]s3TestObject)}
	server := httptest.NewServer(serverState)
	defer server.Close()

	storage, err := NewMinIO(strings.TrimPrefix(server.URL, "http://"), "access", "secret", "bucket", false)
	if err != nil {
		t.Fatalf("create MinIO storage: %v", err)
	}
	ctx := context.Background()
	if err := storage.EnsureBucket(ctx); err != nil {
		t.Fatalf("ensure bucket: %v", err)
	}
	if err := storage.EnsureBucket(ctx); err != nil {
		t.Fatalf("ensure existing bucket: %v", err)
	}

	uploads := []struct {
		name string
		put  func() (string, error)
		body string
	}{
		{
			name: "product image",
			put: func() (string, error) {
				return storage.PutProductImage(
					ctx, "product", "cover.jpg", "image/jpeg", strings.NewReader("product-image"), 13,
				)
			},
			body: "product-image",
		},
		{
			name: "product file",
			put: func() (string, error) {
				return storage.PutProductFile(
					ctx, "product", "scheme.pdf", "application/pdf", strings.NewReader("product-file"), 12,
				)
			},
			body: "product-file",
		},
		{
			name: "testimonial image",
			put: func() (string, error) {
				return storage.PutTestimonialImage(
					ctx, "1", "avatar.png", "image/png", strings.NewReader("testimonial"), 11,
				)
			},
			body: "testimonial",
		},
		{
			name: "blog image",
			put: func() (string, error) {
				return storage.PutBlogImage(
					ctx, "post", "post.webp", "image/webp", strings.NewReader("blog-image"), 10,
				)
			},
			body: "blog-image",
		},
		{
			name: "gallery image",
			put: func() (string, error) {
				return storage.PutGalleryImage(
					ctx, "1", "gallery.gif", "image/gif", strings.NewReader("gallery"), 7,
				)
			},
			body: "gallery",
		},
	}

	for _, upload := range uploads {
		t.Run(upload.name, func(t *testing.T) {
			objectName, err := upload.put()
			if err != nil {
				t.Fatalf("put object: %v", err)
			}
			reader, info, err := storage.Get(ctx, objectName)
			if err != nil {
				t.Fatalf("get object: %v", err)
			}
			defer reader.Close()
			body, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("read object: %v", err)
			}
			if string(body) != upload.body {
				t.Fatalf("unexpected object body: %q", body)
			}
			if info.Size != int64(len(upload.body)) || info.ContentType == "" {
				t.Fatalf("unexpected object info: %+v", info)
			}
		})
	}
}

func TestMinIOStorageErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeS3Error(w, http.StatusInternalServerError, "InternalError", "storage unavailable")
	}))
	defer server.Close()

	storage, err := NewMinIO(strings.TrimPrefix(server.URL, "http://"), "access", "secret", "bucket", false)
	if err != nil {
		t.Fatalf("create MinIO storage: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if err := storage.EnsureBucket(ctx); err == nil {
		t.Fatal("expected ensure bucket error")
	}
	if _, err := storage.PutProductImage(
		ctx, "product", "cover.jpg", "image/jpeg", strings.NewReader("x"), 1,
	); err == nil {
		t.Fatal("expected put object error")
	}
	if _, _, err := storage.Get(ctx, "missing"); err == nil {
		t.Fatal("expected get object error")
	}
}

type s3TestServer struct {
	mu      sync.Mutex
	bucket  bool
	objects map[string]s3TestObject
}

type s3TestObject struct {
	body        []byte
	contentType string
}

func (s *s3TestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleanPath := strings.TrimPrefix(r.URL.Path, "/")
	if cleanPath == "bucket" || cleanPath == "bucket/" {
		switch r.Method {
		case http.MethodGet:
			if _, ok := r.URL.Query()["location"]; !ok {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "application/xml")
			_, _ = io.WriteString(w, "<LocationConstraint>us-east-1</LocationConstraint>")
		case http.MethodHead:
			if !s.bucket {
				writeS3Error(w, http.StatusNotFound, "NoSuchBucket", "bucket not found")
				return
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodPut:
			s.bucket = true
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	const prefix = "bucket/"
	if !strings.HasPrefix(cleanPath, prefix) {
		writeS3Error(w, http.StatusNotFound, "NoSuchKey", "object not found")
		return
	}
	objectName := strings.TrimPrefix(cleanPath, prefix)
	switch r.Method {
	case http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeS3Error(w, http.StatusInternalServerError, "InternalError", err.Error())
			return
		}
		if strings.Contains(r.Header.Get("Content-Encoding"), "aws-chunked") ||
			bytes.Contains(body, []byte(";chunk-signature=")) {
			body, err = decodeAWSChunked(body)
			if err != nil {
				writeS3Error(w, http.StatusBadRequest, "InvalidRequest", err.Error())
				return
			}
		}
		s.objects[objectName] = s3TestObject{body: body, contentType: r.Header.Get("Content-Type")}
		w.Header().Set("ETag", `"test-etag"`)
		w.WriteHeader(http.StatusOK)
	case http.MethodHead:
		object, ok := s.objects[objectName]
		if !ok {
			writeS3Error(w, http.StatusNotFound, "NoSuchKey", "object not found")
			return
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(object.body)))
		w.Header().Set("Content-Type", object.contentType)
		w.Header().Set("ETag", `"test-etag"`)
		w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		w.WriteHeader(http.StatusOK)
	case http.MethodGet:
		object, ok := s.objects[objectName]
		if !ok {
			writeS3Error(w, http.StatusNotFound, "NoSuchKey", "object not found")
			return
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(object.body)))
		w.Header().Set("Content-Type", object.contentType)
		w.Header().Set("ETag", `"test-etag"`)
		w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		_, _ = w.Write(object.body)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func decodeAWSChunked(data []byte) ([]byte, error) {
	var decoded bytes.Buffer
	for len(data) > 0 {
		lineEnd := bytes.Index(data, []byte("\r\n"))
		if lineEnd < 0 {
			return nil, fmt.Errorf("invalid chunk header")
		}
		header := string(data[:lineEnd])
		sizeText := strings.SplitN(header, ";", 2)[0]
		size, err := strconv.ParseInt(sizeText, 16, 64)
		if err != nil {
			return nil, fmt.Errorf("parse chunk size: %w", err)
		}
		data = data[lineEnd+2:]
		if size == 0 {
			return decoded.Bytes(), nil
		}
		if int64(len(data)) < size+2 {
			return nil, fmt.Errorf("truncated chunk")
		}
		_, _ = decoded.Write(data[:size])
		data = data[size+2:]
	}
	return decoded.Bytes(), nil
}

func writeS3Error(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(
		w,
		"<Error><Code>%s</Code><Message>%s</Message><Resource>/bucket</Resource><RequestId>test</RequestId></Error>",
		code,
		message,
	)
}
