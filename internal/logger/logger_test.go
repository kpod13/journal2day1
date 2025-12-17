package logger

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log := New(&buf)

	require.NotNil(t, log)
	require.NotNil(t, log.out)
	require.NotNil(t, log.info)
	require.NotNil(t, log.success)
	require.NotNil(t, log.warn)
	require.NotNil(t, log.err)
	require.NotNil(t, log.bold)
	require.NotNil(t, log.dim)
}

func TestInfo(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log := New(&buf)
	log.Info("test %s", "message")

	output := buf.String()

	require.Contains(t, output, "test message")
	require.Contains(t, output, "ℹ")
}

func TestSuccess(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log := New(&buf)
	log.Success("done %d", 42)

	output := buf.String()

	require.Contains(t, output, "done 42")
	require.Contains(t, output, "✓")
}

func TestWarn(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log := New(&buf)
	log.Warn("warning %s", "here")

	output := buf.String()

	require.Contains(t, output, "warning here")
	require.Contains(t, output, "⚠")
}

func TestError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log := New(&buf)
	log.Error("error %s", "occurred")

	output := buf.String()

	require.Contains(t, output, "error occurred")
	require.Contains(t, output, "✗")
}

func TestStep(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log := New(&buf)
	log.Step("processing %s", "data")

	output := buf.String()

	require.Contains(t, output, "processing data")
	require.Contains(t, output, "→")
}

func TestBold(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log := New(&buf)
	log.Bold("bold %s", "text")

	output := buf.String()

	require.Contains(t, output, "bold text")
}

func TestDim(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log := New(&buf)
	log.Dim("dim %s", "text")

	output := buf.String()

	require.Contains(t, output, "dim text")
}

func TestPrint(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log := New(&buf)
	log.Print("plain %s", "text")

	output := buf.String()

	require.Equal(t, "plain text", output)
}

func TestPrintln(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log := New(&buf)
	log.Println("line %d", 1)

	output := buf.String()

	require.Equal(t, "line 1\n", output)
}

func TestHeader(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log := New(&buf)
	log.Header("My Header")

	output := buf.String()

	require.Contains(t, output, "My Header")
	require.Contains(t, output, "─")
}

func TestKeyValue(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log := New(&buf)
	log.KeyValue("Key", "Value")

	output := buf.String()

	require.Contains(t, output, "Key")
	require.Contains(t, output, "Value")
}
