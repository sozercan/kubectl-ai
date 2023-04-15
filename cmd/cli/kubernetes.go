package cli

import (
	"bytes"
	"context"
	"path/filepath"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const defaultNamespace = "default"

func applyManifest(completion string) error {
	var kubeConfig string
	if *kubernetesConfigFlags.KubeConfig == "" {
		kubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	} else {
		kubeConfig = *kubernetesConfigFlags.KubeConfig
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return err
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	dd, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	var namespace string
	if *kubernetesConfigFlags.Namespace == "" {
		clientConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfig},
			&clientcmd.ConfigOverrides{
				CurrentContext: "",
			}).RawConfig()
		if err != nil {
			return err
		}
		if clientConfig.Contexts[clientConfig.CurrentContext].Namespace == "" {
			namespace = defaultNamespace
		} else {
			namespace = clientConfig.Contexts[clientConfig.CurrentContext].Namespace
		}
	} else {
		namespace = *kubernetesConfigFlags.Namespace
	}

	manifest := []byte(completion)
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 100)
	for {
		var rawObj runtime.RawExtension
		if err = decoder.Decode(&rawObj); err != nil {
			break
		}

		obj, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			return err
		}
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return err
		}

		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

		gr, err := restmapper.GetAPIGroupResources(c.Discovery())
		if err != nil {
			return err
		}

		mapper := restmapper.NewDiscoveryRESTMapper(gr)
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}

		var dri dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			if unstructuredObj.GetNamespace() == "" {
				unstructuredObj.SetNamespace(namespace)
			}
			dri = dd.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
		} else {
			dri = dd.Resource(mapping.Resource)
		}

		if _, err := dri.Apply(context.Background(), unstructuredObj.GetName(), unstructuredObj, metav1.ApplyOptions{FieldManager: "application/apply-patch"}); err != nil {
			return err
		}
	}
	return nil
}
