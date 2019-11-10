package clog

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	testGood := "TestNew_good"
	testGoodIniter := noopIniter(testGood)

	testBad := "TestNew_bad"
	testBadIniter := func(string, ...interface{}) (Logger, error) {
		return nil, errors.New("random error")
	}

	defer Remove(testGood)
	defer Remove(testBad)

	tests := []struct {
		name       string
		mode       string
		initer     Initer
		bufferSize interface{}
		want       error
	}{
		{
			name:       "success",
			mode:       testGood,
			initer:     testGoodIniter,
			bufferSize: 1,
			want:       nil,
		},
		{
			name:       "success",
			mode:       testGood,
			initer:     testGoodIniter,
			bufferSize: int32(1),
			want:       nil,
		},
		{
			name:       "success",
			mode:       testGood,
			initer:     testGoodIniter,
			bufferSize: int64(1),
			want:       nil,
		},
		{
			name:   "initialize error",
			mode:   testBad,
			initer: testBadIniter,
			want:   errors.New("initialize logger: random error"),
		},
		{
			name:   "success overwrite",
			mode:   testGood,
			initer: testGoodIniter,
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.mode, tt.initer, tt.bufferSize)
			assert.Equal(t, tt.want, err)
		})
	}
}

func TestRemove(t *testing.T) {
	test1 := "TestRemove_1"
	test1Initer := noopIniter(test1)
	assert.Nil(t, New(test1, test1Initer, -1))

	test2 := "TestRemove_2"
	test2Initer := noopIniter(test2)
	assert.Nil(t, New(test2, test2Initer, -1))

	tests := []struct {
		name       string
		mode       string
		numLoggers int
	}{
		{
			name:       "remove nothing",
			mode:       "TestRemove_nothing",
			numLoggers: 2,
		},
		{
			name:       "remove one",
			mode:       test1,
			numLoggers: 1,
		},
		{
			name:       "remove two",
			mode:       test2,
			numLoggers: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Remove(tt.mode)
			assert.Equal(t, tt.numLoggers, mgr.len())
		})
	}
}
