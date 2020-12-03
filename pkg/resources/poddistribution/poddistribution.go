package poddistribution

import (
	"context"
	"fmt"
	integreatlyv1alpha1 "github.com/integr8ly/integreatly-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/integreatly-operator/pkg/resources"
	appsv1 "github.com/openshift/api/apps/v1"
	"github.com/sirupsen/logrus"
	k8appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strconv"
	"strings"
	"time"
)

const (
	// ZoneLabel is the label that specifies the zone where a node is
	ZoneLabel = "topology.kubernetes.io/zone"
	// Annotation counter on the pods controller, dc, rs, ss
	PodRebalanceAttempts = "pod-rebalance-attempts"
	maxBalanceAttempts   = 3
	name                 = "name"
	ns                   = "ns"
	kind                 = "kind"
)

type KindNameSpaceName struct {
	*k8sTypes.NamespacedName
	Kind string
}

func (knn KindNameSpaceName) String() string {
	return fmt.Sprintf("%s/%s/%s", knn.Kind, knn.Namespace, knn.Name)
}

// Check the PodBalanceAttempts, ensure less than maxBalanceAttempts
func verifyRebalanceCount(ctx context.Context, client k8sclient.Client, obj metav1.Object) (bool, error) {
	ant := obj.GetAnnotations()
	// If there are no annotations then 0 balance attempts have been made
	if ant == nil {
		return true, nil
	}
	if val, ok := ant[PodRebalanceAttempts]; ok {
		i, err := strconv.Atoi(val)
		if err != nil {
			return false, fmt.Errorf("Error converting string annotations %s", ant[PodRebalanceAttempts])
		} else {
			if i >= maxBalanceAttempts {
				logrus.Warningf("Reached max balance attempts for %s on %s", obj.GetName(), obj.GetNamespace())
				return false, nil
			}
		}
	}

	return true, nil
}

func getObject(ctx context.Context, client k8sclient.Client, objDtls map[string]string) (metav1.Object, error) {
	if objDtls[kind] == "DeploymentConfig" {
		return getDeploymentConfig(ctx, client, objDtls)
	}
	if objDtls[kind] == "ReplicaSet" {
		return getReplicaSet(ctx, client, objDtls)
	}
	if objDtls[kind] == "StatefulSet" {
		return getStatefulSet(ctx, client, objDtls)
	}
	return nil, fmt.Errorf("Unexpected obj kind %s", objDtls[kind])
}

func getDeploymentConfig(ctx context.Context, client k8sclient.Client, objDtls map[string]string) (metav1.Object, error) {
	deploymentConfig := &appsv1.DeploymentConfig{}
	err := client.Get(ctx, k8sTypes.NamespacedName{Name: objDtls[name], Namespace: objDtls[ns]}, deploymentConfig)
	if err != nil {
		return nil, fmt.Errorf("Error getting DeploymentConfig: %w", err)
	}
	return deploymentConfig, nil
}

func getReplicaSet(ctx context.Context, client k8sclient.Client, objDtls map[string]string) (metav1.Object, error) {
	rs := &k8appsv1.ReplicaSet{}
	err := client.Get(ctx, k8sTypes.NamespacedName{Name: objDtls[name], Namespace: objDtls[ns]}, rs)
	if err != nil {
		return nil, fmt.Errorf("Error getting ReplicaSet: %w", err)
	}
	return rs, nil
}

func getStatefulSet(ctx context.Context, client k8sclient.Client, objDtls map[string]string) (metav1.Object, error) {
	ss := &k8appsv1.StatefulSet{}
	err := client.Get(ctx, k8sTypes.NamespacedName{Name: objDtls[name], Namespace: objDtls[ns]}, ss)
	if err != nil {
		logrus.Errorf("Error getting StatefulSet %v", err)
		return nil, err
	}
	return ss, nil
}

func parseKnn(knn string) (map[string]string, error) {
	objDtls := map[string]string{}
	s := strings.Split(knn, "/")
	if len(s) != 3 {
		return nil, fmt.Errorf("Error parsing knn :%s", knn)
	}
	objDtls[kind] = s[0]
	objDtls[ns] = s[1]
	objDtls[name] = s[2]
	return objDtls, nil
}

func getNamespaces(nsPrefix string, installType string) []string {
	namespaces := []string{
		nsPrefix + "3scale",
		nsPrefix + "rhsso",
		nsPrefix + "user-sso",
	}

	if installType == string(integreatlyv1alpha1.InstallationTypeManagedApi) {
		namespaces = append(namespaces, nsPrefix+"marin3r")
	}
	return namespaces
}

