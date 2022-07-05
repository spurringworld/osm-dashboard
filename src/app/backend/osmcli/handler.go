package osmcli

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	_ "embed" // required to embed resources

	"github.com/pkg/errors"
	"golang.org/x/net/xsrftoken"

	cli "github.com/openservicemesh/osm/pkg/cli"

	helm "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/openservicemesh/osm/pkg/constants"

	restful "github.com/emicklei/go-restful/v3"
	"github.com/kubernetes/dashboard/src/app/backend/api"
	clientapi "github.com/kubernetes/dashboard/src/app/backend/client/api"
	backenderrors "github.com/kubernetes/dashboard/src/app/backend/errors"
)

const ()

var settings = cli.New()

// chartTGZSource is the `helm package`d representation of the default Helm chart.
// Its value is embedded at build time.
//go:embed chart.tgz
var chartTGZSource []byte

type OsmCliHandler struct {
	clientManager clientapi.ClientManager
}

func (self OsmCliHandler) Install(ws *restful.WebService) {
	ws.Route(
		ws.POST("/osm/cmd/cli/install").
			To(self.handleOsmInstall).
			Reads(OsmInstallSpec{}).
			Writes(api.CsrfToken{}))

	ws.Route(
		ws.POST("/osm/cmd/cli/uninstall").
			To(self.handleOsmUninstall).
			Reads(OsmUninstallSpec{}).
			Writes(api.CsrfToken{}))
}

func debug(format string, v ...interface{}) {
	if settings.Verbose() {
		format = fmt.Sprintf("[debug] %s\n", format)
		fmt.Printf(format, v...)
	}
}

func (self OsmCliHandler) handleOsmInstall(request *restful.Request, response *restful.Response) {
	osmInstallSpec := NewOsmInstallSpec()
	if err := request.ReadEntity(&osmInstallSpec); err != nil {
		backenderrors.HandleInternalError(response, err)
		return
	}

	actionConfig := new(helm.Configuration)
	_ = actionConfig.Init(settings.RESTClientGetter(), osmInstallSpec.Namespace, "secret", debug)

	installClient := helm.NewInstall(actionConfig)

	installClient.ReleaseName = osmInstallSpec.MeshName
	installClient.Namespace = osmInstallSpec.Namespace
	installClient.CreateNamespace = true
	installClient.Wait = false
	installClient.Atomic = false
	installClient.Timeout = 5 * time.Minute

	values := map[string]interface{}{}
	values["deployGrafana"] = true
	values["deployJaeger"] = true
	values["deployPrometheus"] = true
	values["enablePermissiveTrafficPolicy"] = true
	values["enforceSingleMesh"] = osmInstallSpec.EnforceSingleMesh
	values["meshName"] = osmInstallSpec.MeshName
	values["osmNamespace"] = osmInstallSpec.Namespace

	chartRequested, err := loader.LoadArchive(bytes.NewReader(chartTGZSource))
	if err != nil {
		backenderrors.HandleInternalError(response, err)
		return
	}

	k8sClient, err := self.clientManager.Client(request)
	if err != nil {
		backenderrors.HandleInternalError(response, err)
		return
	}

	if _, err = installClient.Run(chartRequested, values); err != nil {
		if !settings.Verbose() {
			backenderrors.HandleInternalError(response, err)
			return
		}
		pods, _ := k8sClient.CoreV1().Pods(settings.Namespace()).List(context.Background(), metav1.ListOptions{})

		for _, pod := range pods.Items {
			fmt.Printf("Status for pod %s in namespace %s:\n %v\n\n", pod.Name, pod.Namespace, pod.Status)
		}
	}

	fmt.Printf("OSM installed successfully in namespace [%s] with mesh name [%s]\n", settings.Namespace(), osmInstallSpec.MeshName)

	// TODO
	action := request.PathParameter("action")
	token := xsrftoken.Generate(self.clientManager.CSRFKey(), "none", action)
	response.WriteHeaderAndEntity(http.StatusOK, api.CsrfToken{Token: token})
}

