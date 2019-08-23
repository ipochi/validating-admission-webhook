package config

import (
	"strings"
	"testing"
)

func TestLoadConfigExactSuccess(t *testing.T) {
	configData := `
rules:
- name: "test-rule"
  jsonpath: "{.metadata.annotations['seccomp.security.alpha.kubernetes.io/defaultProfileName', 'seccomp.security.alpha.kubernetes.io/allowededProfileNames']}"
  regex: '(unconfined|\*|^$)'
  admissionMessage: "Disallowed to disable seccomp policy"
`

	_, err := read(strings.NewReader(configData))
	if err != nil {
		t.Errorf("Failed to load the config ... ")
	}
}

func TestValidateConfigSuccess(t *testing.T) {
	c := &Config{
		Rules: []Rule{
			Rule{
				Name:             "test",
				JSONPath:         "dummy-jsonpath",
				Regex:            "dummy-regex",
				AdmissionMessage: "admit it dummy",
			},
		},
	}

	err := ValidateConfig(c)
	if err != nil {
		t.Errorf("Failed to validate config ... ")
	}
}

func TestValidateConfigNoNameFail(t *testing.T) {
	c := &Config{
		Rules: []Rule{
			Rule{
				JSONPath:         "dummy-jsonpath",
				Regex:            "dummy-regex",
				AdmissionMessage: "admit",
			},
		},
	}

	err := ValidateConfig(c)
	if err == nil {
		t.Errorf("Failed to validate config ... ")
	}
}
