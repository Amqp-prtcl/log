# Log

## Installation

In order to add this library to your Go module run the following command:
```
go get http://github.com/Amqp-prtcl/log
```

## Usage

### Use a Logger

In order to create a Logger use
``` go
var logger = log.NewLogger()
```

And then simply use the Log function with the desired `LogLevel`
```go
logger.Log(L_Info, "this is an awesome log")
logger.Log(L_Warn, "this is also an %v", "awesome log")
```

Or you can use helper functions
```go
logger.Debug("this is a debug log")
logger.Info("this is a info log")
logger.Warn("this is a info log")
logger.Error("this is an %v %v", "error", "log")
```

There is also the `L_FATAL` `LogLevel` that acts like all the other ones
in a `Logger.Log` call but a call to `Logger.Fatal` will wait for any previous call to end before making his and will then make a call to `os.Exit(1)`.

### Outputs

Each new Logger (created by a call to `Log.NewLogger()`) creates a log manager that handles the actual parsing and outputting.
By default a log manager has no output so any log call to it will not result in any action.
For it to output anything Outputs must be added.

Here is an example of adding an Output to a logger:
```go
Logger.AddOutput(NewTextOutput(os.Stdout, F_Std, false))
```
This add a textOutput to the underlying log manager of the Logger.

The log library provides 3 outputs:
- textOutput
- JsonOutput
- FileOutput (which can be either text or json)

Custom Output can be created (see [Creating Custom Output](#custom-outputs))

### Modifier Functions

As said previously, each new `Logger` as its own underlying log manager, but the `Logger`s  returned by functions that modify its behavior (any function that returns another Logger) create a Logger with given parameters but the returned Logger has the same log Manager and thus the same outputs.

This means that when a Logger is closed, any Logger that share its log manager (or in other words any Logger that derive from the same original Logger returned by the `Log.NewLogger()` call) are also closed.

### Sync and Async

By default a newly created Logger is Async, this means that any log call (except for a `Logger.Fatal` call) will return immediately and the log entry will be parsed sometime into the future.

On the contrary, a Sync Logger will wait for the log entry to be completely done before resuming code executing

Please Note that all log calls (Sync and Async) are buffered by the log manager and parsed in order of arriving this means that any Sync log call will wait any previous Async log call.

Please Also Note that if the log manager buffer is full Async log calls will wait for a space to be freed form the buffer before resuming execution. The default size is 10 but it can be changed with `Log.NewLoggerWithCapacity(capacity int)`

### Custom Outputs

// TODO: add doc