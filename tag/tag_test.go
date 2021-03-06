package tag

import (
	"testing"
)

func TestParse(t *testing.T) {
	validCases := map[string][]string{
		"server-state:live":                          {"server", "state:live"},
		"discovery-status:live":                      {"discovery", "status:live"},
		"server-dc:lhr":                              {"server", "dc:lhr"},
		"lb-pool:www":                                {"lb", "pool:www"},
		"custom-favorites:btyler":                    {"custom", "favorites:btyler"},
		"server-interfaces:eth1:ip_address:10_1_2_3": {"server", "interfaces:eth1:ip_address:10_1_2_3"},
	}

	for valid, expected := range validCases {
		service, kv, err := Parse(valid)
		if err != nil {
			t.Errorf("tag test: %q failed to parse: %q", valid, err)
			continue
		}

		if service != expected[0] {
			t.Errorf("tag test: %q ought to have service %q, but it has %q instead", valid, expected[0], service)
		}

		if kv != expected[1] {
			t.Errorf("tag test: %q ought to have kv %q, but it has %q instead", valid, expected[1], kv)
		}
	}

	invalidCases := []string{
		"asdfasdfaqwerioqwr",
		"::::-:--:;;;:0",
		"dc:lhr",
		"server",
		"btyler:favorites-custom",
		"btyler:favorites-custom",
		"server-interfaces:eth1:ip_address:10.1.2.3",
	}

	for _, invalid := range invalidCases {
		service, kv, err := Parse(invalid)
		if err == nil {
			t.Errorf("tag test: %q failed to error while parsing. tokens: %q %q", invalid, service, kv)
		}
	}
}
