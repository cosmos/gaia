package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/gaia/v28/tests/e2e/common"
)

// TestLogger provides verbose logging for E2E test execution
type TestLogger struct {
	t              testing.TB
	testStartTime  time.Time
	currentTest    string
	verboseLogging bool
}

// init registers NewTestLogger with the common package so common.TestingSuite
// can construct logger without importing the e2e package
func init() {
	common.NewTestLogger = NewTestLogger
}

// NewTestLogger creates new TestLogger and returns it as common.Logger
func NewTestLogger(t testing.TB, verbose bool) common.Logger {
	return &TestLogger{
		t:              t,
		verboseLogging: verbose,
	}
}

func (tl *TestLogger) StartTest(testName string) {
	tl.currentTest = testName
	tl.testStartTime = time.Now()

	logMsg := "\n========================================"
	logMsg += fmt.Sprintf("\n[START] Running test: %s", testName)
	logMsg += fmt.Sprintf("\n[TIME] Started at: %s", tl.testStartTime.Format("15:04:05"))
	logMsg += "\n========================================"

	tl.t.Log(logMsg)
}

func (tl *TestLogger) PassTest(testName string) {
	duration := time.Since(tl.testStartTime)

	logMsg := "\n========================================"
	logMsg += fmt.Sprintf("\n[✓ PASS] Test passed: %s", testName)
	logMsg += fmt.Sprintf("\n[TIME] Duration: %v", duration)
	logMsg += "\n========================================"

	tl.t.Log(logMsg)
}

func (tl *TestLogger) FailTest(testName string, err error) {
	duration := time.Since(tl.testStartTime)

	logMsg := "\n========================================"
	logMsg += fmt.Sprintf("\n[✗ FAIL] Test failed: %s", testName)
	logMsg += fmt.Sprintf("\n[ERROR] %v", err)
	logMsg += fmt.Sprintf("\n[TIME] Duration: %v", duration)
	logMsg += "\n========================================"

	tl.t.Log(logMsg)
}

func (tl *TestLogger) LogStep(stepName string, details ...interface{}) {
	logMsg := fmt.Sprintf("\n  [STEP] %s", stepName)
	for _, detail := range details {
		logMsg += fmt.Sprintf(" | %v", detail)
	}
	tl.t.Log(logMsg)
}

func (tl *TestLogger) LogInfo(format string, args ...interface{}) {
	tl.t.Log(fmt.Sprintf("\n  [INFO] "+format, args...))
}

func (tl *TestLogger) LogError(format string, args ...interface{}) {
	tl.t.Log(fmt.Sprintf("\n  [ERROR] "+format, args...))
}

func (tl *TestLogger) LogSubTest(subTestName string) {
	tl.t.Log(fmt.Sprintf("\n    → %s", subTestName))
}

func (tl *TestLogger) LogDebug(format string, args ...interface{}) {
	if !tl.verboseLogging {
		return
	}
	tl.t.Log(fmt.Sprintf("\n  [DEBUG] "+format, args...))
}

func (tl *TestLogger) LogWarning(format string, args ...interface{}) {
	tl.t.Log(fmt.Sprintf("\n  [WARN] "+format, args...))
}

func (tl *TestLogger) LogSuccess(format string, args ...interface{}) {
	tl.t.Log(fmt.Sprintf("\n  [✓] "+format, args...))
}

func (tl *TestLogger) LogFailure(format string, args ...interface{}) {
	tl.t.Log(fmt.Sprintf("\n  [✗] "+format, args...))
}

func (tl *TestLogger) SetVerbose(verbose bool) {
	tl.verboseLogging = verbose
}

func (tl *TestLogger) IsVerbose() bool {
	return tl.verboseLogging
}

func (tl *TestLogger) GetCurrentTest() string {
	return tl.currentTest
}

func (tl *TestLogger) GetElapsedTime() time.Duration {
	return time.Since(tl.testStartTime)
}

// LogWithTime logs a message with the current elapsed time prepended.
func (tl *TestLogger) LogWithTime(format string, args ...interface{}) {
	elapsed := time.Since(tl.testStartTime)
	msg := fmt.Sprintf(format, args...)
	tl.t.Log(fmt.Sprintf("\n  [%v] %s", elapsed, msg))
}

func (tl *TestLogger) LogSeparator() {
	tl.t.Log("\n═══════════════════════════════════════════════════════════")
}

func (tl *TestLogger) LogSection(sectionName string) {
	tl.LogSeparator()
	tl.t.Log(fmt.Sprintf("\n  ► %s", sectionName))
	tl.LogSeparator()
}
