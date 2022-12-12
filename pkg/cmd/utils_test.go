package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// promptResponse is a prompt-response pair representing an expected prompt from stdout and
// the response to write on stdin when such a prompt is encountered
type promptResponse struct {
	prompt   string // optional, because
	response string // A new line character is added to the end of the response,
	// unless if the last character of the response is already a new line.
	// This makes it so that an empty response is equivalent to sending a newline to stdin.
}

type promptResponseSlice []promptResponse

func (promptResponses promptResponseSlice) respondToPrompts(
	stdinWriter io.Writer,
	stdoutReader io.Reader,
	doneSignal chan<- error,
	endSignal <-chan struct{},
) {
	stdOutBufferedReader := bufio.NewReader(stdoutReader)

	var goRoutineErr error
	promptResponsesIndex := 0

	timer := time.NewTimer(maxWaitTimeForPrompt)
	defer timer.Stop()

	ticker := time.NewTicker(promptCheckInterval)
	defer ticker.Stop()

L:
	for promptResponsesIndex < len(promptResponses) {
		pr := promptResponses[promptResponsesIndex]
		// Empty prompt means immediately write to stdin without waiting for a stdout prompt
		// This is useful when a command is reading from stdin such as policy loading.
		if pr.prompt == "" {
			promptResponsesIndex++

			stdinWriter.Write([]byte(maybeAddNewLineSuffix(pr.response)))
			continue
		}

		select {
		// This case makes sure that you don't wait forever for a prompt that never shows up.
		// Here we timebox the wait on a prompt.
		// Additionally ensures we don't leak this goroutine because it ensures that after
		// a second of waiting we break the circuit.
		case <-timer.C:
			goRoutineErr = fmt.Errorf(
				"maxWaitTimeForPrompt was exceeded without the command issuing the expected prompt: %q",
				pr.prompt,
			)
			break L
		// This case avoids wait for the timebox above. When we know the command is finished we can simply
		// break the loop knowing that there will be no further reading.
		case <-endSignal:
			break L
		case <-ticker.C:
			line, err := stdOutBufferedReader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// Sleep is important to avoid a busy goroutine
					// that never yields the CPU to anything else
					continue
				}

				goRoutineErr = err
				break L
			}

			// Strip ANSI because it's just noise
			line = stripAnsi(line)

			// Try to match promptResponses in order
			if strings.Contains(line, pr.prompt) {
				promptResponsesIndex++

				stdinWriter.Write([]byte(maybeAddNewLineSuffix(pr.response)))
			}
		}
	}

	if goRoutineErr == nil && promptResponsesIndex < len(promptResponses) {
		goRoutineErr = fmt.Errorf(
			"promptResponses not exhausted, remaining: %+v",
			promptResponses[promptResponsesIndex:],
		)
	}
	doneSignal <- goRoutineErr
}

// executeCommandForTest executes a cobra command in-memory and returns stdout, stderr and error
func executeCommandForTest(t *testing.T, c *cobra.Command, args ...string) (string, string, error) {
	return executeCommandForTestWithPromptResponses(t, c, nil, args...)
}

// executeCommandForTestWithPromptResponses executes a cobra command in-memory and returns stdout, stderr and error.
// It takes a as input a slice representing a sequence of prompt-responses, each containing an expected
// prompt from stdout and the response to write on stdin when such a prompt is encountered.
func executeCommandForTestWithPromptResponses(
	t *testing.T, c *cobra.Command, promptResponses []promptResponse, args ...string,
) (string, string, error) {
	t.Helper()

	cmd := newRootCommand()
	cmd.AddCommand(c)
	cmd.SetArgs(args)

	c.Short = "HELP SHORT"
	c.Long = "HELP LONG"

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	// Will have to make this recursive if we have deeper subcommands.
	for _, subCmd := range c.Commands() {
		subCmd.Short = "HELP SHORT"
		subCmd.Long = "HELP LONG"
	}

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)

	cmdErr, stdinErr := executeCommandWithPromptResponses(
		cmd,
		promptResponses,
	)

	if stdinErr != nil {
		t.Error(stdinErr)
	}

	// strip ansi from stdout and stderr because we're using promptui
	return stripAnsi(stdoutBuf.String()), stripAnsi(stderrBuf.String()), cmdErr
}

func maybeAddNewLineSuffix(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s
	}

	return s + "\n"
}

const maxWaitTimeForPrompt = 2 * time.Second
const promptCheckInterval = 10 * time.Millisecond

func executeCommandWithPromptResponses(
	cmd *cobra.Command, promptResponses []promptResponse,
) (cmdErr error, stdinErr error) {
	// For no prompt-responses we simply execute with an empty stdin :)
	if len(promptResponses) == 0 {
		cmd.SetIn(bytes.NewReader(nil))
		return cmd.Execute(), nil
	}

	// For prompt-responses we create a pipe for stdin and use goroutine to carry out the
	// required writes to stdin in response to prompts
	stdinReader, stdinWriter := io.Pipe()
	defer func() {
		stdinReader.Close()
		stdinWriter.Close()
	}()
	cmd.SetIn(stdinReader)

	endSignal := make(chan struct{}, 1)
	doneSignal := make(chan error, 1)

	// Create a multiwriter to so that we have a seperate buffer for checking for prompts on
	// stdout without interferring with the original stdout buffer of the command.
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(io.MultiWriter(cmd.OutOrStdout(), stdoutBuf))

	// This goroutine schedules writes to the write-end of the stdin pipe based on the prompt-responses
	go func() {
		// Signal that there will be no further writing
		defer stdinWriter.Close()

		promptResponseSlice(promptResponses).respondToPrompts(
			stdinWriter,
			stdoutBuf,
			doneSignal,
			endSignal,
		)
	}()

	err := cmd.Execute()
	// Signal that there will be no further reading
	stdinReader.Close()

	// Inform the goroutine that the command is done
	endSignal <- struct{}{}

	// Wait for goroutine to finish
	if goroutineErr := <-doneSignal; goroutineErr != nil {
		return err, goroutineErr
	}

	return err, nil
}

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

// stripAnsi is inspired by https://github.com/acarl005/stripansi
func stripAnsi(str string) string {
	return regexp.MustCompile(ansi).ReplaceAllString(str, "")
}
