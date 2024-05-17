package geancount

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsOpenIsClosed(t *testing.T) {
	acc := Account{
		opened: []time.Time{
			time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2000, time.January, 4, 0, 0, 0, 0, time.UTC),
			time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC),
			time.Date(2000, time.January, 8, 0, 0, 0, 0, time.UTC),
		},
		closed: []time.Time{
			time.Date(2000, time.January, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2000, time.January, 5, 0, 0, 0, 0, time.UTC),
			time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC),
			time.Date(2000, time.January, 10, 0, 0, 0, 0, time.UTC),
		},
	}
	//
	assert.True(t, acc.IsClosed(time.Date(1999, time.December, 31, 0, 0, 0, 0, time.UTC)), "Before first open is closed")
	assert.True(t, acc.IsOpen(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)), "On the open date is open")
	assert.True(t, acc.IsOpen(time.Date(2000, time.January, 2, 0, 0, 0, 0, time.UTC)), "On the close date is open")
	assert.True(t, acc.IsClosed(time.Date(2000, time.January, 3, 0, 0, 0, 0, time.UTC)), "After is closed is closed")
	assert.True(t, acc.IsOpen(time.Date(2000, time.January, 4, 0, 0, 0, 0, time.UTC)), "After is reopen is open")
	assert.True(t, acc.IsClosed(time.Date(2000, time.January, 6, 0, 0, 0, 0, time.UTC)), "After is closed is closed")
	assert.True(t, acc.IsOpen(time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)), "Open and close on the same date is open")
	assert.True(t, acc.IsOpen(time.Date(2000, time.January, 8, 0, 0, 0, 0, time.UTC)), "After is reopen is open")
	assert.True(t, acc.IsOpen(time.Date(2000, time.January, 9, 0, 0, 0, 0, time.UTC)), "After is reopen is open")
	assert.True(t, acc.IsOpen(time.Date(2000, time.January, 10, 0, 0, 0, 0, time.UTC)), "On the close date is open")

	assert.True(t, acc.IsClosed(time.Date(2000, time.January, 11, 0, 0, 0, 0, time.UTC)),
		"After the last close is closed")
}
