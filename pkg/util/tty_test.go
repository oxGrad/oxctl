package util_test

import (
	"os"
	"testing"

	"github.com/oxGrad/oxctl/pkg/util"
)

func TestIsCI_WhenEnvSet(t *testing.T) {
	t.Setenv("CI", "true")
	if !util.IsCI() {
		t.Fatal("expected IsCI() == true when CI=true")
	}
}

func TestIsCI_WhenEnvUnset(t *testing.T) {
	os.Unsetenv("CI")
	if util.IsCI() {
		t.Fatal("expected IsCI() == false when CI is unset")
	}
}
