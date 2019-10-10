package mobilesecurityservice

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/integr8ly/integreatly-operator/pkg/apis/integreatly/v1alpha1"
	mobilesecurityservice "github.com/integr8ly/integreatly-operator/pkg/apis/mobilesecurityservice/v1alpha1"
	"github.com/integr8ly/integreatly-operator/pkg/controller/installation/marketplace"
	"github.com/integr8ly/integreatly-operator/pkg/controller/installation/products/config"
	"github.com/integr8ly/integreatly-operator/pkg/resources"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/ownerutil"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	defaultInstallationNamespace = "mobile-security-service"
	defaultSubscriptionName      = "integreatly-mobile-security-service"
	serverClusterName            = "mobile-security-service"
	dbClusterName                = "mobile-security-service-db"
)

type Reconciler struct {
	Config        *config.MobileSecurityService
	ConfigManager config.ConfigReadWriter
	mpm           marketplace.MarketplaceInterface
	logger        *logrus.Entry
	*resources.Reconciler
}

func NewReconciler(configManager config.ConfigReadWriter, instance *v1alpha1.Installation, mpm marketplace.MarketplaceInterface) (*Reconciler, error) {
	config, err := configManager.ReadMobileSecurityService()
	if err != nil {
		return nil, errors.Wrap(err, "could not read mobile security service config")
	}

	if config.GetNamespace() == "" {
		config.SetNamespace(instance.Spec.NamespacePrefix + defaultInstallationNamespace)
	}

	logger := logrus.NewEntry(logrus.StandardLogger())

	return &Reconciler{
		ConfigManager: configManager,
		Config:        config,
		mpm:           mpm,
		logger:        logger,
		Reconciler:    resources.NewReconciler(mpm),
	}, nil
}

func (r *Reconciler) GetPreflightObject(ns string) runtime.Object {
	return nil
}

