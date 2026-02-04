package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/akarashov/urltamer/internal/repository"
)

func BenchmarkCreateShortURL(b *testing.B) {
	repo := repository.NewMemoryRepository()
	defer func() {
		if err := repo.Close(); err != nil {
			b.Fatalf("repo close error: %v", err)
		}
	}()
	svc := NewURLService(repo)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		original := fmt.Sprintf("https://example.com/%d", i)
		_, err := svc.CreateShortURL(ctx, original, 1)
		if err != nil {
			b.Fatalf("CreateShortURL error: %v", err)
		}
	}
}

func BenchmarkGetOriginalURL(b *testing.B) {
	repo := repository.NewMemoryRepository()
	defer func() {
		if err := repo.Close(); err != nil {
			b.Fatalf("repo close error: %v", err)
		}
	}()
	svc := NewURLService(repo)
	ctx := context.Background()

	// benchmark setup - fill repository with some URLs
	const cnt = 10000
	for i := 0; i < cnt; i++ {
		orig := fmt.Sprintf("http://foo.bar/%d", i)
		_, err := svc.CreateShortURL(ctx, orig, 1)
		if err != nil {
			b.Fatalf("fill repo with CreateShortURL error: %v", err)
		}
	}

	tamers, err := svc.GetAllURLs(ctx)
	if err != nil || len(tamers) == 0 {
		b.Fatalf("failed to collect tamers for lookup: %v", err)
	}
	shorts := make([]string, len(tamers))
	for i, t := range tamers {
		shorts[i] = t.ShortURL
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := shorts[i%len(shorts)]
		_, err := svc.GetOriginalURL(ctx, s)
		if err != nil {
			b.Fatalf("GetOriginalURL error: %v", err)
		}
	}
}
