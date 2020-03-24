package threescale

import (
	goctx "context"
	"crypto/tls"
	"net/http"
	"testing"

	"k8s.io/apimachinery/pkg/types"

	integreatlyv1alpha1 "github.com/integr8ly/integreatly-operator/pkg/apis/integreatly/v1alpha1"

	"github.com/integr8ly/integreatly-operator/pkg/products/threescale"

	"github.com/integr8ly/integreatly-operator/test/common"
)

func getThreeScaleClient(rhmi *integreatlyv1alpha1.RHMI) threescale.ThreeScaleInterface {
	httpc := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: rhmi.Spec.SelfSignedCerts},
		},
	}

	return threescale.NewThreeScaleClient(httpc, rhmi.Spec.RoutingSubdomain)
}

func TestAPICRUDL(t *testing.T, ctx *common.TestingContext) {
	rhmi := &integreatlyv1alpha1.RHMI{}

	// get the RHMI custom resource to check what storage type is being used
	err := ctx.Client.Get(goctx.TODO(), types.NamespacedName{Name: common.InstallationName, Namespace: common.RHMIOperatorNamespace}, rhmi)
	if err != nil {
		t.Fatal(err)
	}

	tsClient := getThreeScaleClient(rhmi)

}
