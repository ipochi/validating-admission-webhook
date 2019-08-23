package webhook

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/ipochi/psp-validating-admission-webhook/config"
	"github.com/ipochi/psp-validating-admission-webhook/validate"

	"k8s.io/api/admission/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// Validate_old to validate the PSP admission review request
// Please note that this was the earlier version of the code
// where i wanted to get a solution running with some hard coded values
// Validate is the one that is being used.
func Validate_old(requestedAdmissionReview *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {

	req := requestedAdmissionReview.Request

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
			annotations := pspObject.ObjectMeta.Annotations

			seccompAnnotationPresent := false
			for annotation, value := range annotations {
				if strings.Contains(annotation, "seccomp.security.alpha.kubernetes.io") {
					seccompAnnotationPresent = true
					glog.Info("Found seccomp annotation -- ")
					if strings.Contains(value, "unconfined") || strings.Contains(value, "*") {
						glog.Info("Bad seccomp value found")
						response := ToAdmissionResponse("Cannot admit pods with Seccomp profile unconfined or *")
						response.Allowed = false
						return response
					}
				}
			}
			if !seccompAnnotationPresent {
				glog.Info("Seccomp annotation not found")
				response := ToAdmissionResponse("Cannot admit pods if there is no seccomp annotation")
				response.Allowed = false
				return response
			}

		default:
			glog.Errorf("Kind=%v not supported", req.Kind.Kind)
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