func ReconcilePodDistribution(ctx context.Context, client k8sclient.Client, nsPrefix string, installType string) *resources.MultiErr {
	var mErr = &resources.MultiErr{}

	isMultiAZCluster, err := resources.IsMultiAZCluster(ctx, client)
	if err != nil {
		mErr.Add(err)
		return mErr
	}
	if !isMultiAZCluster {
		return mErr
	}

	for _, ns := range getNamespaces(nsPrefix, installType) {
		logrus.Infof("Reconciling Pod Balance in ns %s", ns)
		unbalanced, err := findUnbalanced(ctx, ns, client)
		if err != nil {
			mErr.Add(fmt.Errorf("Error getting pods to balance on namespace %s. %w", ns, err))
			continue
		}
		for knn, pods := range unbalanced {
			objDtls, err := parseKnn(knn)
			if err != nil {
				mErr.Add(err)
				continue
			}
			obj, err := getObject(ctx, client, objDtls)
			if err != nil {
				mErr.Add(err)
				continue
			}
			rebalance, err := verifyRebalanceCount(ctx, client, obj)
			if err != nil {
				mErr.Add(err)
				continue
			}
			if rebalance {
				err := forceRebalance(ctx, client, objDtls, pods)
				if err != nil {
					mErr.Add(err)
				}
			}
		}
	}
	return mErr
}

func findUnbalanced(ctx context.Context, nameSpace string, client k8sclient.Client) (map[string][]string, error) {
	nodes := &corev1.NodeList{}
	unBalanced := map[string][]string{}
	if err := client.List(ctx,
		nodes, &k8sclient.ListOptions{}); err != nil {
		return unBalanced, err
	}
	nodesToZone := map[string]string{}
	for _, n := range nodes.Items {
		for _, a := range n.Status.Addresses {
			if a.Type == corev1.NodeInternalIP {
				nodesToZone[a.Address] = n.Labels[ZoneLabel]
				break
			}
		}
	}
	logrus.Debugf("nodes to zone %v", nodesToZone)
	balance := map[string][]string{}
	objPods := map[string][]string{}
	l := &corev1.PodList{}
	listOpts := []k8sclient.ListOption{
		k8sclient.InNamespace(nameSpace),
	}
	if err := client.List(ctx, l, listOpts...); err != nil {
		return unBalanced, fmt.Errorf("Error getting pod lists %w", err)
	}

	// need to check if there is more than 1 pod of a kind
	podCount := map[string]int{}
	logrus.Debugf("total pods in ns %s: %d", nameSpace, len(l.Items))
	for _, p := range l.Items {
		if p.Status.Phase != "Running" {
			continue
		}

		for _, o := range p.OwnerReferences {
			if o.Controller != nil && *o.Controller {
				knn := &KindNameSpaceName{
					NamespacedName: &k8sTypes.NamespacedName{
						Namespace: nameSpace,
					},
				}

				if o.Kind == "ReplicationController" {
					knn.Name = p.Annotations["openshift.io/deployment-config.name"]
					knn.Kind = "DeploymentConfig"
				} else {
					knn.Name = o.Name
					knn.Kind = o.Kind
				}
				if _, ok := podCount[knn.String()]; !ok {
					podCount[knn.String()] = 1
				} else {
					podCount[knn.String()] = podCount[knn.String()] + 1
				}
				if balance[knn.String()] == nil {
					balance[knn.String()] = []string{}
				}

				if objPods[knn.String()] == nil {
					objPods[knn.String()] = []string{}
				}
				objPods[knn.String()] = append(objPods[knn.String()], p.Name)

				zone := nodesToZone[p.Status.HostIP]
				if !zoneExists(balance[knn.String()], zone) {
					balance[knn.String()] = append(balance[knn.String()], nodesToZone[p.Status.HostIP])
				}
				break
			}
		}
	}
	for knn := range balance {
		if len(balance[knn]) == 1 && podCount[knn] > 1 {
			logrus.Warningf("Requires pod rebalance %s", knn)
			unBalanced[knn] = objPods[knn]
		}
	}

	return unBalanced, nil
}

func zoneExists(zones []string, zone string) bool {
	for _, z := range zones {
		if z == zone {
			return true
		}
	}
	return false
}

// Delete a single pod to force redistribution
func forceRebalance(ctx context.Context, client k8sclient.Client, objDtls map[string]string, pods []string) error {

	deletePod(ctx, client, pods[0], objDtls[ns])
	// The wait prevents version clash errors when updating the controller
	err := wait.Poll(time.Second*5, time.Second*5, func() (done bool, err error) {
		if objDtls[kind] == "DeploymentConfig" {
			err = updatePodBalanceAttemptsOnDeploymentConfig(objDtls, ctx, client)
		} else if objDtls[kind] == "ReplicaSet" {
			err = updatePodBalanceAttemptsOnReplicaSet(objDtls, ctx, client)
		} else if objDtls[kind] == "StatefulSet" {
			err = updatePodBalanceAttemptsOnStatefulSet(objDtls, ctx, client)
		} else {
			return true, fmt.Errorf("Unexpected object kind %s. Can not scale pods", objDtls[kind])
		}
		return true, err
	})
	if err != nil {
		return err
	}
	return nil
}

