package validate

import (
	"encoding/json"
	"github.com/ipochi/psp-validating-admission-webhook/config"
	"testing"
)

func TestLoadRulesSuccess(t *testing.T) {

	v := NewValidationRules()
	c := &config.Config{
		Rules: []config.Rule{
			config.Rule{
				Name:     "test",
				JSONPath: "{.metadata.annotations['hello.this.is.test', 'this.is.another.test.annotation']}",
				Regex:    "hello",
			},
		},
	}
	err := v.LoadRules(c)
	if err != nil {
		t.Errorf("Failed to validate config ... ")
	}
}

func TestLoadRulesFail(t *testing.T) {
	v := NewValidationRules()
	c := &config.Config{
		Rules: []config.Rule{
			config.Rule{
				Name:     "test",
				JSONPath: "{.metadata.annotations['hello.this.is.test', 'this.is.another.test.annotation']}",
				Regex:    "(?<=\\s)",
			},
		},
	}
	err := v.LoadRules(c)
	if err == nil {
		t.Errorf("Invalid regex ... ")
	}
}

func TestValidateAdmissionDontAllow(t *testing.T) {

	objstring := `{"annotations": "annotation1,annotation2"}`
	var obj interface{}
	_ = json.Unmarshal([]byte(objstring), &obj)

	v := NewValidationRules()
	c := &config.Config{
		Rules: []config.Rule{
			config.Rule{
				Name:             "test",
				JSONPath:         "{.annotations}",
				Regex:            "annotation1",
				AdmissionMessage: "not allowed ...",
			},
		},
	}
	err := v.LoadRules(c)
	if err != nil {
		t.Errorf("no rules loaded ... ")
	}

	allow, message := v.ValidateAdmission(obj)
	if allow {
		t.Errorf("Should not be allowed ... ")
	}
	if message != c.Rules[0].AdmissionMessage {
		t.Errorf("Admission message not matching ... ")
	}
}

func TestValidateAdmissionShouldAllow(t *testing.T) {

	objstring := `{"annotations": "annotation1,annotation2"}`
	var obj interface{}
	_ = json.Unmarshal([]byte(objstring), &obj)

	v := NewValidationRules()
	c := &config.Config{
		Rules: []config.Rule{
			config.Rule{
				Name:             "test",
				JSONPath:         "{.annotations}",
				Regex:            "annotaton3",
				AdmissionMessage: "allowed ...",
			},
		},
	}
	err := v.LoadRules(c)
	if err != nil {
		t.Errorf("no rules loaded ... ")
	}

	allow, _ := v.ValidateAdmission(obj)
	if !allow {
		t.Errorf("Should be allowed ... ")
	}
}
