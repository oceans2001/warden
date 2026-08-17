package policy

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func BuildFQDNPolicy(sessionName string, allowedDomains []string) *unstructured.Unstructured {
	var fqdnRules []interface{}
	for _, domain := range allowedDomains {
		fqdnRules = append(fqdnRules, map[string]interface{}{
			"matchName": domain,
		})
	}

	policy := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cilium.io/v2",
			"kind":       "CiliumNetworkPolicy",
			"metadata": map[string]interface{}{
				"name": sessionName + "-egress-policy",
			},
			"spec": map[string]interface{}{
				"endpointSelector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"warden-session": sessionName,
					},
				},
				"egress": []interface{}{
					map[string]interface{}{
						"toFQDNs": fqdnRules,
						"toPorts": []interface{}{
							map[string]interface{}{
								"ports": []interface{}{
									map[string]interface{}{"port": "443", "protocol": "TCP"},
								},
							},
						},
					},
					map[string]interface{}{
						"toEndpoints": []interface{}{
							map[string]interface{}{
								"matchLabels": map[string]interface{}{
									"k8s:io.kubernetes.pod.namespace": "kube-system",
									"k8s-app":                         "kube-dns",
								},
							},
						},
						"toPorts": []interface{}{
							map[string]interface{}{
								"ports": []interface{}{
									map[string]interface{}{"port": "53", "protocol": "UDP"},
									map[string]interface{}{"port": "53", "protocol": "TCP"},
								},
								"rules": map[string]interface{}{
									"dns": []interface{}{
										map[string]interface{}{"matchPattern": "*"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return policy
}

func ApplyPolicy(dynamicClient dynamic.Interface, pol *unstructured.Unstructured) error {
	gvr := schema.GroupVersionResource{
		Group:    "cilium.io",
		Version:  "v2",
		Resource: "ciliumnetworkpolicies",
	}

	_, err := dynamicClient.Resource(gvr).Namespace("default").Create(context.TODO(), pol, metav1.CreateOptions{})
	return err
}

func DeletePolicy(dynamicClient dynamic.Interface, sessionName string) error {
	gvr := schema.GroupVersionResource{
		Group:    "cilium.io",
		Version:  "v2",
		Resource: "ciliumnetworkpolicies",
	}

	return dynamicClient.Resource(gvr).Namespace("default").Delete(context.TODO(), sessionName+"-egress-policy", metav1.DeleteOptions{})
}
