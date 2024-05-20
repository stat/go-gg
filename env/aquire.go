package env

//
// Imports
//

import (
	"fmt"
	"os"

	"github.com/jinzhu/copier"
)

//
// Vars
//

var (
	DefaultOptions = &AquireOptions{}

	// Errors
	AquireEmptyError     = &aquireEmptyError{}
	AquireUnmarshalError = &aquireUnmarshalError{}
)

//
// Errors
//

// Empty

type aquireEmptyError struct {
	Key string
}

func (err *aquireEmptyError) Error() string {
	return fmt.Sprintf("Could not aquire key: %s, aquired value is empty", err.Key)
}

// Unmarshal

type aquireUnmarshalError struct {
	Key            string
	UnmarshalError error
}

func (err *aquireUnmarshalError) Error() string {
	return fmt.Errorf("Could not unmarshal key: %s due to %w", err.Key, err.UnmarshalError).Error()
}

//
// Options
//

type AquireOptions struct {
	IgnoreEmpty bool
}

func (options *AquireOptions) Load(opts ...AquireOption) error {
	// copy defaults

	if err := copier.Copy(options, DefaultOptions); err != nil {
		return err
	}

	// iter opts

	for _, opt := range opts {
		if err := opt(options); err != nil {
			return err
		}
	}

	// success

	return nil
}

// Option

type AquireOption func(*AquireOptions) error

// Ignore Empty

func IgnoreEmpty() AquireOption {
	return func(o *AquireOptions) error {
		o.IgnoreEmpty = true
		return nil
	}
}

//
// Aquire
//

func Aquire[T any](key string, opts ...AquireOption) (*T, error) {
	options := &AquireOptions{}

	if err := options.Load(opts...); err != nil {
		return nil, err
	}

	return AquireWithOptions[T](key, options)
}

// With Options

func AquireWithOptions[T any](key string, options *AquireOptions) (*T, error) {
	// read cache

	iface, found := aquired[key]
	value, ok := iface.(T)

	if found && ok {
		return &value, nil
	}

	// read env

	data := os.Getenv(key)

	if data == "" && !options.IgnoreEmpty {
		return nil, &aquireEmptyError{key}
	}

	// unmarshal

	v, err := Unmarshal[T]([]byte(data))

	if err != nil {
		return nil, &aquireUnmarshalError{Key: key, UnmarshalError: err}
	}

	aquired[key] = v

	return v, nil
}
