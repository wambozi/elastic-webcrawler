package logging

import "testing"

import "fmt"

func TestGetIPAddr(t *testing.T) {
	IPAddr, err := GetIPAddr()
	if err != nil {
		t.Errorf("Unexpected Error when getting the IP address of the machine. Error: %s", err)
	}

	if fmt.Sprintf("%T", IPAddr) != "net.IP" {
		t.Errorf("\n%s:\n\n%s\n\n%s:\n\n%s", green("[expected]"), "net.IP", red("[actual]"), fmt.Sprintf("%T", IPAddr))
	}
}
