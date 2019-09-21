# PodSecurityPolicy validating admission webhook server

## Slides
This topic was presented by me at a Meetup in Bangalore on 21st September 2019.

Slides for the same are [here](https://docs.google.com/presentation/d/1lfCKQxgseX3FXVgLT-UPbvHP8hnesgI8_aASlOYPKVk/edit?usp=sharing) 
## Deployment steps.

1. Start minikube with PodSecurityPolicy enabled.

   Reference - https://suraj.io/post/psp-on-existing-cluster/

    ```
    minikube start --extra-config=apiserver.enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,ResourceQuota,DefaultTolerationSeconds,PodSecurityPolicy,MutatingAdmissionWebhook,ValidatingAdmissionWebhook
    ```

2. Once the server is started, due to the enabling of PSP, we need to provide policies. 
   
   Reference -  https://github.com/appscodelabs/tasty-kube/blob/dc7b32a3ee8375f03218f6b10c3b51aa82c91a96/minikube/1.10/psp/README.md

    ```
    $ kubectl apply -f manifests/minikube-yamls-after-enabling-psp/psp.yaml
    $ kubectl auth reconcile -f manifests/minikube-yamls-after-enabling-psp/clusterroles.yaml
    $ kubectl auth reconcile -f manifests/minikube-yamls-after-enabling-psp/rolebindings.yaml
    ```

3. Building docker image locally
    ```
    $ eval $(minikube docker-env) # This will use the docker daemon of minikube
    $ docker build -t webhookserver .
    ```
4. Deploying the all-in-one-yaml manifest

    ```
    $ kubectl apply -f manifests/solution-deployment-yamls/all-in-one.yaml
    ```

5. Generate TLS certs for our webhookserver and create a secret
   Copy the certificate string to template and create a `ValidatingWebhookConfiguration`

   Reference- https://github.com/banzaicloud/admission-webhook-example
  
    ```
    $ cd manifests/solution-deployment-yamls/certs
    $ ./webhook-create-signed-cert.sh
    $ cat validatingwebhook.yaml.template | ./webhook-patch-ca-bundle.sh > webhookconfiguration.yaml
    $ kubectl apply -f webhookconfiguration.yaml
    ```
6. Test the solution with some test yamls
    ```
    $ cd manifests/test-yamls
    $ kubectl apply -f invalid-profiles.yaml
    $ kubectl apply -f invalid-wildcard-policy.yaml
    $ kubectl apply -f no-policy.yaml
    $ kubectl apply -f valid-policy.yaml
    ```

## More References: 
- https://msazure.club/podsecuritypolicy-explained-by-examples/
- https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#webhook-configuration
- https://github.com/kubernetes/kubernetes/blob/v1.13.0/test/images/webhook/main.go
- https://banzaicloud.com/blog/k8s-admission-webhooks/
- https://kubernetes.io/docs/concepts/policy/pod-security-policy/#seccomp
- https://github.com/kubernetes/client-go/blob/master/util/jsonpath/jsonpath_test.go



## Other ways of deploying
- As a serverless solution
  > https://github.com/kelseyhightower/denyenv-validating-admission-webhook
- Deploying webhook as an aggregated API Server
  > https://github.com/openshift/generic-admission-server