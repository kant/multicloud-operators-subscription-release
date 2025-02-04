/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package helmrelease

import (
	"github.com/IBM/multicloud-operators-subscription-release/pkg/apis"
	// "strconv"
	"context"
	"strings"

	// "archive/tar"
	// "bytes"
	// "compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	// "net/http"

	"github.com/IBM/multicloud-operators-subscription-release/pkg/apis/app/v1alpha1"
	"github.com/pborman/uuid"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	//"sigs.k8s.io/controller-runtime/pkg/client"
	//"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/martinlindhe/base36"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	//	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

//This test is not working because the helm-operator needs a config.Config and not a client.
//So the fake client is not passed along.
func TestHelmReleaseReconcileCreate(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	tempDir, _ := ioutil.TempDir("/tmp", "charts")
	var (
		name        = "multicloud-operators-subscription-release"
		namespace   = "default"
		releaseName = "nginx-ingress"
		chartName   = "nginx-ingress"
		chartsDir   = filepath.Join(tempDir, "test")
	)
	sr := &v1alpha1.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "app.ibm.com/v1alpha1",
			Kind:       "HelmRelease",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID("89e6052a-d566-11e9-b55f-fa163e0cb658"),
		},
		Spec: v1alpha1.HelmReleaseSpec{
			Source: &v1alpha1.Source{
				SourceType: v1alpha1.HelmRepoSourceType,
				HelmRepo: &v1alpha1.HelmRepo{
					Urls: []string{"https://helm.nginx.com/stable/nginx-ingress-0.3.5.tgz"},
				},
			},
			ChartName:   chartName,
			ReleaseName: releaseName,
			//			Version:     "",
		},
	}

	// Check if deployment has been created and has the correct size.
	dep := &appsv1.Deployment{}
	t.Log("sr.GetUID() " + sr.GetUID())
	shorthenUID := shortenUID(sr.GetUID())
	t.Log("shorthenUID " + shorthenUID)
	deploymentName := releaseName[0:4] + "-" + shorthenUID + "-" + chartName
	namespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      deploymentName,
	}

	os.Setenv("CHARTS_DIR", chartsDir)
	// Objects to track in the fake client.
	objs := []runtime.Object{sr}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, sr)
	apis.AddToScheme(scheme.Scheme)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	// cfg, err := config.GetConfig()
	// if err != nil {
	// 	t.Error(err.Error())
	// }
	// cl, err := client.New(cfg, client.Options{
	// 	Scheme: s,
	// })
	// if err != nil {
	// 	t.Error(err.Error())
	// }
	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileHelmRelease{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
	err = r.client.Get(context.TODO(), namespacedName, dep)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}
}

// func getReleaseName(cr *unstructured.Unstructured) string {
// 	return fmt.Sprintf("%s-%s", cr.GetName(), shortenUID(cr.GetUID()))
// }

func shortenUID(uid types.UID) string {
	u := uuid.Parse(string(uid))
	uidBytes, err := u.MarshalBinary()
	if err != nil {
		return strings.Replace(string(uid), "-", "", -1)
	}
	return strings.ToLower(base36.EncodeBytes(uidBytes))
}
