# Retry [![Build Status](https://secure.travis-ci.org/adamliesko/retry.svg)](http://travis-ci.org/adamliesko/retry) [![Go Report Card](https://goreportcard.com/badge/github.com/adamliesko/retry)](https://goreportcard.com/report/github.com/adamliesko/retry) [![GoDoc](https://godoc.org/github.com/adamliesko/retry?status.svg)](https://godoc.org/github.com/adamliesko/retry) [![Coverage Status](https://img.shields.io/coveralls/adamliesko/retry.svg)](https://coveralls.io/r/adamliesko/retry?branch=master)

![Lego Batman gif retry](https://media.giphy.com/media/JJhiRdcYfcokU/giphy.gif)

Retry is a Go package which wraps a function and invokes it repeatedly, until it succeeds - not returning an error. 
Multiple retry-able options are provided, based on number of attempts, delay options between attempts, errors to retry 
on or ignore, post attempts callback etc. Usable for interaction with flaky web services and similar unreliable sources
of frustration.

## Installation

```
go get -u github.com/adamliesko/retry
```

## Usage

In the simplest and default configuration only about calling package level function `Do()`, with the desired function. 
If the failed function fails after 10 retries a custom error of Max Attempts reached is returned. An alternative is 
using a reusable Retryer value struct (=instance in OOP), which will be reset after each `Do()` method call.

```go

import "external"

func poll() error{
    return external.IsItDone() 
}
    
// equivalent calls
err1 := retry.Do(poll)
err2 := retry.New().Do(poll)
r := retry.New()
err3 := r.Do(poll)
```

The usual usage would be to use either directly function, which returns an error or wrap the function call with a function,
which sets the error according to the inner function output.

```go
// can be used directly
func poll() error{
    return external.IsItDone() 
}

// has to be wrapped
func pollNoError bool{
    return external.HasItSucceeded()
}

func wrappedPoll() error{
    if !pollNoError(){
        return errors.New("pollNoError has failed")
    }
    return nil
}

result := retry.Do(wrappedPoll)
```

### Options on Retryer (listed below in greater detail):
- constant sleep delay after a failure
- custom function sleep delay (e.g. exponential back off)
- recovery of panics
- calling ensure function, regardless of the Retryer's work inside, once that it finishes
- calling a custom function after each failure
- ignoring certain errors
- retrying only on certain errors

### Constant delay of of 100ms between failing attempts
```go
func poll() error { return external.IsItDone() }
    
err := retry.New(retry.Sleep(100))
result := r.Do(poll)
```

### Using an exponential back off (or any other custom function) after each failed attempt
```go
func poll() error { return external.IsItDone() }
        
sleepFn := func(attempts int) {
    sleep := time.Duration(2^attempts) * time.Millisecond
    time.Sleep(sleep)
}

err := retry.New(retry.SleepFn(sleepFn)).Do(poll)
```

### Calling an ensure function, which is called after whole Retryer execution
```go
func poll() error { return external.IsItDone() }
        
func ensure(err error){
    fmt.Println("ensure will be called regardless of err value")
}

err := retry.Do(poll, retry.Ensure(ensure))
```

### Ignoring failures with errors of listed types (whitelist) and considering them as success
```go
type MyError struct {}

func (e MyError) Error() string { return "this is my custom error" }
	
func poll() error { return external.IsItDone() }
        
err := retry.New(Not([]errors{MyError{}})).Do(poll)
```

### Retrying only on listed error types (blacklist), other errors will be considered as success
```go
type MyError struct {}

func (e MyError) Error() string { return "this is my custom error" }

func poll() error { return external.IsItDone() }
        
err := retry.New(On([]errors{MyError{}})).Do(poll)
```

### Retry allows to combine many options in one Retryer. The code block below will enable:

- recovery of panics
- attempting to call the function up to 15 times
- sleeping for 200 ms after each failed attempt
- printing the failures to the Stdout

```go
func poll() error { return external.IsItDone() }
     
failCallback := func(err error){ fmt.Println("failed with error", error) }

r := retry.New(retry.Sleep(200), retry.Tries(15), retry.Recover(), retry.AfterEachFail(failCallback)
err := r.Do(poll)
```

## License
See [LICENSE](LICENSE).
