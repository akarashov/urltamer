package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/akarashov/urltamer/internal/model"
)

func BenchmarkMemory_InsertTamer(b *testing.B) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := model.Tamer{
			ShortURL:    fmt.Sprintf("tmr%d", i),
			OriginalURL: fmt.Sprintf("http://foo.bar/%d", i),
			UserID:      1,
		}
		_, err := repo.InsertTamer(ctx, t)
		if err != nil {
			b.Fatalf("InsertTamer error: %v", err)
		}
	}
}

func BenchmarkMemory_GetTamerByShortURL(b *testing.B) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	const cnt = 10000
	shorts := make([]string, 0, cnt)
	for i := 0; i < cnt; i++ {
		t := model.Tamer{
			ShortURL:    fmt.Sprintf("tmr%d", i),
			OriginalURL: fmt.Sprintf("http://foo.bar/%d", i),
			UserID:      1,
		}
		_, err := repo.InsertTamer(ctx, t)
		if err != nil {
			b.Fatalf("fill repo InsertTamer error: %v", err)
		}
		shorts = append(shorts, t.ShortURL)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := shorts[i%len(shorts)]
		_, err := repo.GetTamerByShortURL(ctx, s)
		if err != nil {
			b.Fatalf("GetTamerByShortURL error: %v", err)
		}
	}
}

func BenchmarkMemory_GetTamerByOriginalURL(b *testing.B) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	const cnt = 10000
	originals := make([]string, 0, cnt)
	for i := 0; i < cnt; i++ {
		t := model.Tamer{
			ShortURL:    fmt.Sprintf("tmr%d", i),
			OriginalURL: fmt.Sprintf("http://foo.bar/%d", i),
			UserID:      1,
		}
		_, err := repo.InsertTamer(ctx, t)
		if err != nil {
			b.Fatalf("fill repo InsertTamer error: %v", err)
		}
		originals = append(originals, t.OriginalURL)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o := originals[i%len(originals)]
		_, err := repo.GetTamerByOriginalURL(ctx, o)
		if err != nil {
			b.Fatalf("GetTamerByOriginalURL error: %v", err)
		}
	}
}

func BenchmarkMemory_DeleteTamer(b *testing.B) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	const cnt = 10000
	shorts := make([]string, 0, cnt)
	for i := 0; i < cnt; i++ {
		t := model.Tamer{
			ShortURL:    fmt.Sprintf("tmr%d", i),
			OriginalURL: fmt.Sprintf("http://foo.bar/%d", i),
			UserID:      1,
		}
		_, err := repo.InsertTamer(ctx, t)
		if err != nil {
			b.Fatalf("fill repo InsertTamer error: %v", err)
		}
		shorts = append(shorts, t.ShortURL)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := shorts[i%len(shorts)]
		_, err := repo.DeleteTamer(ctx, 1, []string{s})
		if err != nil {
			b.Fatalf("DeleteTamer error: %v", err)
		}
	}
}
