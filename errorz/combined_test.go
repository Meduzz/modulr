package errorz

import (
	"fmt"
	"strings"
	"testing"
)

func TestCombinedError(t *testing.T) {
	subject := NewError(nil)

	if subject.Error() != nil {
		t.Error("Subject returned an error")
	}

	first := fmt.Errorf("first error")
	subject.Append(first)

	if subject.Error() == nil {
		t.Error("Subject did not return an error")
	}

	if subject.Error() != first {
		t.Error("Subject returned more than first error")
	}

	second := fmt.Errorf("second error")
	subject.Append(second)

	if subject.Error() == nil {
		t.Error("Subject did not return an error")
	}

	if !strings.Contains(subject.Error().Error(), first.Error()) {
		t.Error("Subject did not contain first error")
	}

	if !strings.Contains(subject.Error().Error(), second.Error()) {
		t.Error("Subject did not contain second error")
	}
}
