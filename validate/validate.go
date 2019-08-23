package validate

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"github.com/ipochi/psp-validating-admission-webhook/config"
	"k8s.io/client-go/util/jsonpath"
	"regexp"
	"strings"
)

type ValidationRules struct {
	Rules []*Rule
}

type Rule struct {
	Name             string
	JSONPath         *jsonpath.JSONPath
	Regex            *regexp.Regexp
	AdmissionMessage string
}

func NewValidationRules() *ValidationRules {
	return &ValidationRules{
		Rules: make([]*Rule, 0),
	}
}

// LoadRules loads validation rules
func (v *ValidationRules) LoadRules(cf *config.Config) error {

	var errors []string

	// Load the rules
	for _, rule := range cf.Rules {
		newValidationRule := &Rule{
			Name:             rule.Name,
			AdmissionMessage: rule.AdmissionMessage,
		}
		newValidationRule.JSONPath = jsonpath.New(rule.Name)
		newValidationRule.JSONPath.AllowMissingKeys(true)

		err := newValidationRule.JSONPath.Parse(rule.JSONPath)
		if err != nil {
			glog.Errorf("Error parsing jsonpath : %s", err.Error())
			errors = append(errors, err.Error())
			continue
		}

		regex, err := regexp.Compile(rule.Regex)

		if err != nil {
			glog.Errorf("Error parsing regex : %s", err.Error())
			errors = append(errors, err.Error())
			continue
		}
		newValidationRule.Regex = regex

		v.Rules = append(v.Rules, newValidationRule)
	}
	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "\n"))
	}
	if len(v.Rules) == 0 {
		glog.Errorf("Error, no validation rules loaded")
		return fmt.Errorf("No rules loaded ::")
	}
	return nil
}

//ValidateAdmission checks whether to admit the AdmissionRequest or not
func (v *ValidationRules) ValidateAdmission(pspObject interface{}) (bool, string) {
	isAllowed := true
	admissionMessage := ""
	for _, rule := range v.Rules {
		buf := new(bytes.Buffer)
		err := rule.JSONPath.Execute(buf, pspObject)
		if err != nil {
			glog.Errorf("Error executing jsonpath rule :: %s", err.Error())
			continue
		}
		out := buf.String()
		patternFound := rule.Regex.MatchString(out)

		glog.Infof(" Matching string :: %s and match --- %t", out, patternFound)

		if patternFound {
			isAllowed = false
			admissionMessage = rule.AdmissionMessage
			break
		}
	}
	return isAllowed, admissionMessage
}
