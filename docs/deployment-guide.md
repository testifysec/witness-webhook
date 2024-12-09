# Deploying witness-webhook

witness-webhook includes a Helm chart to help deploy the service. At the core of it, witness-webhook is a simple service that listens for configured webhook events, creates an in-toto attestation describing the captured events, signs the attestation with a configured signer, and then stores the attestation either to a configured directory or to a configured Archivista instance.

witness-webhook is configured via some environment variables for basic service configuration, and a yaml file that configures the different webhooks the service will listen for. witness-webhook will create an endpoint for each configured webhook in this yaml file at the path `/webhook/{webhook-name}`. Each configured webhook has it's own signer configured to support the case where webhooks from different sources may need to be signed differently. For more details on the specific configuration variables and this yaml file, see the README in the root directory.

## Github and Vault

This example will walk through a sample deployment of witness-webhook that listens for events from Github and signs attestations with a Vault transit key with Kubernetes authentication to Vault. This example will use a local minikube instance to deploy witness-webhook into and a local Vault development instance. This example will assume you have the npm, [minikube](https://minikube.sigs.k8s.io/docs/start), [Vault](https://www.markdownguide.org/basic-syntax/), [Helm](https://helm.sh/docs/intro/install), and [Kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)  CLI tools installed.

### Setting up minikube

Once minikube is installed, simply running `minikube start` will be enough to bring up a minikube instance sufficient for this example. Your kubecontext should automatically be switched to point to the minikube instance, but just in case or if you switch your kubecontext, running `minikube update-context` will set it back.

Any files referenced in the following steps are available in the docs/example-vault directory, and all commands assume you are running in the docs/example-vault directory.

### Setting up Vault

1. Start the Vault server with `vault server -dev -dev-root-token-id root -dev-listen-address 0.0.0.0:8200 -log-level=debug`
1. Set VAULT_ADDR to point to the local dev server with `export VAULT_ADDR=http://127.0.0.1:8200`
1. Enable kubernetes authentication `vault auth enable kubernetes`
1. Enable the transit secrets engine `vault secrets enable transit`
1. Create a transit key with `vault write transit/keys/testkey type=ecdsa-p521`
1. Create a vault service account for the Token Reviewer: `kubectl create serviceaccount vault-token-reviewer`
1. Apply the cluster role for Token Reivew `kubectl apply -f cluterrole.yaml`
1. Create a token for the Token Reviewer service account `export REVIEWER_TOKEN=$(kubectl create token vault-token-reviewer)`
1. Get the Kubernetes CA `export KUBE_CA=$(kubectl config view --raw --minify --flatten --output='jsonpath={.clusters[].cluster.certificate-authority-data}' | base64 --decode)`
1. Get the Kubernetes host `` export KUBE_HOST=`$(kubectl config view --raw --minify --flatten --output='jsonpath={.clusters[].cluster.server}')` ``
1. Configure kubernetes auth for Vault `vault write auth/kubernetes/config token_reviewer_jwt="${REVIEWER_TOKEN}" kubernetes_host="${KUBE_HOST}" kubernetes_ca_cert="${KUBE_CA}" issuer="https://kubernetes.default.svc.cluster.local"`
1. Apply a Vault policy allowing access to the transit secret engine `vault write sys/policy/witness-webhook policy=@vaultpolicy.hcl`
1. Create a role that will bind to witness-webhook's service account `vault write auth/kubernetes/role/witness-webhook bound_service_account_names="witness-webhook" bound_service_account_namespaces="default" policies="witness-webhook" ttl=5m`

### Deploying witness-webhook

1. Create a service account for witness-webhook `kubectl create service account witness-webhook`
1. Create a secret for the webhook, `echo -n supersecretkey > githubsecret` and `kubectl create secret generic githubwebhooksecret --from-file githubsecret`
1. SSH into minikube to get the IP we'll talk to Vault with `minikube ssh dig +short host.docker.internal`
1. Use the IP address above to modify the values.yaml in the example's directory, updating the `hashivault-addr` variable's value. While here, observe the values.yaml. A few things of note: we are mounting the previously created secret into the pod. This mounted secret is then referred to in the witness-webhook's yaml configuration. We are setting a few global Vault configuration variables with the VAULT_ environment variables.
1. Install the Helm chart `helm install witness-webhook ../../chart -f values.yaml`
1. Port forward to the pod:

```
1. Get the application URL by running these commands:
  export POD_NAME=$(kubectl get pods --namespace default -l "app.kubernetes.io/name=witness-webhook,app.kubernetes.io/instance=witness-webhook" -o jsonpath="{.items[0].metadata.name}")
  export CONTAINER_PORT=$(kubectl get pod --namespace default $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
  echo "Visit http://127.0.0.1:8080 to use your application"
  kubectl --namespace default port-forward $POD_NAME 8080:$CONTAINER_PORT
```

1. Use smee.io to setup forwarding to the locally running witness-webhook `npm install -g smee-client` and `smee -t http://localhost:8080/webhook/github` . Make note of the URL smee provides.

### Setting up webhook events from a Github repository

Navigate to a repository of your choice, and enter the Settings tag. From here find the Webhooks section. Add a new webhook.

For the Payload URL, use the URL smee provided to you. It should look similar to <https://smee.io/tjywPLhwk6tJyLus> .

Select application/json as the Content type

For the Secret, insert `supersecretkey`, matching the secret we created earlier.

Select which events you would like to receive.

Once this is setup, try to trigger one of the events.

You can follow the witness-webhook logs with `kubectl logs -f $POD_NAME` and should see something similar to the following if all is successful

```
2024/12/09 23:08:39 attestation stored in archivista for webhook github with gitoid d3b05a8d9b2771b082f845d418936b144adcc8d38b5f5cb10a66e063f6090d5c
```
