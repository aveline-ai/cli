package slug

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrom(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Hello", "hello"},
		{"Hello World", "hello-world"},
		{"  leading and trailing  ", "leading-and-trailing"},
		{"multiple   spaces", "multiple-spaces"},
		{"Punc!t!u@a$tion###", "punc-t-u-a-tion"},
		{"already-dashed", "already-dashed"},
		{"--edges--", "edges"},
		{"MiXeD CaSe 123", "mixed-case-123"},
		{"a/b\\c", "a-b-c"},
	}
	for _, c := range cases {
		got, err := From(c.in)
		require.NoError(t, err, "input=%q", c.in)
		assert.Equal(t, c.want, got, "input=%q", c.in)
	}
}

func TestFromEmpty(t *testing.T) {
	for _, in := range []string{"", "   ", "!!!", "---", "你好"} {
		_, err := From(in)
		assert.ErrorIs(t, err, ErrEmpty, "input=%q", in)
	}
}
