package clog

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRegister(t *testing.T) {
	tests := []struct {
		name string
		run  func()
		want string
	}{
		{
			name: "success",
			run: func() {
				NewRegister("TestNewRegister_success",
					func(_ interface{}) (Logger, error) { return nil, nil },
				)
			},
			want: "",
		},
		{
			name: "nil register",
			run: func() {
				NewRegister("", nil)
			},
			want: "register is nil",
		},
		{
			name: "duplicated register",
			run: func() {
				NewRegister("TestNewRegister_duplicated_register",
					func(_ interface{}) (Logger, error) { return nil, nil },
				)
				NewRegister("TestNewRegister_duplicated_register",
					func(_ interface{}) (Logger, error) { return nil, nil },
				)
			},
			want: `register with mode "TestNewRegister_duplicated_register" already exists`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				err := recover()
				if tt.want == "" {
					assert.Nil(t, err)
				} else {
					assert.Equal(t, tt.want, err)
				}
			}()

			tt.run()
		})
	}
}

var _ Logger = (*noopLogger)(nil)

type noopLogger struct {
	mode  Mode
	level Level
}

func (l *noopLogger) Mode() Mode             { return l.mode }
func (l *noopLogger) Level() Level           { return l.level }
func (l *noopLogger) Write(_ Messager) error { return nil }

func TestNew(t *testing.T) {
	testModeGood := Mode("TestNew_good")
	testModeBad := Mode("TestNew_bad")
	defer func() {
		Remove(testModeGood)
		Remove(testModeBad)
	}()

	NewRegister(testModeGood,
		func(_ interface{}) (Logger, error) {
			return &noopLogger{
				mode: testModeGood,
			}, nil
		},
	)
	NewRegister(testModeBad,
		func(_ interface{}) (Logger, error) {
			return nil, errors.New("random error")
		},
	)

	tests := []struct {
		name string
		mode Mode
		want error
	}{
		{
			name: "success",
			mode: testModeGood,
			want: nil,
		},
		{
			name: "no register",
			mode: "no_register",
			want: errors.New(`no register for "no_register"`),
		},
		{
			name: "initialize error",
			mode: testModeBad,
			want: errors.New("initialize logger: random error"),
		},
		{
			name: "success overwrite",
			mode: testModeGood,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.mode, 10, nil)
			assert.Equal(t, tt.want, err)
		})
	}
}

func TestRemove(t *testing.T) {
	testMode1 := Mode("TestRemove1")
	NewRegister(testMode1, func(_ interface{}) (Logger, error) {
		return &noopLogger{
			mode: testMode1,
		}, nil
	})
	assert.Nil(t, New(testMode1, 10, nil))

	testMode2 := Mode("TestRemove2")
	NewRegister(testMode2, func(_ interface{}) (Logger, error) {
		return &noopLogger{
			mode: testMode2,
		}, nil
	})
	assert.Nil(t, New(testMode2, 10, nil))

	tests := []struct {
		name       string
		mode       Mode
		numLoggers int
	}{
		{
			name:       "remove nothing",
			mode:       "TestRemove_nothing",
			numLoggers: 2,
		},
		{
			name:       "remove one",
			mode:       testMode1,
			numLoggers: 1,
		},
		{
			name:       "remove two",
			mode:       testMode2,
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
