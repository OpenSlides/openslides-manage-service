package shared_test

import (
	"bytes"
	"errors"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/shared"
)

func TestCreateFile(t *testing.T) {
	t.Run("running shared.CreateFile() in empty directory", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "openslides-manage-service-")
		if err != nil {
			t.Fatalf("generating temporary directory failed: %v", err)
		}
		defer os.RemoveAll(testDir)
		fileName := "test_file_eeGhu6du_3"
		content := "test_content_kohv2EoT_3"

		shared.CreateFile(testDir, false, fileName, []byte(content))

		testContentFile(t, testDir, fileName, content)
	})

	t.Run("running shared.CreateFile() in non existing directory", func(t *testing.T) {
		testDir := "non_existing_directory"
		fileName := "test_file_eeGhu6du_4"
		content := "test_content_kohv2EoT_4"
		hasErrMsg := "no such file or directory"

		err := shared.CreateFile(testDir, false, fileName, []byte(content))

		if !strings.Contains(err.Error(), hasErrMsg) {
			t.Fatalf("running shared.CreateFile() with invalid directory, got error message %q, expected %q", err.Error(), hasErrMsg)
		}
	})

	t.Run("running shared.CreateFile() on existing file with force true", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "openslides-manage-service-")
		if err != nil {
			t.Fatalf("generating temporary directory failed: %v", err)
		}
		defer os.RemoveAll(testDir)
		fileName := "test_file_eeGhu6du_5"
		content := "test_content_kohv2EoT_5"
		shared.CreateFile(testDir, false, fileName, []byte(content))
		content = "test_content_kohv2EoT_5b"

		shared.CreateFile(testDir, true, fileName, []byte(content))

		testContentFile(t, testDir, fileName, content)
	})

	t.Run("running shared.CreateFile() on existing file with force false", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "openslides-manage-service-")
		if err != nil {
			t.Fatalf("generating temporary directory failed: %v", err)
		}
		defer os.RemoveAll(testDir)
		fileName := "test_file_eeGhu6du_6"
		content := "test_content_kohv2EoT_6"
		shared.CreateFile(testDir, false, fileName, []byte(content))
		content2 := "test_content_kohv2EoT_6b"

		err2 := shared.CreateFile(testDir, false, fileName, []byte(content2))

		if err2 != nil {
			t.Fatalf("running shared.CreateFile() with invalid directory, got error message %q, expected nil error", err2.Error())
		}
		testContentFile(t, testDir, fileName, content)

	})
}

func testContentFile(t testing.TB, dir, name, expected string) {
	t.Helper()

	p := path.Join(dir, name)
	if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
		t.Fatalf("file %q does not exist, expected existance", p)
	}
	content, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("error reading file %q: %v", p, err)
	}

	got := string(content)
	if got != expected {
		t.Fatalf("wrong content of file %q, got %q, expected %q", p, got, expected)
	}
}

func TestReadFromFileOrStdin(t *testing.T) {
	testString := "test string Aequoh2aey9Aiyiechoo"
	f, err := os.CreateTemp("", "somefile-*.txt")
	if err != nil {
		t.Fatalf("creating temporary file: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString(testString)
	if err := f.Close(); err != nil {
		t.Fatalf("closing temporary file: %v", err)
	}

	t.Run("running ReadFromFileOrStdin() with regular file", func(t *testing.T) {
		c, err := shared.ReadFromFileOrStdin(f.Name())
		if err != nil {
			t.Fatalf("error reading file: %v", err)
		}
		if !bytes.Equal(c, []byte(testString)) {
			t.Fatalf("wrong content of file %q, got %q, expected %q", f.Name(), string(c), testString)
		}
	})

	t.Run("running ReadFromFileOrStdin() with not existing file", func(t *testing.T) {
		hasErrMsg := "no such file"
		_, err := shared.ReadFromFileOrStdin("unknown_file")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !strings.Contains(err.Error(), hasErrMsg) {
			t.Fatalf("got error message %q, expected %q", err.Error(), hasErrMsg)
		}
	})

	t.Run("running ReadFromFileOrStdin() with - so reading from stdin", func(t *testing.T) {
		_, err := shared.ReadFromFileOrStdin("-")
		if err != nil {
			t.Fatalf("error reading from stdin: %v", err)
		}
	})
}

func TestInputOrFileOrStdin(t *testing.T) {
	t.Run("running TestInputOrFileOrStdin with both arguments", func(t *testing.T) {
		hasErrMsg := "input or filename must be empty"
		_, err := shared.InputOrFileOrStdin("some input", "some-file")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !strings.Contains(err.Error(), hasErrMsg) {
			t.Fatalf("got error message %q, expected %q", err.Error(), hasErrMsg)
		}
	})

	t.Run("running TestInputOrFileOrStdin with no arguments", func(t *testing.T) {
		hasErrMsg := "input and filename must not both be empty"
		_, err := shared.InputOrFileOrStdin("", "")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !strings.Contains(err.Error(), hasErrMsg) {
			t.Fatalf("got error message %q, expected %q", err.Error(), hasErrMsg)
		}
	})

	t.Run("running TestInputOrFileOrStdin with input", func(t *testing.T) {
		expected := "some-string-icds59ifuisofuw09re"
		got, err := shared.InputOrFileOrStdin(expected, "")
		if err != nil {
			t.Fatalf("error reading input: %v", err)
		}
		if !bytes.Equal(got, []byte(expected)) {
			t.Fatalf("wrong content, got %q, expected %q", string(got), expected)
		}

	})

	t.Run("running TestInputOrFileOrStdin with file", func(t *testing.T) {
		expected := "test string A74832refssdjfAiyiechoo"
		f, err := os.CreateTemp("", "somefile-*.txt")
		if err != nil {
			t.Fatalf("creating temporary file: %v", err)
		}
		defer os.Remove(f.Name())
		f.WriteString(expected)
		if err := f.Close(); err != nil {
			t.Fatalf("closing temporary file: %v", err)
		}

		got, err := shared.InputOrFileOrStdin("", f.Name())
		if err != nil {
			t.Fatalf("error reading file: %v", err)
		}
		if !bytes.Equal(got, []byte(expected)) {
			t.Fatalf("wrong content, got %q, expected %q", string(got), expected)
		}
	})
}
