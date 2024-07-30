# Contributor Guidelines - Logging

## Framework

We use structured logging, and specifically we use [Zap](https://github.com/uber-go/zap)
as our logging framework.

Why not something else?

- [Zerolog](https://github.com/rs/zerolog) is more performant and better
  looking, but its syntax is more error-prone (no logging in case you forget
  to call `Msg("...")` at the end)
- Go's [Slog](https://pkg.go.dev/log/slog) is new, slower (according to Zap
  benchmarks), and doesn't provide useful benefits

## Default Level

The default minimum logging level is [Info](https://pkg.go.dev/go.uber.org/zap/zapcore#InfoLevel).
To change this, set the following AutoKitteh config value to the lower-case or
all-caps name of the level you want. For example:

- CLI: `ak config set logger.level debug`

- Config YAML file:

  ```yaml
  logger:
    level: debug
  ```

- Environment variable: `AK_LOGGER__LEVEL=DEBUG`

## Logger Naming

- Local variables: `l` is preferable when its meaning is clear, otherwise
  `logger` is acceptable too

  - `z` is discouraged - we use Zap, but we don’t want the code to be tied to
    that assumption
  - Other combinations of `l` and `logger` are possible for code readability,
    but not recommended in general

- Struct fields: `logger` or `Logger`

## Logger Initialization

> [!NOTE]
> TODO: Refactor `extrazap`
>
> TODO: Godoc link below

Pass a logger reference (`*zap.Logger`) as a function parameter, or use
[extrazap](/TODO) to initialize a logger instance:

```go
func one(l *zap.Logger, ...) {
    l.Info("...")

    ctx := extrazap.AttachLoggerToContext(context.Background(), l)
    two(ctx, ...)
}

func two(ctx context.Context, ...) {
    l := extrazap.ExtractLoggerFromContext(ctx)
    l.Info("...")
}
```

You can use `extrazap` even if your function doesn't have a context with an
embedded logger:

```go
func three() {
    l := extrazap.ExtractLoggerFromContext(context.Background())
    l.Info("...")
}
```

In other words, you **should** attach a logger to a context before passing it
to a function that uses `extrazap`, but it will work even if you don't.

Because of that, there's no need to call [zap.L()](https://pkg.go.dev/go.uber.org/zap#L)
directly.

Anyway, do not pass both a context and a logger as function parameters, this
is wasteful.

## Log Messages

1. Should be short
   - Avoid complex or lengthy sentences
   - Prefer adding details as fields, not in the message string
2. Should be informative - if the message is too generic or vague, add vital
   details to it (in addition to adding them as fields)
3. Should begin with a capital letter
   - This is unlike Go error strings, but for the
     [same reason](https://go.dev/wiki/CodeReviewComments#error-strings)
4. Shouldn’t end with punctuation
   - This is like Go error strings, but only to support (1) above

## Adding Log Fields to the Logger

As noted above, we always use loggers by reference (`*zap.Logger`), not by
value (`zap.Logger`).

Because of that, when you call [Logger.With()](https://pkg.go.dev/go.uber.org/zap#Logger.With)
in a nested code block, redeclare the logger with the `:=` operator:

```go
l := l.With(...)
```

Instead of just reassigning it with the `=` operator:

```go
l = l.With(...)
```

Whenever possible, this ensures that the added fields don't remain in the
logger after leaving exiting the code block.

Example:

```go
func foo(l *zap.Logger) {
	l.Info("...", zap.String("aaa", "111"))

    // Good: "bbb" doesn't remain after the loop's block.
	for _, s := range []string{"ccc"} {
		l := l.With(zap.String("bbb", "222"))
		l.Info("...", zap.String(s, "333"))
	}

    // Bad: "ddd" remains after the condition's block.
    if ... {
		l := l.With(zap.String("ddd", "444"))
		l.Info("...", zap.String("eee", "555"))
    }

	l.Info("...", zap.String("fff", "666"))
}
```

## Secure Error Handling and Logging

OWASP secure coding practices checklist - error handling and logging:
https://owasp.org/www-project-secure-coding-practices-quick-reference-guide/stable-en/02-checklist/05-checklist#error-handling-and-logging

## Other Recommendations

Usually, you should write logs as close as possible to the reason for logging.

For example, this:

```go
func foo() {
    // ...
    if err := bar(); err != nil {
        return
    }
    // ...
}

func bar() error {
    if problem1 {
        l.Error("Boom")
        return errors.New("problem 1")
    }

    if err := problem2(); err != nil {
        l.Error("Pow", zap.Error(err))
        return fmt.Errorf("problem 2: %w", err)
    }

    l.Info("Successful bar")
    return nil
}
```

Is better than this:

```go
func foo() {
    // ...
    if err := bar(); err != nil {
        l.Error("Bar failed", zap.Error(err))
        return
    }
    l.Info("Successful bar")
    // ...
}

func bar() error {
    if problem1 {
        return errors.New("problem 1")
    }

    if err := problem2(); err != nil {
        return fmt.Errorf("problem 2: %w", err)
    }

    return nil
}
```

Unless you actually want all the different error conditions inside `bar()` to
result in the same common log entry, to keep `bar()` as a "black box".

Anyway, use your judgement.
