package handler

import (
	"context"
	"cpu-mem-scheduler/prometehus"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/changqings/k8scrd/client"
)

var (
	SchedulerName = "cpu-mem-scheduler"
)

type Scheduler struct {
	Ctx           context.Context
	Client        *kubernetes.Clientset
	PrometheusUrl string
}

func NewScheduler(url string) *Scheduler {

	return &Scheduler{
		Ctx:           context.TODO(),
		Client:        client.GetClient(),
		PrometheusUrl: url,
	}
}

func (s *Scheduler) Run() error {

	watcher, err := s.Client.CoreV1().Pods(corev1.NamespaceAll).Watch(s.Ctx, metav1.ListOptions{Watch: true})
	if err != nil {
		return err
	}

	for event := range watcher.ResultChan() {
		if event.Type != "ADDED" {
			continue
		}
		pod, ok := event.Object.(*corev1.Pod)
		if !ok {
			slog.Error("unexpected type", "object", event.Object)
			continue
		}
		if pod.Status.Phase == corev1.PodPending && pod.Spec.SchedulerName == SchedulerName {
			slog.Info("get node name", "pod", pod.Name)
			nodeName, err := getNodeName(s.Ctx, s.PrometheusUrl)
			if err != nil {
				slog.Error("get node name error", "error", err)
				continue
			}

			slog.Info("bind node name", "pod", pod.Name, "node", nodeName)
			err = bindPod(s.Client, pod, nodeName)
			if err != nil {
				slog.Error("bind pod error", "error", err)
				continue
			}
		}
	}
	return nil
}

func getNodeName(ctx context.Context, url string) (string, error) {

	var nodeName string
	client, err := prometehus.GetP8sClient(url)
	if err != nil {
		return nodeName, err
	}

	metrics, err := prometehus.GetNodeMetrics(ctx, client)
	if err != nil {
		return nodeName, err
	}
	// cpu and memory usage must smaller than 2
	var littleValue float64 = 2
	for n, v := range metrics {
		sum := v[0] + v[1]
		if sum < littleValue {
			nodeName = n
			littleValue = sum
		}
	}

	return nodeName, nil
}

func bindPod(client *kubernetes.Clientset, pod *corev1.Pod, nodeName string) error {
	return client.CoreV1().Pods(pod.Namespace).Bind(context.Background(), &corev1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		Target: corev1.ObjectReference{
			Kind:       "Node",
			APIVersion: "v1",
			Name:       nodeName,
		},
	}, metav1.CreateOptions{})
}
