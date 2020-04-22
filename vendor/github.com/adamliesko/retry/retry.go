package retry

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"time"
)

// MaxRetries is the maximum number of retries.
const MaxRetries = 10

// Retryer is configurable runner, which repeats function calls until it succeeds.
type Retryer struct {
	Tries    int
	On       []error       // On is the slice of errors, on which Retryer will retry a function
	Not      []error       // Not is the slice of errors which Retryer won't consider as needed to retry
	SleepDur time.Duration // Sleep duration in ms
	Recover  bool          // If enabled, panics will be recovered.

	SleepFn         func(int)   // Custom sleep function with access to the current # of attempts
	EnsureFn        func(error) // DeferredFn is called after repeated function finishes, regardless of outcome
	AfterEachFailFn func(error) // Callback called after each of the failures (for example some logging)

	attempts int
}

// Do is wrapper around Retryer, which doesn't expose the Retryer itself, only calls the function until it succeeds.
func Do(fn func() error, opts ...func(*Retryer)) error {
	r := New(opts...)
	return r.Do(fn)
}

// New creates a Retryer with applied options.
func New(opts ...func(*Retryer)) *Retryer {
	r := &Retryer{Tries: MaxRetries}

	// setup the options
	for _, o := range opts {
		o(r)
	}

	return r
}

// Reset resets the state of the Retryer to the default starting one, resetting the number of attempts to 0.
func (r *Retryer) Reset() {
	r.attempts = 0
}

// Do calls the passed in function until it succeeds. The behaviour of the retry mechanism heavily relies on the config
// of the Retryer.
func (r *Retryer) Do(fn func() error) (err error) {
	// reset the state to starting one, 0 attempts
	r.Reset()

	// define the deferred functions
	if r.Recover {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("retryer has recovered panic: %v %s", r, debug.Stack())
			}
		}()
	}
	if r.EnsureFn != nil {
		defer r.EnsureFn(err)
	}

	// retry the function
	for {
		if r.attempts >= r.Tries {
			break
		}
		r.attempts++

		err = fn()
		if r.succeeded(err) {
			return nil
		}
		if r.AfterEachFailFn != nil {
			r.AfterEachFailFn(err)
		}
		r.trySleep()
	}

	return fmt.Errorf("max number of retries reached: %d, last error %v", r.attempts, err)
}

// Attempts return the number of times Retryer has invoked a function call.
func (r *Retryer) Attempts() int {
	return r.attempts
}

func (r *Retryer) succeeded(err error) bool {
	for _, e := range r.Not {
		if reflect.TypeOf(err) == reflect.TypeOf(e) {
			return true
		}
	}
	for _, e := range r.On {
		if reflect.TypeOf(err) == reflect.TypeOf(e) {
			return false
		}
	}

	if len(r.On) > 0 {
		return true
	}

	return err == nil
}

func (r *Retryer) trySleep() {
	if r.SleepFn != nil {
		r.SleepFn(r.attempts)
	} else if r.SleepDur != 0 {
		time.Sleep(r.SleepDur)
	}
}
