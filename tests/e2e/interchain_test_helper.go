package e2e

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// InterchainTestHelper provides verbose logging for interchain and IBC tests
type InterchainTestHelper struct {
	t               testing.TB
	logger          *TestLogger
	testMutex       sync.Mutex
	testProgressMap map[string]*TestProgress
	verboseMode     bool
	startTime       time.Time
}

// TestProgress tracks progress information for individual tests
type TestProgress struct {
	Name      string
	Status    string
	StartTime time.Time
	EndTime   time.Time
	Error     error
	StepCount int
	Duration  time.Duration
}

// NewInterchainTestHelper creates a new InterchainTestHelper instance
func NewInterchainTestHelper(t testing.TB, verbose bool) *InterchainTestHelper {
	return &InterchainTestHelper{
		t: t,
		logger: &TestLogger{
			t:              t,
			verboseLogging: verbose,
		},
		testProgressMap: make(map[string]*TestProgress),
		verboseMode:     verbose,
		startTime:       time.Now(),
	}
}

func (ith *InterchainTestHelper) StartInterchainTest(testName string) {
	ith.testMutex.Lock()
	defer ith.testMutex.Unlock()

	ith.testProgressMap[testName] = &TestProgress{
		Name:      testName,
		Status:    "RUNNING",
		StartTime: time.Now(),
	}

	header := "\n╔════════════════════════════════════════════════════════╗"
	header += fmt.Sprintf("\n║  [INTERCHAIN TEST] %s", testName)
	header += fmt.Sprintf("\n║  Started: %s", time.Now().Format("15:04:05"))
	header += "\n╚════════════════════════════════════════════════════════╝"

	ith.t.Log(header)
}

func (ith *InterchainTestHelper) CompleteInterchainTest(testName string, err error) {
	ith.testMutex.Lock()
	progress := ith.testProgressMap[testName]
	progress.EndTime = time.Now()
	progress.Duration = progress.EndTime.Sub(progress.StartTime)
	progress.Error = err
	if err != nil {
		progress.Status = "FAILED"
	} else {
		progress.Status = "PASSED"
	}
	ith.testMutex.Unlock()

	statusIcon := "✓"
	statusText := "PASSED"
	if err != nil {
		statusIcon = "✗"
		statusText = "FAILED"
	}

	footer := "\n╔════════════════════════════════════════════════════════╗"
	footer += fmt.Sprintf("\n║  [%s %s] %s", statusIcon, statusText, testName)
	footer += fmt.Sprintf("\n║  Duration: %v", progress.Duration)
	footer += fmt.Sprintf("\n║  Steps Executed: %d", progress.StepCount)
	if err != nil {
		footer += fmt.Sprintf("\n║  Error: %v", err)
	}
	footer += "\n╚════════════════════════════════════════════════════════╝"

	ith.t.Log(footer)
}

func (ith *InterchainTestHelper) LogChainSetup(chainID string, validators int, ports ...string) {
	msg := fmt.Sprintf("\n[CHAIN SETUP] %s | Validators: %d", chainID, validators)
	for i, port := range ports {
		msg += fmt.Sprintf(" | Port%d: %s", i+1, port)
	}
	ith.t.Log(msg)
}

func (ith *InterchainTestHelper) LogRelayerSetup(relayerName, chain1, chain2 string) {
	ith.t.Log(fmt.Sprintf("\n[RELAYER SETUP] %s | Connecting: %s <-> %s", relayerName, chain1, chain2))
}

func (ith *InterchainTestHelper) LogConnectionEstablished(connID, counterpartyConnID string) {
	ith.t.Log(fmt.Sprintf("\n[CONNECTION ESTABLISHED] Connection: %s | Counterparty: %s", connID, counterpartyConnID))
}

func (ith *InterchainTestHelper) LogChannelCreation(portID, channelID, counterpartyPort, counterpartyChannel string) {
	ith.t.Log(fmt.Sprintf("\n[CHANNEL CREATED] Port: %s | Channel: %s | Counterparty: %s/%s",
		portID, channelID, counterpartyPort, counterpartyChannel))
}

func (ith *InterchainTestHelper) LogPacketFlow(srcChain, dstChain, msgType string, sequence uint64) {
	ith.t.Log(fmt.Sprintf("\n[PACKET FLOW] %s -> %s | Type: %s | Sequence: %d", srcChain, dstChain, msgType, sequence))
}

func (ith *InterchainTestHelper) LogInterchainStep(testName, stepDesc string) {
	ith.testMutex.Lock()
	if progress, exists := ith.testProgressMap[testName]; exists {
		progress.StepCount++
	}
	stepNum := 0
	if progress, exists := ith.testProgressMap[testName]; exists {
		stepNum = progress.StepCount
	}
	ith.testMutex.Unlock()

	ith.t.Log(fmt.Sprintf("\n  [STEP %d] %s", stepNum, stepDesc))
}

func (ith *InterchainTestHelper) LogTransactionExecution(chainID, txHash string, success bool) {
	status := "✓ SUCCESS"
	if !success {
		status = "✗ FAILED"
	}
	ith.t.Log(fmt.Sprintf("\n  [TX %s] Chain: %s | Hash: %s", status, chainID, txHash))
}

func (ith *InterchainTestHelper) LogWaitForRelayer(duration time.Duration) {
	ith.t.Log(fmt.Sprintf("\n  [WAIT] Waiting for relayer... (timeout: %v)", duration))
}

func (ith *InterchainTestHelper) LogRelayerEvent(relayerName, eventType, details string) {
	ith.t.Log(fmt.Sprintf("\n  [RELAYER] %s | %s | %s", relayerName, eventType, details))
}

