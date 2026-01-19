---
name: QA Automation Engineer
description: Quality Assurance and Test Automation skill. Use for writing test strategies, generating test cases, implementing unit/integration tests, and CI/CD validation.
---

# QA Automation Engineer Skill

This skill acts as the gatekeeper of quality. It focuses on breaking the system to ensure its robustness, automating repetitive verification, and ensuring API contracts are honored.

## Core Capabilities

- **Test Strategy**: Defining what to test (Pyramid: Unit > Integration > E2E).
- **Go Testing**: Expert in `testing` package, `testify/assert`, and `mock` generation.
- **API Testing**: Automated validation of HTTP endpoints (Status code, Body, Headers).
- **Load Testing**: Simulating concurrency to verify Key Pool logic.

---

## Testing Workflow

### Step 1: Test Plan Generation
Before writing test code, generate a plan:
1.  **Happy Path**: The ideal usage flow.
2.  **Sad Path**: Invalid input, Missing headers, Unauthorized.
3.  **Edge Cases**: Zero values, Max length strings, Network timeouts.
4.  **State Cases**: Testing behaviors like "What happens when ALL Keys are rate-limited?".

### Step 2: Implementation (Go)

#### A. Unit Tests (`*_test.go`)
- Test *pure functions* first (e.g., Format Converters).
- Use **Table-Driven Tests**:
```go
tests := []struct{
    name string
    input string
    want string
    wantErr bool
}{...}
```

#### B. Integration Tests (API)
- Spin up a test instance of the server (`httptest.NewServer`).
- Send real HTTP requests to the router.
- **SSE Testing**: Verify specific chunks are received in order.

#### C. Mocks
- Use Interfaces for external dependencies (Gemini Client, DB).
- Generate Mocks (using `mockery` or manually) to simulate network failures.

#### D. Benchmark Tests
For performance-critical code (converters, stream processing):
```go
func BenchmarkConvertOpenAIToGemini(b *testing.B) {
    req := createSampleOpenAIRequest()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ConvertToGeminiRequest(req)
    }
}

// Run with: go test -bench=. -benchmem ./...
```

### Step 3: CI/CD Integration
- Ensure `go test ./...` runs on every Push.
- Check Test Coverage (`go test -coverprofile=coverage.out`).

---

## QA Best Practices for MxlnAPI

### 1. Testing the Proxy
- **Mock Gemini API**: Do not hit real Google APIs during tests. Create a local mock server that mimics Gemini responses.
- **Verification**: Ensure the transformed JSON matches the OpenAI spec *exactly*.

### 2. Concurrency Testing
- Use `errgroup` (from `golang.org/x/sync/errgroup`) to fire concurrent requests with proper error handling:
```go
func TestKeyPoolConcurrency(t *testing.T) {
    g, ctx := errgroup.WithContext(context.Background())
    for i := 0; i < 100; i++ {
        g.Go(func() error {
            key, err := pool.GetKey(ctx)
            if err != nil {
                return err
            }
            defer pool.ReleaseKey(key)
            // ... perform request
            return nil
        })
    }
    require.NoError(t, g.Wait())
}
```
- Verify `TokenCount` is accurate under load.

### 3. Android Integration Testing
- Since pure Go cannot test Android WebView, focus on testing the `Bind` interface logic that will be exposed to Android.

---

## Pre-Release Checklist

- [ ] **Unit Tests**: Pass with reasonable coverage (aim for >70% on pure logic, IO code may be lower).
- [ ] **Benchmark Tests**: Run `go test -bench=. -benchmem` and document baseline performance.
- [ ] **Race Detection**: Run `go test -race ./...` to catch concurrency bugs.
- [ ] **Linter**: Run `golangci-lint` and fix all high-priority issues.
- [ ] **Build Check**: Verify binaries compile for all targets (Win, Linux, Android).
- [ ] **Clean Startup**: Ensure app starts/stops without leaving zombie processes.
