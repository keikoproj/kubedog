package structured

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

func validatePrometheusPVLabels(kubeClientset kubernetes.Interface, volumeClaimTemplatesName string) error {
	// Get prometheus PersistentVolume list
	pv, err := ListPersistentVolumes(kubeClientset)
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range pv.Items {
		pvcname := item.Spec.ClaimRef.Name
		if pvcname == volumeClaimTemplatesName+"-prometheus-k8s-prometheus-0" || pvcname == volumeClaimTemplatesName+"-prometheus-k8s-prometheus-1" {
			if k1, k2 := item.Labels["failure-domain.beta.kubernetes.io/zone"], item.Labels["topology.kubernetes.io/zone"]; k1 == "" && k2 == "" {
				return errors.Errorf("Prometheus volumes does not exist label - kubernetes.io/zone")
			}
		}
	}
	return nil
}
