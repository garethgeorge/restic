package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/restic/restic/internal/feature"
	rtest "github.com/restic/restic/internal/test"
	"golang.org/x/sync/errgroup"
)

func testRunLsWithOpts(t testing.TB, gopts GlobalOptions, opts LsOptions, args []string) []byte {
	buf, err := withCaptureStdout(func() error {
		gopts.Quiet = true
		return runLs(context.TODO(), opts, gopts, args)
	})
	rtest.OK(t, err)
	return buf.Bytes()
}

func testRunLs(t testing.TB, gopts GlobalOptions, snapshotID string) []string {
	out := testRunLsWithOpts(t, gopts, LsOptions{}, []string{snapshotID})
	return strings.Split(string(out), "\n")
}

func assertIsValidJSON(t *testing.T, data []byte) {
	// Sanity check: output must be valid JSON.
	var v []any
	err := json.Unmarshal(data, &v)
	rtest.OK(t, err)
	rtest.Assert(t, len(v) == 4, "invalid ncdu output, expected 4 array elements, got %v", len(v))
}

func TestRunLsNcdu(t *testing.T) {
	env, cleanup := withTestEnvironment(t)
	defer cleanup()

	testSetupBackupData(t, env)
	opts := BackupOptions{}
	// backup such that there are multiple toplevel elements
	testRunBackup(t, env.testdata+"/0", []string{"."}, opts, env.gopts)

	for _, paths := range [][]string{
		{"latest"},
		{"latest", "/0"},
		{"latest", "/0", "/0/9"},
	} {
		ncdu := testRunLsWithOpts(t, env.gopts, LsOptions{Ncdu: true}, paths)
		assertIsValidJSON(t, ncdu)
	}
}

func TestRunLsScriptMode(t *testing.T) {
	feature.Flag.Apply("script-mode-for-ls", func(warn string) {
		t.Fatal(warn)
	})

	env, cleanup := withTestEnvironment(t)
	defer cleanup()

	testSetupBackupData(t, env)
	opts := BackupOptions{}
	testRunBackup(t, env.testdata+"/0", []string{"."}, opts, env.gopts)

	// Test that script mode can be enabled (basic functionality test)
	t.Run("script mode not enabled", func(t *testing.T) {
		err := runLs(context.TODO(), LsOptions{ScriptMode: true}, env.gopts, []string{})
		rtest.Assert(t, err != nil && strings.Contains(err.Error(), "script mode is not enabled"),
			fmt.Sprintf("script mode should be enabled with feature flag, got %v", err))
	})

	// Test that script mode requires the feature flag
	t.Run("script mode can list directory", func(t *testing.T) {
		feature.Flag.Apply("script-mode-for-ls", func(warn string) {
			t.Fatal(warn)
		})

		withRestoreGlobalOptions(func() error {
			stdoutBuf := bytes.NewBuffer(nil)
			globalOptions.stdout = stdoutBuf
			globalOptions.stdin = bytes.NewBuffer([]byte("{\"command\": \"ls\", \"path\": \"/0\", \"snapshot\": \"latest\"}"))

			errgroup := errgroup.Group{}
			errgroup.Go(func() error {
				return runLs(context.TODO(), LsOptions{ScriptMode: true}, env.gopts, []string{})
			})
			err := errgroup.Wait()
			rtest.OK(t, err)

			// Verify JSON output structure matches expected format
			output := stdoutBuf.String()
			lines := strings.Split(strings.TrimSpace(output), "\n")
			rtest.Assert(t, len(lines) == 4, "expected 4 lines of output, got %d", len(lines))

			// Parse each line and verify key fields
			var msgs []map[string]interface{}
			for _, line := range lines {
				var msg map[string]interface{}
				err = json.Unmarshal([]byte(line), &msg)
				rtest.OK(t, err)
				msgs = append(msgs, msg)
			}

			// Verify message sequence: ready -> node(0) -> node(9) -> end
			rtest.Assert(t, msgs[0]["type"] == "ready" && msgs[0]["message"] == "repository initialized", "invalid ready message")
			rtest.Assert(t, msgs[1]["name"] == "0" && msgs[1]["type"] == "dir" && msgs[1]["path"] == "/0", "invalid dir entry for '0'")
			rtest.Assert(t, msgs[2]["name"] == "9" && msgs[2]["type"] == "dir" && msgs[2]["path"] == "/0/9", "invalid dir entry for '9'")
			rtest.Assert(t, msgs[3]["type"] == "end", "invalid end message")

			return nil
		})
	})
}