func deletePod(ctx context.Context, client k8sclient.Client, podName string, ns string) {
	logrus.Infof("Attempting to delete pod %s, on ns %s", podName, ns)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
		},
	}
	err := client.Get(ctx, k8sclient.ObjectKey{
		Name:      podName,
		Namespace: ns,
	}, pod)
	if err != nil {
		logrus.Errorf("Error getting pod %s on namespace %s. %v", podName, ns, err)
		return
	}
	if err := client.Delete(ctx, pod); err != nil {
		logrus.Errorf("Error deleting pod %s on namespace %s. %v", podName, ns, err)
	}
}

func updatePodBalanceAttemptsOnDeploymentConfig(objDtls map[string]string, ctx context.Context, client k8sclient.Client) error {
	dcName := objDtls[name]
	ns := objDtls[ns]
	logrus.Infof("Updating pod balance annotation %s, on ns %s", dcName, ns)

	dc := &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dcName,
			Namespace: ns,
		},
	}
	err := client.Get(ctx, k8sclient.ObjectKey{
		Name:      dcName,
		Namespace: ns,
	}, dc)
	if err != nil {
		return fmt.Errorf("Error getting dc %s on namespace %s. %w", ns, ns, err)
	}
	ant, err := getAnnotations(dc.GetAnnotations(), dcName, ns)
	if err != nil {
		return err
	}
	dc.SetAnnotations(ant)
	if err := client.Update(ctx, dc); err != nil {
		return fmt.Errorf("Error Updating dc %s on %s. %w", dcName, ns, err)
	}
	logrus.Infof("Successfully updated dc %s on %s", dcName, ns)
	return nil
}

func updatePodBalanceAttemptsOnReplicaSet(objDtls map[string]string, ctx context.Context, client k8sclient.Client) error {
	name := objDtls[name]
	ns := objDtls[ns]
	logrus.Infof("Updating pod balance annotation on replicaset %s, on ns %s", name, ns)

	rs := &k8appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
	err := client.Get(ctx, k8sclient.ObjectKey{
		Name:      name,
		Namespace: ns,
	}, rs)
	if err != nil {
		return fmt.Errorf("Error getting ReplicaSet %s on namespace %s. %w", ns, ns, err)
	}

	ant, err := getAnnotations(rs.GetAnnotations(), rs.GetName(), rs.GetNamespace())
	if err != nil {
		return err
	}
	rs.SetAnnotations(ant)

	if err := client.Update(ctx, rs); err != nil {
		return fmt.Errorf("Error Updating ReplicaSet %s on %s. %w", rs.Name, rs.Namespace, err)
	}
	logrus.Infof("Successfully updated ReplicaSet %s on %s", name, ns)
	return nil
}

func updatePodBalanceAttemptsOnStatefulSet(objDtls map[string]string, ctx context.Context, client k8sclient.Client) error {
	name := objDtls[name]
	ns := objDtls[ns]
	logrus.Infof("Updating pod balance ss annotation %s, on ns %s", name, ns)

	ss := &k8appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
	err := client.Get(ctx, k8sclient.ObjectKey{
		Name:      name,
		Namespace: ns,
	}, ss)
	if err != nil {
		return fmt.Errorf("Error getting ss %s on namespace %s. %w", ns, ns, err)
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, client, ss, func() error {
		ant, err := getAnnotations(ss.GetAnnotations(), ss.GetName(), ss.GetNamespace())
		if err != nil {
			return err
		}
		ss.SetAnnotations(ant)
		return nil
	}); err != nil {
		return fmt.Errorf("Error Updating ss %s on %s. %w", ss.Name, ss.Namespace, err)
	}
	logrus.Infof("Successfully updated statefulset %s on %s", name, ns)
	return nil
}

func getAnnotations(ant map[string]string, name string, ns string) (map[string]string, error) {
	if ant == nil {
		ant = map[string]string{}
	}
	if val, ok := ant[PodRebalanceAttempts]; ok {
		i, err := strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("Error converting annotation to int %s, %s, %s", PodRebalanceAttempts, name, ns)
		}
		i = i + 1
		ant[PodRebalanceAttempts] = strconv.Itoa(i)
	} else {
		ant[PodRebalanceAttempts] = "1"
	}

	return ant, nil
}
