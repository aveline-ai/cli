package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrimerVersionOf(t *testing.T) {
	assert.Equal(t, 3, primerVersionOf([]byte(`{"ok":true,"primer":{"version":3,"blocks":[]},"doc":{}}`)))
	assert.Equal(t, 0, primerVersionOf([]byte(`{"ok":true,"doc":{}}`)), "older server, no primer")
	assert.Equal(t, 0, primerVersionOf([]byte(`{"ok":true,"primer":{"version":0}}`)))
	assert.Equal(t, 0, primerVersionOf([]byte(`{"ok":true,"primer":{"version":"3"}}`)), "wrong type")
	assert.Equal(t, 0, primerVersionOf([]byte(`not json`)))
}

func TestPrimerNotice(t *testing.T) {
	cases := []struct {
		name      string
		prev, cur int
		want      bool
	}{
		{"first sight is silent", 0, 1, false},
		{"no primer from server is silent", 2, 0, false},
		{"unchanged is silent", 2, 2, false},
		{"moved forward speaks", 1, 2, true},
		{"moved back speaks too", 3, 2, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			msg, ok := primerNotice(c.prev, c.cur)
			assert.Equal(t, c.want, ok)
			if ok {
				assert.Contains(t, msg, "v"+itoa(c.prev)+" -> v"+itoa(c.cur))
			} else {
				assert.Empty(t, msg)
			}
		})
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
