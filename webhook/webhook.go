package webhook

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/ipochi/psp-validating-admission-webhook/config"
	"github.com/ipochi/psp-validating-admission-webhook/validate"

	"k8s.io/api/admission/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Validate(requestedAdmissionReview *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {

	req := requestedAdmissionReview.Request

	// Load up the config for psp-admission-controller
	cf, err := config.LoadConfig()
	if err != nil {
		glog.Fatalf("Error loading config ...  %s", err.Error())
	}

	err = config.ValidateConfig(cf)
	if err != nil {
		glog.Fatalf("Config file couldnt not be validated :: %s", err.Error())
	}

	validationRules := validate.NewValidationRules()
	err = validationRules.LoadRules(cf)
	if err != nil {
		glog.Fatalf("Error loading rules ...  %s", err.Error())
	}

	switch req.Operation {
	case "CREATE", "UPDATE":
		switch req.Kind.Kind {
		case "PodSecurityPolicy":
			var pspObject policyv1beta1.PodSecurityPolicy
			err := json.Unmarshal(req.Object.Raw, &pspObject)
			if err != nil {
				glog.Errorf("Could not unmarshal raw object: %v", err)
				response := ToAdmissionResponse(err.Error())
				response.Allowed = false
				return response
			}

			validationResult, admissionMessage := validationRules.ValidateAdmission(pspObject)
			response := ToAdmissionResponse(admissionMessage)
			response.Allowed = validationResult
			return response

		default:
			glog.Errorf("Unsupported Kind :: %s", req.Kind.Kind)
			response := ToAdmissionResponse("Kind is not supported")
			response.Allowed = false
			return response
		}
	default:
		glog.Errorf("Operation=%s not supported", req.Operation)
		response := ToAdmissionResponse("Operation not permitted")
		response.Allowed = false
		return response
	}

	response := ToAdmissionResponse("All Good , allowed.")
	response.Allowed = true
	return response
}

func ToAdmissionResponse(msg string) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: msg,
		},
	}
}
