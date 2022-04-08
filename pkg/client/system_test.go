package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/OpenSlides/openslides-manage-service/pkg/client"
)

const systemTest = "OPENSLIDES_MANAGE_SERVICE_SYSTEM_TEST"

const configYML = `---
services:
  manage:
    additionalContent:
      build: %s
`

func TestSystemInTotal(t *testing.T) {
	if ok, _ := strconv.ParseBool(os.Getenv(systemTest)); !ok {
		// Error value does not matter here. In case of an error ok is false and
		// this is the expected behavior.
		t.SkipNow()
	}

	dir := setupTestDir(t)

	dockerCompose(context.Background(), t, dir, "pull", os.Stdout)
	dockerCompose(context.Background(), t, dir, "build", os.Stdout)

	down := dockerComposeUp(t, dir)
	defer down()

	time.Sleep(1 * time.Second)
	waitFor(t, dir, "proxy")
	waitFor(t, dir, "backendManage")
	t.Logf("OpenSlides is ready\n")
	time.Sleep(10 * time.Second)

	t.Run("initial-data", func(t *testing.T) {
		cmd := client.RootCmd()
		cmd.SetArgs([]string{"initial-data", "--password-file", path.Join(dir, "secrets", "manage_auth_password")})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("executing command returns error %v", err)
		}
	})

	t.Run("version", func(t *testing.T) {
		cmd := client.RootCmd()
		cmd.SetArgs([]string{"version", "--password-file", path.Join(dir, "secrets", "manage_auth_password")})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("executing command returns error %v", err)
		}
	})
}

func setupTestDir(t testing.TB) string {
	t.Helper()

	t.Logf("Setup temporary directory\n")

	dir := t.TempDir()

	p := path.Join(dir, "config.yml")
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("recovering caller's information failed")
	}
	content := fmt.Sprintf(configYML, path.Join(thisFile, "..", "..", ".."))

	if err := os.WriteFile(p, []byte(content), 0666); err != nil {
		t.Fatalf("creating and writing to file %q: %v", p, err)
	}

	cmd := client.RootCmd()
	cmd.SetArgs([]string{"setup", "--config", p, dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("executing setup command returns error: %v", err)
	}

	return dir
}

func dockerCompose(ctx context.Context, t testing.TB, dir string, args string, writer io.Writer) {
	t.Helper()

	t.Logf("Running docker compose %s\n", args)

	cmdArgs := []string{
		"compose",
		"--file",
		path.Join(dir, "docker-compose.yml"),
	}
	cmdArgs = append(cmdArgs, strings.Split(args, " ")...)

	cmd := exec.CommandContext(ctx, "docker", cmdArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("creating stdout pipe %v", err)
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("creating stderr pipe %v", err)
	}
	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		t.Fatalf("starting command docker compose %s: %v", args, err)
	}

	if _, err := io.Copy(writer, io.MultiReader(stderr, stdout)); err != nil {
		t.Fatalf("piping data: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		t.Logf("Error while waiting for the command: %v", err)
	}
}

func dockerComposeUp(t testing.TB, dir string) func() {
	t.Helper()

	t.Logf("Running docker compose up\n")

	upCmd := exec.Command("docker", "compose", "--file", path.Join(dir, "docker-compose.yml"), "up")
	if err := upCmd.Start(); err != nil {
		t.Fatalf("starting docker compose up command: %v", err)
	}

	down := func() {
		t.Log("Running docker compose down\n")
		if err := exec.Command("docker", "compose", "--file", path.Join(dir, "docker-compose.yml"), "down").Run(); err != nil {
			t.Logf("Error while running docker compose down: %v", err)
		}
	}

	return down
}

func waitFor(t testing.TB, dir string, service string) {
	t.Helper()

	t.Logf("Checking state of %s\n", service)

	for {
		var buf bytes.Buffer
		ctx := context.Background()
		args := fmt.Sprintf("ps %s --format json", service)
		dockerCompose(ctx, t, dir, args, &buf)
		b, err := io.ReadAll(&buf)
		if err != nil {
			t.Fatalf("reading output of docker compose ps: %v", err)
		}
		if isRunning(t, b) {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func isRunning(t testing.TB, b []byte) bool {
	t.Helper()

	if bytes.HasPrefix(b, []byte("No such service")) {
		return false
	}

	var out []struct {
		State string `json:"State"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshalling JSON %s: %v", b, err)
	}

	if len(out) == 0 {
		return false
	}
	if out[0].State != "running" {
		return false
	}

	return true
}