func (ith *InterchainTestHelper) LogPacketReceived(chainID string, sequence uint64, success bool) {
	status := "✓"
	if !success {
		status = "✗"
	}
	ith.t.Log(fmt.Sprintf("\n  [PACKET RECEIVED %s] Chain: %s | Sequence: %d", status, chainID, sequence))
}

func (ith *InterchainTestHelper) LogAckPacket(srcChain, dstChain string, sequence uint64, success bool) {
	status := "✓"
	if !success {
		status = "✗"
	}
	ith.t.Log(fmt.Sprintf("\n  [ACK %s] %s -> %s | Sequence: %d", status, dstChain, srcChain, sequence))
}

func (ith *InterchainTestHelper) LogTimeoutPacket(srcChain string, sequence uint64) {
	ith.t.Log(fmt.Sprintf("\n  [TIMEOUT PACKET] Chain: %s | Sequence: %d", srcChain, sequence))
}

func (ith *InterchainTestHelper) LogBalance(chainID, address, amount, denom string) {
	ith.t.Log(fmt.Sprintf("\n  [BALANCE] %s on %s | Address: %s | Amount: %s %s",
		denom, chainID, address, amount, denom))
}

func (ith *InterchainTestHelper) LogValidation(description string, passed bool) {
	status := "✓"
	if !passed {
		status = "✗"
	}
	ith.t.Log(fmt.Sprintf("\n  [VALIDATION %s] %s", status, description))
}

func (ith *InterchainTestHelper) LogICACreation(ownerChain, hostChain, icaAddress string) {
	ith.t.Log(fmt.Sprintf("\n[ICA CREATED] Owner: %s | Host: %s | Address: %s",
		ownerChain, hostChain, icaAddress))
}

func (ith *InterchainTestHelper) LogICAProgramming(ownerChain, hostChain string, msgCount int, success bool) {
	status := "✓"
	if !success {
		status = "✗"
	}
	ith.t.Log(fmt.Sprintf("\n[ICA PROGRAM %s] Owner: %s | Host: %s | Messages: %d",
		status, ownerChain, hostChain, msgCount))
}

func (ith *InterchainTestHelper) PrintTestSummary() {
	ith.testMutex.Lock()
	defer ith.testMutex.Unlock()

	var passedCount, failedCount int
	var totalDuration time.Duration

	summary := "\n\n╔════════════════════════════════════════════════════════╗"
	summary += "\n║              TEST EXECUTION SUMMARY                      ║"
	summary += "\n╠════════════════════════════════════════════════════════╣"

	for testName, progress := range ith.testProgressMap {
		if progress.Status == "PASSED" {
			passedCount++
			summary += fmt.Sprintf("\n║ ✓ %-50s ║", testName)
		} else if progress.Status == "FAILED" {
			failedCount++
			summary += fmt.Sprintf("\n║ ✗ %-50s ║", testName)
		}
		totalDuration += progress.Duration
	}

	summary += "\n╠════════════════════════════════════════════════════════╣"
	summary += fmt.Sprintf("\n║ Total Tests: %-42d  ║", len(ith.testProgressMap))
	summary += fmt.Sprintf("\n║ Passed: %-46d  ║", passedCount)
	summary += fmt.Sprintf("\n║ Failed: %-46d  ║", failedCount)
	summary += fmt.Sprintf("\n║ Total Duration: %-38v  ║", totalDuration)
	summary += "\n╚════════════════════════════════════════════════════════╝\n"

	ith.t.Log(summary)
}

func (ith *InterchainTestHelper) GetTestStatus(testName string) *TestProgress {
	ith.testMutex.Lock()
	defer ith.testMutex.Unlock()
	if progress, exists := ith.testProgressMap[testName]; exists {
		return progress
	}
	return nil
}

func (ith *InterchainTestHelper) GetTestCount() int {
	ith.testMutex.Lock()
	defer ith.testMutex.Unlock()
	return len(ith.testProgressMap)
}

func (ith *InterchainTestHelper) GetPassedCount() int {
	ith.testMutex.Lock()
	defer ith.testMutex.Unlock()
	count := 0
	for _, progress := range ith.testProgressMap {
		if progress.Status == "PASSED" {
			count++
		}
	}
	return count
}

func (ith *InterchainTestHelper) GetFailedCount() int {
	ith.testMutex.Lock()
	defer ith.testMutex.Unlock()
	count := 0
	for _, progress := range ith.testProgressMap {
		if progress.Status == "FAILED" {
			count++
		}
	}
	return count
}

func (ith *InterchainTestHelper) GetTotalDuration() time.Duration {
	ith.testMutex.Lock()
	defer ith.testMutex.Unlock()
	var total time.Duration
	for _, progress := range ith.testProgressMap {
		total += progress.Duration
	}
	return total
}

func (ith *InterchainTestHelper) LogDebugInfo(format string, args ...interface{}) {
	if !ith.verboseMode {
		return
	}
	ith.t.Log(fmt.Sprintf("\n  [DEBUG] "+format, args...))
}

func (ith *InterchainTestHelper) LogError(format string, args ...interface{}) {
	ith.t.Log(fmt.Sprintf("\n  [ERROR] "+format, args...))
}

func (ith *InterchainTestHelper) LogWarning(format string, args ...interface{}) {
	ith.t.Log(fmt.Sprintf("\n  [WARN] "+format, args...))
}

func (ith *InterchainTestHelper) SetVerbose(verbose bool) {
	ith.verboseMode = verbose
	if ith.logger != nil {
		ith.logger.SetVerbose(verbose)
	}
}

func (ith *InterchainTestHelper) IsVerbose() bool {
	return ith.verboseMode
}

func (ith *InterchainTestHelper) GetLogger() *TestLogger {
	return ith.logger
}
