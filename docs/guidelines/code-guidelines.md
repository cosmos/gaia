# Contributing

If you want to contribute to a project and improve it, your help is welcome. We want to make Gaia as good as it can be. Contributing is also a great way to learn more about blockchain technology and improve it. Please read this document and follow our guidelines to make the process as smooth as possible. We are happy to review your code but please ensure that you have a reasonable and clean pull request.

This documents idiomatic conventions in the Go code that we follow at Uber. A lot of these are general guidelines for Go, while others extend upon external resources:

1. [Effective Go](https://golang.org/doc/effective_go.html)
2. [Go Common Mistakes](https://github.com/golang/go/wiki/CommonMistakes)
3. [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

## Run tests

- Run unit tests

```shell
make test-unit
```

- Run the unit tests and output the coverage file (coverage.txt).

```shell
make test-unit-cover
```

- Run the unit tests with the race condition flag on.

```shell
make test-race
```

- Run end-to-end integration tests (Docker needed).

```shell
make docker-build-hermes && \
make docker-build-debug && \
make test-e2e
```

# Guidelines

These guidelines are the conventions that govern our code. These conventions cover far more than just source file formatting. Can `gofmt` and `goimports` handle that for us.

The goal of this guide is to manage this complexity by describing in detail the Dos and Don'ts of writing Go code. These rules keep the code base manageable while allowing engineers to use Go language features productively.

Try to avoid extensive methods and always test your code. All PRs should have at least 95% of code coverage.

- [Project organization](#project-organization)
- [How to test this project locally](#how-to-test-this-project-locally)
    - [Unit Tests](#unit-tests)
    - [End-to-End Tests](#end-to-end-tests)
    - [Upgrade Test](#upgrade-test)
- [Guidelines](#guidelines)
    - [Line Length](#line-length)
    - [Doc Comments](#doc-comments)
    - [Declaring Empty Slices](#declaring-smpty-slices)
    - [Indent Error Flow](#indent-error-flow)
    - [Unnecessary Else](#unnecessary-else)
    - [Named Result Parameters](#named-result-parameters)
    - [Package Comments](#package-comments)
    - [Package Names](#package-names)
    - [Function Names](#function-names)
    - [Pointers](#pointers)
    - [Receiver Names](#receiver-names)
    - [Variable Names](#variable-names)
    - [Zero-value Mutexes](#zero-value-mutexes)
    - [Copy Slices and Maps at Boundaries](#copy-slices-and-maps-at-boundaries)
    - [Receiving Slices and Maps](#receiving-slices-and-maps)
    - [Returning Slices and Maps](#returning-slices-and-maps)
    - [Errors](#errors)
        - [Error Types](#error-types)
        - [Error Wrapping](#error-wrapping)
        - [Error Naming](#error-naming)
    - [Handle Type Assertion Failures](#handle-type-assertion-failures)
    - [Avoid Embedding Types in Public Structs](#avoid-embedding-types-in-public-structs)
    - [Avoid `init()`](#avoid-init)
    - [Performance](#Performance)
        - [Prefer strconv over fmt](#prefer-strconv-over-fmt)
        - [Avoid string-to-byte conversion](#avoid-string-to-byte-conversion)
        - [Prefer Specifying Container Capacity](#prefer-specifying-container-capacity)
            - [Specifying Map Capacity Hints](#specifying-map-capacity-hints)
            - [Specifying Slice Capacity](#specifying-slice-capacity)
    - [Function Grouping and Ordering](#function-grouping-and-ordering)
    - [Reduce Nesting](#reduce-nesting)
    - [Writing Tests](#writing-tests)
      - [Use Subtests](#use-subtests)
    - [Avoid writing directly in the stdout](#avoid-writing-directly-in-the-stdout)
    - [Avoid panic](#avoid-panic)
    - [Handle error on the top level](#handle-error-on-the-top-level)

## Project organization

- /ante: Where the ante-handler logic is defined.

- /app: Where the application is defined.

- /client: OpenAPI/Swagger specs, JSON schema files, protocol definition files.
    - /swagger-ui

- /cmd/gaiad: Main applications for this project.
    - cmd/
    - main.go

- /contrib (scripts): Scripts to perform various build, install, analysis, etc operations.
    - /devtools
    - /generate_release_note
    - /githooks
    - /scripts
    - /testnets

- /docs: Gaia docs.

- /pkg: Library code that's to be reusable.
    - /address
    - /genesis

- /proto: Proto type definitions

- /tests/e2e: Additional external test apps and test data.

- /third_party/proto: External proto type definitions

- /tools: Supporting tools for this project.

- /x: Cosmos Modules.

## How to test this project locally

### Unit Tests

Running unit tests locally should ensure that the tests inside of `/tests/e2e` are not run. These tests require active running docker containers.

```sh
make test-unit
```

### End-to-End Tests

To run the E2E tests you need to have an instance of Docker running. Then make sure you have the most recent version of the code built in the containers by running:

```sh
make docker-build-debug
```

Then run the tests:

```sh
make test-e2e
```

### Upgrade Test

Instructions for running the upgrade test locally

#### Build current version and move into ./build:
```sh
git checkout v8.0.0
make build 
mv ./build/gaiad ./build/gaiad8
```

#### Build gaia v9.0.0 and move into ./build:
```sh
git checkout v9.0.0
make build 
mv ./build/gaiad ./build/gaiad9
```

#### Go back to your previous working branch
```sh
git checkout -
```

#### Install cosmovisor
```sh
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v1.3.0
```

#### Run the Chain

This script prepares the chain and starts it using cosmovisor
```sh
./contrib/scripts/run-gaia-v8.sh
```

#### Run the upgrade
In another terminal window, run the script that waits 10 seconds for gaia to start then makes gov proposal to perform an upgrade at height 15
```sh
./contrib/scripts/run-upgrade-commands.sh 30
```

#### Monitor for success
In a third window run the upgrade monitoring script that will exit without error when the upgrade succeeds.
```sh
./contrib/scripts/test_upgrade.sh 20 5 16 localhost
```

This should show logs that demonstrate a successful upgrade by reaching block height 16.

## Guidelines

### Line Length

Avoid uncomfortably long lines. Similarly, don't add line breaks to keep lines short when they are more readable long--for example if they are repetitive. The maximum line length is 120. If your line is over 120 characters, break it;

### Doc Comments

All top-level, exported names should have doc comments, as should non-trivial unexported type or function declarations. See https://go.dev/doc/effective_go#commentary for more information about commentary conventions.

### Declaring Empty Slices

When declaring an empty slice, prefer

```go
var t []string
```

over

```go
t := []string{}
```

The former declares a nil slice value, while the latter is non-nil but zero-length. They are functionally equivalent—their `len` and `cap` are both zero—but the nil slice is the preferred style.

Note that there are limited circumstances where a non-nil but the zero-length slice is preferred, such as when encoding JSON objects (a `nil` slice encodes to `null`, while `[]string{}` encodes to the JSON array `[]`).

When designing interfaces, avoid distinguishing between a nil slice and a non-nil, zero-length slice, as this can lead to subtle programming errors. It's also important to distinguish if a map key exists from whether its value is `zero`/`nil`/`false`.

For more discussion about nil in Go see Francesc Campoy's talk [Understanding Nil](https://www.youtube.com/watch?v=ynoY2xz-F8s).


### Indent Error Flow

Try to keep the normal code path at a minimal indentation and indent the error handling, dealing with it first. This improves the readability of the code by permitting visual scanning of the normal path quickly. For instance, don't write:

```go
if err != nil {
	// error handling
} else {
	// normal code
}
```

Instead, write:

```go
if err != nil {
	// error handling
	return // or continue, etc.
}
// normal code
```

### Unnecessary Else

If a variable is set in both branches of an if, it can be replaced with a single if.

- Bad

```go
var a int
if b {
  a = 100
} else {
  a = 10
}
```

- Good

```go
a := 10
if b {
  a = 100
}
```

### Named Result Parameters

Consider what it will look like in godoc. Named result parameters like:

```go
func (n *Node) Parent1() (node *Node) {}
func (n *Node) Parent2() (node *Node, err error) {}
```

It will be repetitive in godoc; better to use:

```go
func (n *Node) Parent1() *Node {}
func (n *Node) Parent2() (*Node, error) {}
```

On the other hand, adding names may be helpful in some contexts if a function returns two or three parameters of the same type or if the meaning of a result isn't clear from the context. Don't name result parameters just to avoid declaring a var inside the function; that trades off minor implementation brevity at the cost of unnecessary API verbosity.

```go
func (f *Foo) Location() (float64, float64, error)
```

It is less clear than the:

```go
// Location returns f's latitude and longitude.
// Negative values mean south and west, respectively.
func (f *Foo) Location() (lat, long float64, err error)
```

Naked returns are okay if the function is a handful of lines. Once it's a medium-sized function, be explicit with your return values. Corollary: it's not worth naming result parameters just because it enables you to use naked returns. Clarifying docs is always more important than saving a line or two in your function.

Finally, it would help if you named a result parameter in some cases to change it in a deferred closure. That is always OK.


### Package Comments

Package comments, like all comments to be presented by godoc, must appear adjacent to the package clause with no blank line.

```go
/*
Package template implements data-driven templates for generating textual
output such as HTML.
....
*/
package template
```

For "package main" comments, other styles of comment are fine after the binary name (and it may be capitalized if it comes first). For example, for a `package main` in the directory `seedgen` you could write:

```go
// Seedgen ..
package main
```

See https://go.dev/doc/effective_go#commentary for more information about commentary conventions.

### Package Names

All references to names in your package will be done using the package name so that you can omit that name from the identifiers. For example, if you are in package chubby, you don't need to type ChubbyFile, which clients will write as `chubby.ChubbyFile`. Instead, name the type `File`, which clients will write as `chubby.File`. Avoid meaningless package names like util, common, misc, API, types, and interfaces. See https://go.dev/doc/effective_go#package-names and https://go.dev/blog/package-names for more.

When naming packages, choose a name that is:

- All lowercase. No capitals or underscores.
- Does not need to be renamed using named imports at most call sites.
- Short and succinct. Remember that the name is identified in full at every call site.
- Not plural. For example, `net/url`, not `net/urls`.
- Not `common`, `util`, `shared`, or `lib`. These are bad, uninformative names.
- To distinguish SDK and Gaia with the same package name, add SDK or Gaia or the module name as the prefix. E.g.: `sdk/types`, `gaia/types` and `gaia/x/globalfee/types`, can use `sdktype`, `gaiatype`, `globalfeetype`.

See also [Package Names] and [Style guideline for Go packages].

[Package Names]: https://blog.golang.org/package-names
[Style guideline for Go packages]: https://rakyll.org/style-packages/

### Function Names

We follow the Go community's convention of using [MixedCaps for function names](https://golang.org/doc/effective_go.html#mixed-caps). An exception is made for test functions, which may contain underscores
for grouping related test cases, e.g., `TestMyFunction_WhatIsBeingTested`.

## Pointers

Try to avoid pointers if you don't need them. Don't pass pointers as function arguments to save a few bytes. If a function refers to its argument `x` only as `*x` throughout, then the argument shouldn't be a pointer. Common instances of this include passing a pointer to a string (`*string`) or a pointer to an interface value (`*io.Reader`). In both cases, the value itself is a fixed size and can be passed directly. This advice does not apply to large structs or even small structs that might grow.

Choosing whether to use a value or pointer receiver on methods can be difficult, especially for new Go programmers. If in doubt, use a pointer, but there are times when a value receiver makes sense, usually for reasons of efficiency, such as for small unchanging structs or values of basic type. Some useful guidelines:

* If the receiver is a map, func, or chan, don't use a pointer to them. If the receiver is a slice and the method doesn't reslice or reallocate the slice, don't use a pointer.
* If the method needs to mutate the receiver, the receiver must be a pointer.
* If the receiver is a struct that contains a sync.Mutex or similar synchronizing field, the receiver must be a pointer to avoid copying.
* A pointer receiver is more efficient if the receiver is a large struct or array. How large is large? Assume it's equivalent to passing all its elements as arguments to the method. If that feels too large, it's also too large for the receiver.
* Can functions or methods, either concurrently or when called from this method, mutate the receiver? A value type creates a copy of the receiver when the method is invoked, so outside updates will not be applied to this receiver. If changes must be visible in the original receiver, the receiver must be a pointer.
* If the receiver is a struct, array, or slice and any of its elements is a pointer to something that might be mutating, prefer a pointer receiver, as it will make the intention clearer to the reader.
* If the receiver is a small array or struct that is naturally a value type (for instance, something like the `time.Time` type), with no mutable fields and no pointers, or is just a simple basic type such as int or string, a value receiver makes sense. A value receiver can reduce the amount of garbage generated; if a value is passed to a value method, an on-stack copy can be used instead of allocating it to the heap. (The compiler tries to be smart about avoiding this allocation, but it can't always succeed.) Don't choose a value receiver type for this reason without profiling first.
* Don't mix receiver types. Choose either pointers or struct types for all available methods.
* Finally, when in doubt, use a pointer receiver.

### Receiver Names

The name of a method's receiver should be a reflection of its identity; often, a one or two-letter abbreviation of its type suffices (such as "c" or "cl" for "Client"). Don't use generic names such as "me", "this", or "self", identifiers typical of object-oriented languages that give the method a special meaning. In Go, the receiver of a method is just another parameter and, therefore, should be named accordingly. The name need not be as descriptive as that of a method argument, as its role is evident and serves no documentary purpose. It can be very short as it will appear on almost every line of every type of method; familiarity admits brevity. Be consistent, too: if you call the receiver "c" in one method, don't call it "cl" in another.

eg:

```go
func (s *IntegrationTestSuite) TestDecode()
```

### Variable Names

Variable names in Go should be short rather than long. This is especially true for local variables with limited scope. Prefer `c` to `lineCount`.  Prefer `i` to `sliceIndex`.

The basic rule: the further from its declaration that a name is used, the more descriptive the name must be. For a method receiver, one or two letters are sufficient. Common variables such as loop indices and readers can be a single letter (`i', `r`). More unusual things and global variables need more descriptive names.


### Zero-value Mutexes

The zero-value of `sync.Mutex` and `sync.RWMutex` is valid, so you rarely need a pointer to a mutex.

- Bad

```go
mu := new(sync.Mutex)
mu.Lock()
```

- Good

```go
var mu sync.Mutex
mu.Lock()
```

If you use a struct by pointer, then the mutex should be a non-pointer field. Do not embed the mutex on the struct, even if the struct is not exported.

- Bad

```go
type SMap struct {
  sync.Mutex

  data map[string]string
}

func (m *SMap) Get(k string) string {
  m.Lock()
  defer m.Unlock()

  return m.data[k]
}
```

The `Mutex` field and the `Lock` and `Unlock` methods are unintentionally part of the exported API of `SMap`.

- Good

```go
type SMap struct {
  mu sync.Mutex

  data map[string]string
}

func (m *SMap) Get(k string) string {
  m.mu.Lock()
  defer m.mu.Unlock()

  return m.data[k]
}
```

### Copy Slices and Maps at Boundaries

Slices and maps contain pointers to the underlying data, so be wary of scenarios when they need to be copied.

#### Receiving Slices and Maps

Remember that users can modify a map or slice you received as an argument if you store a reference to it.

- Bad

```go
func (d *Driver) SetTrips(trips []Trip) {
  d.trips = trips
}

trips := ...
d1.SetTrips(trips)

// Did you mean to modify d1.trips?
trips[0] = ...
```

- Good

```go
func (d *Driver) SetTrips(trips []Trip) {
  d.trips = make([]Trip, len(trips))
  copy(d.trips, trips)
}

trips := ...
d1.SetTrips(trips)

// We can now modify trips[0] without affecting d1.trips.
trips[0] = ...
```

#### Returning Slices and Maps

Similarly, be wary of user modifications to maps or slices exposing the internal state.

- Bad

```go
type Stats struct {
  mu sync.Mutex
  counters map[string]int
}

// Snapshot returns the current stats.
func (s *Stats) Snapshot() map[string]int {
  s.mu.Lock()
  defer s.mu.Unlock()

  return s.counters
}

// snapshot is no longer protected by the mutex, so any
// access to the snapshot is subject to data races.
snapshot := stats.Snapshot()
```

- Good

```go
type Stats struct {
  mu sync.Mutex
  counters map[string]int
}

func (s *Stats) Snapshot() map[string]int {
  s.mu.Lock()
  defer s.mu.Unlock()

  result := make(map[string]int, len(s.counters))
  for k, v := range s.counters {
    result[k] = v
  }
  return result
}

// Snapshot is now a copy.
snapshot := stats.Snapshot()
```

### Errors

#### Error Types

There are a few options for declaring errors. Consider the following before picking the option best suited for your use case.

- Does the caller need to match the error to handle it? If yes, we must support the [`errors.Is`] or [`errors.As`] functions by declaring a top-level error variable or a custom type.
- Is the error message a static string, or is it a dynamic string that requires contextual information? We can use [`errors.New`], but for the latter, we must use [`fmt.Errorf`] or a custom error type.
- Are we propagating a new error returned by a downstream function? See the [section on error wrapping](#error-wrapping).

[`errors.Is`]: https://golang.org/pkg/errors/#Is
[`errors.As`]: https://golang.org/pkg/errors/#As

| Error matching? | Error Message | Guidance                            |
|-----------------|---------------|-------------------------------------|
| No              | static        | [`errors.New`]                      |
| No              | dynamic       | [`fmt.Errorf`]                      |
| Yes             | static        | top-level `var` with [`errors.New`] |
| Yes             | dynamic       | custom `error` type                 |

[`errors.New`]: https://golang.org/pkg/errors/#New
[`fmt.Errorf`]: https://golang.org/pkg/fmt/#Errorf

For example, use [`errors.New`] for an error with a static string. Export this error as a variable to support matching it with `errors.Is` if the caller needs to match and handle this error.

- No error matching:

```go
// package foo

func Open() error {
  return errors.New("could not open")
}

// package bar

if err := foo.Open(); err != nil {
  //Can't handle the error.
  panic("unknown error")
}
```

- Error matching

```go
// package foo

var ErrCouldNotOpen = errors.New("could not open")

func Open() error {
  return ErrCouldNotOpen
}

// package bar

if err := foo.Open(); err != nil {
  if errors.Is(err, foo.ErrCouldNotOpen) {
    // handle the error
  } else {
    panic("unknown error")
  }
}
```

For an error with a dynamic string, use [`fmt.Errorf`] if the caller does not need to match it and a custom `error` if the caller does need to match it.

- No error matching

```go
// package foo

func Open(file string) error {
  return fmt.Errorf("file %q not found", file)
}

// package bar

if err := foo.Open("testfile.txt"); err != nil {
  //Can't handle the error.
  panic("unknown error")
}
```

- Error matching

```go
// package foo

type NotFoundError struct {
  File string
}

func (e *NotFoundError) Error() string {
  return fmt.Sprintf("file %q not found", e.File)
}

func Open(file string) error {
  return &NotFoundError{File: file}
}


// package bar

if err := foo.Open("testfile.txt"); err != nil {
  var notFound *NotFoundError
  if errors.As(err, &notFound) {
    // handle the error
  } else {
    panic("unknown error")
  }
}
```

Note that if you export error variables or types from a package, they will become part of the public API of the package.

#### Error Wrapping

There are three main options for propagating errors if a call fails:

- return the original error as-is
- add context with `fmt.Errorf` and the `%w` verb
- add context with `fmt.Errorf` and the `%v` verb

Return the original error as-is if there is no additional context to add. This maintains the original error type and message. This is well suited for cases when the underlying error message has sufficient information to track down where it came from.

Otherwise, add context to the error message where possible so that instead of a vague error such as "connection refused", you get more valuable errors such as "call service foo: connection refused".

Use `fmt.Errorf` to add context to your errors, picking between the `%w` or `%v` verbs based on whether the caller should be able to match and extract the underlying cause.

- Use `%w` if the caller should have access to the underlying error. This is a good default for most wrapped errors, but be aware that callers may begin to rely on this behavior. So for cases where the wrapped error is a known `var` or type, document and test it as part of your function's contract.
- Use `%v` to obfuscate the underlying error. Callers will be unable to match it, but you can switch to `%w` in the future if needed.

When adding context to returned errors, keep the context succinct by avoiding phrases like "failed to", which state the obvious and pile up as the error percolates up through the stack:

- Bad

```go
s, err := store.New()
if err != nil {
    return fmt.Errorf(
        "failed to create a new store: %w", err)
}
```

```
failed to x: failed to y: failed to create a new store: the error
```

- Good

```go
s, err := store.New()
if err != nil {
    return fmt.Errorf(
        "new store: %w", err)
}
```

```
x: y: new store: the error
```

However, once the error is sent to another system, it should be clear that the message is an error (e.g., an `err` tag or "Failed" prefix in logs).

See also [Don't just check errors, handle them gracefully].

[`"pkg/errors".Cause`]: https://godoc.org/github.com/pkg/errors#Cause
[Don't just check errors, handle them gracefully]: https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully

#### Error Naming

For error values stored as global variables, use the prefix `Err` or `err` depending on whether they're exported.

```go
var (
  // The following two errors are exported
  // so that users of this package can match them
  // with errors.Is.

  ErrBrokenLink = errors.New("link is broken")
  ErrCouldNotOpen = errors.New("could not open")

  // This error is not exported because
  // we don't want to make it part of our public API.
  // We may still use it inside the package
  // with errors.Is.

  errNotFound = errors.New("not found")
)
```

For custom error types, use the suffix `Error` instead.

```go
// Similarly, this error is exported
// so that users of this package can match it
// with errors.As.

type NotFoundError struct {
  File string
}

func (e *NotFoundError) Error() string {
  return fmt.Sprintf("file %q not found", e.File)
}

// And this error is not exported because
// we don't want to make it part of the public API.
// We can still use it inside the package
// with errors.As.

type resolveError struct {
  Path string
}

func (e *resolveError) Error() string {
  return fmt.Sprintf("resolve %q", e.Path)
}
```

### Handle Type Assertion Failures

The single return value form of a [type assertion] will panic on an incorrect type. Therefore, always use the "comma ok" idiom.

[type assertion]: https://golang.org/ref/spec#Type_assertions

- Bad

```go
t := i.(string)
```

- Good

```go
t, ok := i.(string)
if !ok {
  // handle the error gracefully
}
```


### Avoid Embedding Types in Public Structs

These embedded types leak implementation details, inhibit type evolution, and obscure documentation.

Assuming you have implemented a variety of list types using a shared `AbstractList`, avoid embedding the `AbstractList` in your concrete list implementations. Instead, hand-write only the methods to your concrete list that will delegate to the abstract list.

```go
type AbstractList struct {}

// Add adds an entity to the list.
func (l *AbstractList) Add(e Entity) {
  // ...
}

// Remove removes an entity from the list.
func (l *AbstractList) Remove(e Entity) {
  // ...
}
```

- Bad

```go
// ConcreteList is a list of entities.
type ConcreteList struct {
  *AbstractList
}
```

- Good

```go
// ConcreteList is a list of entities.
type ConcreteList struct {
  list *AbstractList
}

// Add adds an entity to the list.
func (l *ConcreteList) Add(e Entity) {
  l.list.Add(e)
}

// Remove removes an entity from the list.
func (l *ConcreteList) Remove(e Entity) {
  l.list.Remove(e)
}
```

Go allows [type embedding] as a compromise between inheritance and composition. The outer type gets implicit copies of the embedded type's methods. These methods, by default, delegate to the same method of the embedded instance.

[type embedding]: https://golang.org/doc/effective_go.html#embedding

The struct also gains a field by the same name as the type. So, if the embedded type is public, the field is public. To maintain backward compatibility, every future version of the outer type must keep the embedded type. An embedded type is rarely necessary. It is a convenience that helps you avoid writing tedious delegate methods.

Even embedding a compatible AbstractList *interface* instead of the struct would offer the developer more flexibility to change in the future but still leak the detail that the concrete lists use an abstract implementation.

- Bad

```go
// AbstractList is a generalized implementation
// for various kinds of lists of entities.
type AbstractList interface {
  Add(Entity)
  Remove(Entity)
}

// ConcreteList is a list of entities.
type ConcreteList struct {
  AbstractList
}
```

- Good

```go
// ConcreteList is a list of entities.
type ConcreteList struct {
  list AbstractList
}

// Add adds an entity to the list.
func (l *ConcreteList) Add(e Entity) {
  l.list.Add(e)
}

// Remove removes an entity from the list.
func (l *ConcreteList) Remove(e Entity) {
  l.list.Remove(e)
}
```

Either with an embedded struct or an embedded interface, the embedded type places limits on the evolution of the type.

- Adding methods to an embedded interface is a breaking change.
- Removing methods from an embedded struct is a breaking change.
- Removing the embedded type is a breaking change.
- Replacing the embedded type, even with an alternative that satisfies the same
  interface, is a breaking change.

Although writing these delegate methods is tedious, the additional effort hides an implementation detail, leaves more opportunities for change, and eliminates indirection for discovering the whole List interface in the documentation.

### Avoid `init()`

Avoid `init()` where possible. When `init()` is unavoidable or desirable, code should attempt to:

1. Be completely deterministic, regardless of program environment or invocation.
2. Avoid depending on the ordering or side-effects of other `init()` functions. While the `init()` order is well-known, code can change, and thus relationships between `init()` functions can make code brittle and error-prone.
3. Avoid accessing or manipulating global or environment states, such as machine information, environment variables, working directory, program arguments/inputs, etc.
4. Avoid I/O, including filesystem, network, and system calls.

Code that cannot satisfy these requirements likely belongs as a helper to be called as part of `main()` (or elsewhere in a program's lifecycle), or be written as part of `main()` itself. In particular, libraries intended to be used by other programs should take special care to be completely deterministic and not perform "init magic".

- Bad

```go
type Foo struct {
    // ...
}

var _defaultFoo Foo

func init() {
    _defaultFoo = Foo{
        // ...
    }
}
```

```go
type Config struct {
    // ...
}

var _config Config

func init() {
    // Bad: based on current directory
    cwd, _ := os.Getwd()

    // Bad: I/O
    raw, _ := os.ReadFile(
        path.Join(cwd, "config", "config.yaml"),
    )

    yaml.Unmarshal(raw, &_config)
}
```

- Good

```go
var _defaultFoo = Foo{
    // ...
}

// or, better, for testability:

var _defaultFoo = defaultFoo()

func defaultFoo() Foo {
    return Foo{
        // ...
    }
}
```

```go
type Config struct {
    // ...
}

func loadConfig() Config {
    cwd, err := os.Getwd()
    // handle err

    raw, err := os.ReadFile(
        path.Join(cwd, "config", "config.yaml"),
    )
    // handle err

    var config Config
    yaml.Unmarshal(raw, &config)

    return config
}
```

Considering the above, some situations in which `init()` may be preferable or necessary might include:

- Complex expressions that cannot be represented as single assignments.
- Pluggable hooks, such as `database/sql` dialects, encoding type registries, etc.
- Optimizations to [Google Cloud Functions] and other forms of deterministic precomputation.

  [Google Cloud Functions]: https://cloud.google.com/functions/docs/bestpractices/tips#use_global_variables_to_reuse_objects_in_future_invocations

## Performance

Performance-specific guidelines apply only to the hot path.

#### Prefer strconv over fmt

When converting primitives to/from strings, `strconv` is faster than `fmt`.

- Bad

```go
for i := 0; i < b.N; i++ {
  s := fmt.Sprint(rand.Int())
}
```

```
BenchmarkFmtSprint-4    143 ns/op    2 allocs/op
```

- Good

```go
for i := 0; i < b.N; i++ {
  s := strconv.Itoa(rand.Int())
}
```

```
BenchmarkStrconv-4    64.2 ns/op    1 allocs/op
```

**avoid use "+" for string concatenation**

#### Avoid string-to-byte conversion

Do not create byte slices from a fixed string repeatedly. Instead, perform the conversion once and capture the result.

- Bad

```go
for i := 0; i < b.N; i++ {
  w.Write([]byte("Hello world"))
}
```

```
BenchmarkBad-4   50000000   22.2 ns/op
```

- Good

```go
data := []byte("Hello world")
for i := 0; i < b.N; i++ {
  w.Write(data)
}
```

```
BenchmarkGood-4  500000000   3.25 ns/op
```

#### Prefer Specifying Container Capacity

Specify container capacity where possible to allocate memory for the container up front. This minimizes subsequent allocations (copying and resizing the container) as elements are added.

##### Specifying Map Capacity Hints

Provide capacity hints when initializing maps with `make()` where possible.

```go
make(map[T1]T2, hint)
```

Providing a capacity hint to `make()` tries to right-size the map at initialization time, which reduces the need for growing the map and allocations as elements are added to the map.

Unlike slices, map capacity hints do not guarantee complete, preemptive allocation but are used to approximate the number of hashmap buckets required. Consequently, allocations may still occur when adding elements to the map, even up to the specified capacity.

- Bad

```go
m := make(map[string]os.FileInfo)

files, _ := os.ReadDir("./files")
for _, f := range files {
    m[f.Name()] = f
}
```

`m' is created without a size hint; there may be more allocations at assignment time.

- Good

```go

files, _ := os.ReadDir("./files")

m := make(map[string]os.DirEntry, len(files))
for _, f := range files {
    m[f.Name()] = f
}
```

`m' is created with a size hint; there may be fewer allocations at assignment time.


##### Specifying Slice Capacity

Where possible, provide capacity hints when initializing slices with `make()`, particularly when appending.

```go
make([]T, length, capacity)
```

Unlike maps, slice capacity is not a hint: the compiler will allocate enough memory for the capacity of the slice as provided to `make()`, which means that subsequent `append()` operations will incur zero allocations (until the length of the slice matches the capacity, after which any appends will require a resize to hold additional elements).

- Bad

```go
for n := 0; n < b.N; n++ {
  data := make([]int, 0)
  for k := 0; k < size; k++{
    data = append(data, k)
  }
}
```

```
BenchmarkBad-4    100000000    2.48s
```

- Good

```go
for n := 0; n < b.N; n++ {
  data := make([]int, 0, size)
  for k := 0; k < size; k++{
    data = append(data, k)
  }
}
```

```
BenchmarkGood-4   100000000    0.21s
```

### Function Grouping and Ordering

- Functions should be sorted in rough call order.
- The receiver should group functions in a file.

Therefore, exported functions should appear first in a file, after `struct`, `const`, and `var` definitions.

A `newXYZ()`/`NewXYZ()` may appear after the type is defined but before the rest of the methods on the receiver.

Since the receiver groups functions, plain utility functions should appear toward the end of the file.

- Bad

```go
func (s *something) Cost() {
  return calcCost(s.weights)
}

type something struct{ ... }

func calcCost(n []int) int {...}

func (s *something) Stop() {...}

func newSomething() *something {
    return &something{}
}
```

- Good

```go
type something struct{ ... }

func newSomething() *something {
    return &something{}
}

func (s *something) Cost() {
  return calcCost(s.weights)
}

func (s *something) Stop() {...}

func calcCost(n []int) int {...}
```

### Reduce Nesting

Code should reduce nesting where possible by handling error cases/special conditions first and returning early or continuing the loop. Reduce the amount of code that is nested on multiple levels.

- Bad

```go
for _, v := range data {
  if v.F1 == 1 {
    v = process(v)
    if err := v.Call(); err == nil {
      v.Send()
    } else {
      return err
    }
  } else {
    log.Printf("Invalid v: %v", v)
  }
}
```

- Good

```go
for _, v := range data {
  if v.F1 != 1 {
    log.Printf("Invalid v: %v", v)
    continue
  }

  v = process(v)
  if err := v.Call(); err != nil {
    return err
  }
  v.Send()
}
```

### Writing Tests

Use table-driven tests with [subtests] to avoid duplicating code when the core
test logic is repetitive.

[subtests]: https://blog.golang.org/subtests

- Bad:

```go
// func TestSplitHostPort(t *testing.T)

host, port, err := net.SplitHostPort("192.0.2.0:8000")
require.NoError(t, err)
assert.Equal(t, "192.0.2.0", host)
assert.Equal(t, "8000", port)

host, port, err = net.SplitHostPort("192.0.2.0:http")
require.NoError(t, err)
assert.Equal(t, "192.0.2.0", host)
assert.Equal(t, "http", port)

host, port, err = net.SplitHostPort(":8000")
require.NoError(t, err)
assert.Equal(t, "", host)
assert.Equal(t, "8000", port)

host, port, err = net.SplitHostPort("1:8")
require.NoError(t, err)
assert.Equal(t, "1", host)
assert.Equal(t, "8", port)
```

- Good:

```go
// func TestSplitHostPort(t *testing.T)

tests := []struct{
  give     string
  wantHost string
  wantPort string
}{
  {
    give:     "192.0.2.0:8000",
    wantHost: "192.0.2.0",
    wantPort: "8000",
  },
  {
    give:     "192.0.2.0:http",
    wantHost: "192.0.2.0",
    wantPort: "http",
  },
  {
    give:     ":8000",
    wantHost: "",
    wantPort: "8000",
  },
  {
    give:     "1:8",
    wantHost: "1",
    wantPort: "8",
  },
}

for _, tt := range tests {
  t.Run(tt.give, func(t *testing.T) {
    host, port, err := net.SplitHostPort(tt.give)
    require.NoError(t, err)
    assert.Equal(t, tt.wantHost, host)
    assert.Equal(t, tt.wantPort, port)
  })
}
```

Test tables make it easier to add context to error messages, reduce duplicate logic, and add new test cases.

We follow the convention that the slice of structs is referred to as `tests` and each test case `tt`. Further, we encourage explicating the input and output  values for each test case with `give` and `want` prefixes.

```go
tests := []struct{
  give     string
  wantHost string
  wantPort string
}{
  // ...
}

for _, tt := range tests {
  // ...
}
```

Parallel tests, like some specialized loops (for example, those that spawn goroutines or capture references as part of the loop body), must take care to explicitly assign loop variables within the loop's scope to ensure that they hold the expected values.

```go
tests := []struct{
  give string
  // ...
}{
  // ...
}

for _, tt := range tests {
  tt := tt // for t.Parallel
  t.Run(tt.give, func(t *testing.T) {
    t.Parallel()
    // ...
  })
}
```

In the example above, we must declare a `tt` variable scoped to the loop iteration because of the use of `t.Parallel()` below. If we do not do that, most or all tests will receive an unexpected value for `tt` or a value that changes as they run.

#### Use Subtests

Always use subtest beside you are using or not table drive tests. This can reduce the scope of the tests and be more transparent and easy to maintain. Each small case of the tests should be a new subtest.

### Avoid writing directly in the stdout

Avoid writing logs directly to the stdout or stderr. Use a proper log package for it.
It's also easier to maintain. We don't need to find all prints and change the code if we need to change.

### Avoid panic

Avoid panic in simple and small methods; all errors should be handled on the top level and application, and we can decide if we will panic or not.
We can also create a proper panic recovery to close all states, open connection from the application, and graceful exit without breaking anything. 