func (self OsmCliHandler) handleOsmUninstall(request *restful.Request, response *restful.Response) {
	osmUninstallSpec := NewOsmUninstallSpec()
	if err := request.ReadEntity(&osmUninstallSpec); err != nil {
		backenderrors.HandleInternalError(response, err)
		return
	}

	actionConfig := new(helm.Configuration)
	_ = actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), "secret", debug)

	helmUninstallClient := helm.NewUninstall(actionConfig)

	deleteClusterWideResources := true

	_, err := helmUninstallClient.Run("osm")
	if err != nil {
		println("error")
	}

	if deleteClusterWideResources {
		var failedDeletions []string

		err := self.uninstallCustomResourceDefinitions(request, response)
		if err != nil {
			failedDeletions = append(failedDeletions, "CustomResourceDefinitions")
		}

		err = self.uninstallMutatingWebhookConfigurations(request, response)
		if err != nil {
			failedDeletions = append(failedDeletions, "MutatingWebhookConfigurations")
		}

		err = self.uninstallValidatingWebhookConfigurations(request, response)
		if err != nil {
			failedDeletions = append(failedDeletions, "ValidatingWebhookConfigurations")
		}

		err = self.uninstallSecrets(request, response)
		if err != nil {
			failedDeletions = append(failedDeletions, "Secrets")
		}

		if len(failedDeletions) != 0 {
			fmt.Printf("Failed to completely delete the following OSM resource types: %+v", failedDeletions)
		}
	}

	// TODO
	action := request.PathParameter("action")
	token := xsrftoken.Generate(self.clientManager.CSRFKey(), "none", action)
	response.WriteHeaderAndEntity(http.StatusOK, api.CsrfToken{Token: token})
}

// uninstallCustomResourceDefinitions uninstalls osm and smi-related crds from the cluster.
func (self OsmCliHandler) uninstallCustomResourceDefinitions(request *restful.Request, response *restful.Response) error {
	crds := []string{
		"egresses.policy.openservicemesh.io",
		"ingressbackends.policy.openservicemesh.io",
		"meshconfigs.config.openservicemesh.io",
		"upstreamtrafficsettings.policy.openservicemesh.io",
		"retries.policy.openservicemesh.io",
		"multiclusterservices.config.openservicemesh.io",
		"httproutegroups.specs.smi-spec.io",
		"tcproutes.specs.smi-spec.io",
		"trafficsplits.split.smi-spec.io",
		"traffictargets.access.smi-spec.io",
	}

	extensionsClient, err := self.clientManager.APIExtensionsClient(request)
	if err != nil {
		backenderrors.HandleInternalError(response, err)
		return nil
	}

	var failedDeletions []string
	for _, crd := range crds {
		err := extensionsClient.ApiextensionsV1().CustomResourceDefinitions().Delete(context.Background(), crd, metav1.DeleteOptions{})

		if err == nil {
			fmt.Printf("Successfully deleted OSM CRD: %s\n", crd)
			continue
		}

		if k8sapierrors.IsNotFound(err) {
			fmt.Printf("Ignoring - did not find OSM CRD: %s\n", crd)
		} else {
			fmt.Printf("Failed to delete OSM CRD %s: %s\n", crd, err.Error())
			failedDeletions = append(failedDeletions, crd)
		}
	}

	if len(failedDeletions) != 0 {
		return errors.Errorf("Failed to delete the following OSM CRDs: %+v", failedDeletions)
	}

	return nil
}

// uninstallMutatingWebhookConfigurations uninstalls osm-related mutating webhook configurations from the cluster.
func (self OsmCliHandler) uninstallMutatingWebhookConfigurations(request *restful.Request, response *restful.Response) error {
	// These label selectors should always match the Helm post-delete hook at charts/osm/templates/cleanup-hook.yaml.
	// TODO
	webhookConfigurationsLabelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			constants.OSMAppNameLabelKey:     constants.OSMAppNameLabelValue,
			constants.OSMAppInstanceLabelKey: "osm",
			constants.AppLabel:               constants.OSMInjectorName,
		},
	}

	webhookConfigurationsListOptions := metav1.ListOptions{
		LabelSelector: labels.Set(webhookConfigurationsLabelSelector.MatchLabels).String(),
	}

	k8sClient, err := self.clientManager.Client(request)
	if err != nil {
		backenderrors.HandleInternalError(response, err)
		return nil
	}

	mutatingWebhookConfigurations, err := k8sClient.AdmissionregistrationV1().MutatingWebhookConfigurations().List(context.Background(), webhookConfigurationsListOptions)

	if err != nil {
		errMsg := fmt.Sprintf("Failed to list OSM MutatingWebhookConfigurations in the cluster: %s", err.Error())
		fmt.Println(errMsg)
		return errors.New(errMsg)
	}

	if len(mutatingWebhookConfigurations.Items) == 0 {
		fmt.Print("Ignoring - did not find any OSM MutatingWebhookConfigurations in the cluster. Use --mesh-name to delete MutatingWebhookConfigurations belonging to a specific mesh if desired\n")
		return nil
	}

	var failedDeletions []string
	for _, mutatingWebhookConfiguration := range mutatingWebhookConfigurations.Items {
		err := k8sClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(context.Background(), mutatingWebhookConfiguration.Name, metav1.DeleteOptions{})

		if err == nil {
			fmt.Printf("Successfully deleted OSM MutatingWebhookConfiguration: %s\n", mutatingWebhookConfiguration.Name)
		} else {
			fmt.Printf("Found but failed to delete OSM MutatingWebhookConfiguration %s: %s\n", mutatingWebhookConfiguration.Name, err.Error())
			failedDeletions = append(failedDeletions, mutatingWebhookConfiguration.Name)
		}
	}

	if len(failedDeletions) != 0 {
		return errors.Errorf("Found but failed to delete the following OSM MutatingWebhookConfigurations: %+v", failedDeletions)
	}

	return nil
}

