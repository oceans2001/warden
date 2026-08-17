package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreatePod(clientset *kubernetes.Clientset, name string, image string, command []string) error {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app":            "warden-agent",
				"warden-session": name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    name,
					Image:   image,
					Command: command,
				},
			},
		},
	}

	_, err := clientset.CoreV1().Pods("default").Create(context.TODO(), pod, metav1.CreateOptions{})
	return err
}

func ListAgents(clientset *kubernetes.Clientset) ([]string, error) {
	pods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=warden-agent",
	})
	if err != nil {
		return nil, err
	}

	var names []string
	for _, pod := range pods.Items {
		names = append(names, pod.Name)
	}

	return names, nil
}

func DeletePod(clientset *kubernetes.Clientset, name string) error {
	return clientset.CoreV1().Pods("default").Delete(context.TODO(), name, metav1.DeleteOptions{})
}