// Reconcile reads that state of the cluster for mobile security service and makes changes based on the state read
// and what is required
func (r *Reconciler) Reconcile(ctx context.Context, inst *v1alpha1.Installation, product *v1alpha1.InstallationProductStatus, serverClient pkgclient.Client) (v1alpha1.StatusPhase, error) {
	ns := r.Config.GetNamespace()
	if ns == "" {
		return v1alpha1.PhaseFailed, errors.New("namespace: value blank")
	}
	version, err := resources.NewVersion(v1alpha1.OperatorVersionMobileSecurityService)

	phase, err := r.ReconcileNamespace(ctx, ns, inst, serverClient)
	if err != nil || phase != v1alpha1.PhaseCompleted {
		return phase, err
	}

	phase, err = r.ReconcileSubscription(ctx, inst, marketplace.Target{Namespace: ns, Channel: marketplace.IntegreatlyChannel, Pkg: defaultSubscriptionName}, serverClient, version)
	if err != nil || phase != v1alpha1.PhaseCompleted {
		return phase, err
	}

	phase, err = r.reconcileComponents(ctx, serverClient, inst)
	if err != nil || phase != v1alpha1.PhaseCompleted {
		return phase, err
	}

	phase, err = r.handleProgressPhase(ctx, serverClient)
	if err != nil || phase != v1alpha1.PhaseCompleted {
		return phase, err
	}

	product.Host = r.Config.GetHost()
	product.Version = r.Config.GetProductVersion()

	r.logger.Infof("%s has reconciled successfully", r.Config.GetProductName())
	return v1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) reconcileComponents(ctx context.Context, client pkgclient.Client, inst *v1alpha1.Installation) (v1alpha1.StatusPhase, error) {

	r.logger.Debug("reconciling mobile security service db custom resource")

	mssDb := &mobilesecurityservice.MobileSecurityServiceDB{
		TypeMeta: metav1.TypeMeta{
			Kind: "MobileSecurityServiceDB",
			APIVersion: fmt.Sprintf(
				"%s/%s",
				mobilesecurityservice.SchemeGroupVersion.Group,
				mobilesecurityservice.SchemeGroupVersion.Version),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbClusterName,
			Namespace: r.Config.GetNamespace(),
		},
		Spec: mobilesecurityservice.MobileSecurityServiceDBSpec{
			ContainerName:          "database",
			DatabaseMemoryLimit:    "512Mi",
			DatabaseMemoryRequest:  "512Mi",
			DatabaseName:           "mobile_security_service",
			DatabaseNameParam:      "POSTGRESQL_DATABASE",
			DatabasePassword:       "postgres",
			DatabasePasswordParam:  "POSTGRESQL_PASSWORD",
			DatabasePort:           5432,
			DatabaseStorageRequest: "1Gi",
			DatabaseUser:           "postgresql",
			DatabaseUserParam:      "POSTGRESQL_USER",
			Image:                  "centos/postgresql-96-centos7",
			Size:                   1,
		},
	}
	ownerutil.EnsureOwner(mssDb, inst)

	// attempt to create the mss db custom resource
	if err := client.Create(ctx, mssDb); err != nil && !k8serr.IsAlreadyExists(err) {
		return v1alpha1.PhaseFailed, errors.Wrap(err, "failed to get or create a mobile security service db custom resource")
	}

	r.logger.Debug("reconciling mobile security service custom resource")

	mss := &mobilesecurityservice.MobileSecurityService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MobileSecurityService",
			APIVersion: mobilesecurityservice.SchemeGroupVersion.String(),
			//APIVersion: fmt.Sprintf(
			//	"%s/%s",
			//	&mobilesecurityservice.SchemeGroupVersion.Group,
			//	&mobilesecurityservice.SchemeGroupVersion.Version),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serverClusterName,
			Namespace: r.Config.GetNamespace(),
		},
		Spec: mobilesecurityservice.MobileSecurityServiceSpec{
			AccessControlAllowCredentials: "false",
			AccessControlAllowOrigin:      "*",
			ClusterProtocol:               "https",
			ConfigMapName:                 "mobile-security-service-config",
			ContainerName:                 "application",
			DatabaseHost:                  "mobile-security-service-db",
			DatabaseName:                  "mobile_security_service",
			DatabasePassword:              "postgres",
			DatabaseUser:                  "postgresql",
			Image:                         "quay.io/aerogear/mobile-security-service:0.2.2",
			LogFormat:                     "json",
			LogLevel:                      "info",
			MemoryLimit:                   "128Mi",
			MemoryRequest:                 "64Mi",
			OAuthContainerName:            "oauth-proxy",
			OAuthImage:                    "quay.io/openshift/origin-oauth-proxy:4.2.0",
			OAuthMemoryLimit:              "64Mi",
			OAuthMemoryRequest:            "32Mi",
			OAuthResourceCpu:              "10m",
			OAuthResourceCpuLimit:         "20m",
			Port:                          3000,
			ResourceCpu:                   "10m",
			ResourceCpuLimit:              "20m",
			RouteName:                     "route",
			Size:                          1,
		},
	}
	ownerutil.EnsureOwner(mss, inst)

	// attempt to create the mss custom resource
	if err := client.Create(ctx, mss); err != nil && !k8serr.IsAlreadyExists(err) {
		return v1alpha1.PhaseFailed, errors.Wrap(err, "failed to get or create a mobile security service custom resource")
	}

	// if there are no errors, the phase is complete
	return v1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) handleProgressPhase(ctx context.Context, client pkgclient.Client) (v1alpha1.StatusPhase, error) {
	r.logger.Debug("checking mobile security service pods are running")

	pods := &v1.PodList{}
	err := client.List(ctx, &pkgclient.ListOptions{Namespace: r.Config.GetNamespace()}, pods)
	if err != nil {
		return v1alpha1.PhaseFailed, errors.Wrap(err, "failed to check mobile security service installation")
	}

	//3 pods expected - operator, db, service
	if len(pods.Items) < 3 { //TODO check if status value to check
		return v1alpha1.PhaseInProgress, nil
	}

	//and they should all be ready
checkPodStatus:
	for _, pod := range pods.Items {
		for _, cnd := range pod.Status.Conditions {
			if cnd.Type == v1.ContainersReady {
				if cnd.Status != v1.ConditionStatus("True") {
					return v1alpha1.PhaseInProgress, nil
				}
				break checkPodStatus
			}
		}
	}

	r.logger.Infof("all pods ready, returning complete")
	return v1alpha1.PhaseCompleted, nil
}
