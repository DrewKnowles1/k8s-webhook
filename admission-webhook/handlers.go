package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func jsonFieldValidation(a *admissionv1.AdmissionRequest) error {
	// If json field is empty, return an error with a message.
	if a == nil {
		return errors.New("empty admission review was sent")
	}
	if a.RequestKind == nil {
		return errors.New("blah blash review was sent")
	}
	// if a.RequestKind.Kind == nil {
	// }
	return nil
}

func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {
	//Set headers
	w.Header().Set("Content-Type", "application/json")
	//Response message
	fmt.Fprintf(w, "%s", `{"msg": "server is healthy"}`)

}

//This function is going to check the labels on the json formatted kubernetes object posted to the /validate endpoint
func (app *application) validate(w http.ResponseWriter, r *http.Request) {
	//Webhooks are send a post, content type application/json
	//the post contains the kubernetesobnject as a json in the body. We recieve an object of AdmissionReview type,
	//and must also post one back

	//Instanciate Admission Review struct
	input := admissionv1.AdmissionReview{}

	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		//COme back to this and write out the helper function to log this properly
		//app.writeErrorMessage(w, "Unable to decode this post request: " + err.Error())
		app.errorLog.Println("Error, json cannot be read: " + err.Error())
		return
	}

	//Lets make sure the object is what we are expecting
	//Need to understand 'input.Request.RequestKind.Kind', is this gleamed from an example json request, jsut to get datastructure?
	//I get that its basically just reading the "Kind" field we would type out in a yaml

	err = jsonFieldValidation(input.Request)

	if err != nil {
		http.Error(w, "Invalid json input", http.StatusBadRequest)
		return
	}

	// if input.Request == nil || input.Request.RequestKind == nil {
	// 	http.Error(w, "Invalid json input", http.StatusMethodNotAllowed)
	// 	return
	// }

	switch input.Request.RequestKind.Kind {
	//Check if field contains the value of pod
	case "Pod":
		//Info logging
		app.infoLog.Println("Request came for object type of pod")

		//Instanciate new pod struct
		pod := v1.Pod{}

		//Default values will assume the pod fails below checks. We will run conditionals to verify labels
		//If pod json passes these checks (i.e. expected label is present), we will override these vars with true/a message
		//stating the pod is allowed
		var requestAllowed bool = false
		var respMsg string = "Denied because pod is missing label"

		//Again this input.Request.Object.Raw data structure, need to fully understand where its coming from
		//Probably based of an example json req - but verify this to solidify understanding
		if err := json.Unmarshal(input.Request.Object.Raw, &pod); err != nil {
			//Below is the same deal as above, need to write the helper method and sort this properly
			//app.writeErrorMessage(w, "Unable to marshall the raw payload into a pod object: " + err.Error())
			app.errorLog.Println("Unable to marshall the raw payload into a pod object: " + err.Error())
			return
		}

		//If there are any labels
		if len(pod.ObjectMeta.Labels) > 0 {
			//if there is a label with the key 'owner'
			if val, ok := pod.ObjectMeta.Labels["owner"]; ok {
				//if the key has a field set maybe??
				if val != "" {
					//above has verified that then pod has met our labeling criteria
					requestAllowed = true
					respMsg = "Pod Allowed because label owner is present"
				}
				app.infoLog.Println("Allowed because label is present ")
			}
		}

		output := admissionv1.AdmissionReview{

			Response: &admissionv1.AdmissionResponse{
				UID:     input.Request.UID,
				Allowed: requestAllowed,
				Result: &metav1.Status{
					Message: respMsg,
				},
			},
		}
		output.TypeMeta.Kind = input.TypeMeta.Kind
		output.TypeMeta.APIVersion = input.TypeMeta.APIVersion

		w.Header().Set("Content-Type", "application/json")

		resp, err := json.Marshal(output)

		if err != nil {
			//app.writeErrorMessage(w, "Unable to marshal the json object: "+err.Error())
			app.infoLog.Printf("Unable to marshal the json object: " + err.Error())
		}

		if _, err := w.Write(resp); err != nil {
			//app.writeErrorMessage(w, "Unable to send HTTP response: "+err.Error())
			app.infoLog.Printf("Unable to send HTTP response: " + err.Error())
			return
		}

	default:
		msg := fmt.Sprintf("Can not work with K8s %v objects, only with Pod", input.Request.RequestKind.Kind)
		//app.writeErrorMessage(w, msg)
		app.infoLog.Printf(msg)
	}

}
