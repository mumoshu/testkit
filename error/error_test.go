package error_test

import (
	"fmt"
	"testing"

	"github.com/mumoshu/testkit/error"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	//lint:ignore S1025 we want to explicitly show that how it is formatted when used with Sprintf
	assert.Equal(t, "foo", fmt.Sprintf("%s", error.New("foo")))
	assert.Equal(t, "foo", fmt.Sprintf("%v", error.New("foo")))

	assert.Equal(t, `╷
│ foo
│
│ bar
╵`, error.New("foo", error.Long("bar")).String())

	assert.Equal(t, `╷
│ foo
│
│   on somefile.json line 2:
│    2: some code
│
│ bar
╵`, error.New("foo",
		error.Source("somefile.json", 2, "some code"),
		error.Long("bar"),
	).String())
}

func ExampleError() {
	var e = error.New(
		"unable to get file from git repository `foo`",
		error.Source("some.json", 2, "some code"),
		error.Long("This application requires access to the file in the repository `foo` for reading the configuration, but was unable to do so due to an internal error."),
		error.Remediation(`Please make sure that the repository is accessible to the application.`),
	)

	fmt.Println("Error() returns the short message:")
	fmt.Println(e)

	fmt.Println("String() returns the full message:")
	fmt.Println(e.String())

	// Output:
	// Error() returns the short message:
	// unable to get file from git repository `foo`
	// String() returns the full message:
	// ╷
	// │ unable to get file from git repository `foo`
	// │
	// │   on some.json line 2:
	// │    2: some code
	// │
	// │ This application requires access to the file in the repository `foo` for reading the configuration, but was unable to do so due to an internal error.
	// │
	// │ Please make sure that the repository is accessible to the application.
	// ╵
}