// uninstallValidatingWebhookConfigurations uninstalls osm-related validating webhook configurations from the cluster.
func (self OsmCliHandler) uninstallValidatingWebhookConfigurations(request *restful.Request, response *restful.Response) error {
	// These label selectors should always match the Helm post-delete hook at charts/osm/templates/cleanup-hook.yaml.
	// TODO
	webhookConfigurationsLabelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			constants.OSMAppNameLabelKey:     constants.OSMAppNameLabelValue,
			constants.OSMAppInstanceLabelKey: "osm",
			constants.AppLabel:               constants.OSMControllerName,
		},
	}

	webhookConfigurationsListOptions := metav1.ListOptions{
		LabelSelector: labels.Set(webhookConfigurationsLabelSelector.MatchLabels).String(),
	}

	k8sClient, err := self.clientManager.Client(request)
	if err != nil {
		backenderrors.HandleInternalError(response, err)
		return nil
	}

	validatingWebhookConfigurations, err := k8sClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(context.Background(), webhookConfigurationsListOptions)

	if err != nil {
		errMsg := fmt.Sprintf("Failed to list OSM ValidatingWebhookConfigurations in the cluster: %s", err.Error())
		fmt.Println(errMsg)
		return errors.New(errMsg)
	}

	if len(validatingWebhookConfigurations.Items) == 0 {
		fmt.Print("Ignoring - did not find any OSM ValidatingWebhookConfigurations in the cluster. Use --mesh-name to delete ValidatingWebhookConfigurations belonging to a specific mesh if desired\n")
		return nil
	}

	var failedDeletions []string
	for _, validatingWebhookConfiguration := range validatingWebhookConfigurations.Items {
		err := k8sClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Delete(context.Background(), validatingWebhookConfiguration.Name, metav1.DeleteOptions{})

		if err == nil {
			fmt.Printf("Successfully deleted OSM ValidatingWebhookConfiguration: %s\n", validatingWebhookConfiguration.Name)
			continue
		} else {
			fmt.Printf("Found but failed to delete OSM ValidatingWebhookConfiguration %s: %s\n", validatingWebhookConfiguration.Name, err.Error())
			failedDeletions = append(failedDeletions, validatingWebhookConfiguration.Name)
		}
	}

	if len(failedDeletions) != 0 {
		return errors.Errorf("Found but failed to delete the following OSM ValidatingWebhookConfigurations: %+v", failedDeletions)
	}

	return nil
}

// uninstallSecrets uninstalls osm-related secrets from the cluster.
func (self OsmCliHandler) uninstallSecrets(request *restful.Request, response *restful.Response) error {
	// TODO
	secrets := []string{
		"osm-ca-bundle",
	}

	k8sClient, err := self.clientManager.Client(request)
	if err != nil {
		backenderrors.HandleInternalError(response, err)
		return nil
	}

	var failedDeletions []string
	for _, secret := range secrets {
		err := k8sClient.CoreV1().Secrets("osm-system").Delete(context.Background(), secret, metav1.DeleteOptions{})

		if err == nil {
			fmt.Printf("Successfully deleted OSM secret %s in namespace %s\n", secret, "osm-system")
			continue
		}

		if k8sapierrors.IsNotFound(err) {
			if secret == "osm-ca-bundle" {
				fmt.Printf("Ignoring - did not find OSM CA bundle secret %s in namespace %s. Use --ca-bundle-secret-name and --osm-namespace to delete a specific mesh namespace's CA bundle secret if desired\n", secret, "osm-system")
			} else {
				fmt.Printf("Ignoring - did not find OSM secret %s in namespace %s. Use --osm-namespace to delete a specific mesh namespace's secret if desired\n", secret, "osm-system")
			}
		} else {
			fmt.Printf("Found but failed to delete the OSM secret %s in namespace %s: %s\n", secret, "osm-system", err.Error())
			failedDeletions = append(failedDeletions, secret)
		}
	}

	if len(failedDeletions) != 0 {
		return errors.Errorf("Found but failed to delete the following OSM secrets in namespace %s: %+v", "osm-system", failedDeletions)
	}

	return nil
}

func NewOsmCliHandler(clientManager clientapi.ClientManager) OsmCliHandler {
	return OsmCliHandler{clientManager: clientManager}
}
