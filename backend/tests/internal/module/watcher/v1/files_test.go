package v1

import (
	watcherv1 "backend/internal/module/watcher/v1"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func createTestFiles(n int) (dir string, cleanup func()) {
	dir = gofakeit.Word()
	err := os.Mkdir(dir, 0755)
	if err != nil {
		panic(err)
	}

	for range n {
		tmpFile := filepath.Join(dir, fmt.Sprintf("%s.pdf", gofakeit.Word()))
		err = os.WriteFile(tmpFile, []byte(gofakeit.Word()), 0644)
		if err != nil {
			panic(err)
		}
	}

	cleanup = func() {
		_ = os.RemoveAll(dir)
	}

	return
}

func testCreateNFilesAndScan(t *testing.T, n int) {
	dir, cleanup := createTestFiles(n)
	defer cleanup()
	_ = cleanup

	w := watcherv1.New(
		watcherv1.WithWatchDir(dir),
		watcherv1.WithWorkers(0), // 避免把worker协程跑起来
	)
	if _, err := w.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	if n != w.Files.Len() {
		t.Errorf("expected %d files scanned, got %d", n, w.Files.Len())
	}
}

func createNFilesWithCrypted(total, encrypted int) (dir string, cleanup func()) {
	dir = gofakeit.Word()
	err := os.Mkdir(dir, 0755)
	if err != nil {
		panic(err)
	}

	for range total {
		basename := gofakeit.Word()
		if encrypted > 0 {
			tmpFile := filepath.Join(dir, fmt.Sprintf("%s.enc", basename))
			_ = os.WriteFile(tmpFile, []byte(gofakeit.Word()), 0644)
			encrypted--
		}

		tmpFile := filepath.Join(dir, fmt.Sprintf("%s.pdf", basename))
		_ = os.WriteFile(tmpFile, []byte(gofakeit.Word()), 0644)
	}

	cleanup = func() {
		_ = os.RemoveAll(dir)
	}

	return
}

func testCreateNFilesWithEncrypted(t *testing.T, total, encrypted int) {
	dir, cleanup := createNFilesWithCrypted(total, encrypted)
	defer cleanup()
	_ = cleanup

	w := watcherv1.New(
		watcherv1.WithWatchDir(dir),
		watcherv1.WithWorkers(0), // 避免把worker协程跑起来
	)
	if _, err := w.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	if total != w.Files.Len() {
		t.Errorf("expected %d files scanned, got %d", total, w.Files.Len())
	}
	if encrypted != w.Files.CountEncrypted() {
		t.Errorf("expected %d encrypted files scanned, got %d", encrypted, w.Files.Len())
	}
}

func TestFilesBasicFunctionality(t *testing.T) {
	t.Run("测试: 创建N个文件并扫描进入缓存", func(t *testing.T) {
		testCreateNFilesAndScan(t, 3)
		testCreateNFilesAndScan(t, 5)
	})

	t.Run("测试: 创建N个文件, 包括M (M <= N)个加密文件, 不能重复计数", func(t *testing.T) {
		testCreateNFilesWithEncrypted(t, 10, 5)
	})
}
