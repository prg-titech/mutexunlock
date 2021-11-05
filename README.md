# pkg `unlockcheck`

![test](../../actions/workflows/test.yml/badge.svg)

Automated fix tool for missing `sync.Mutex` or `sync.RWMutex` unlocks.

## Installation

```
# go install github.com/Qs-F/unlockcheck/cmd/unlockcheck
```

## Usgae

### Lint (without automated fix)

Run in the target project (if you are not sure, run in the directory which contains `go.mod`)

```
# go vet -vettool=$(which unlockcheck) .
```

### Fmt (with automated fix)

For some reasons `go vet` command does not provide to pass flags to external vettool.  

Run in the target project (if you are not sure, run in the directory which contains `go.mod`)

```
# unlockcheck -fix .
```

## Example Behavior

If target code is like this:

```go
type S struct {
	mu sync.Mutex
}

func (s *S) D() {
	s.mu.Lock()
	// You forgot s.mu.Unlock() here!
}
```

After [Fmt (with automated fix)](#Fmt-(with-automated-fix)),

```diff go
type S struct {
	mu sync.Mutex
}

func (s *S) D() {
	s.mu.Lock()
+	s.mu.Unlock()
	// You forgot s.mu.Unlock() here!
}
```

## Auhtor

Qs-F

## License

MIT License
