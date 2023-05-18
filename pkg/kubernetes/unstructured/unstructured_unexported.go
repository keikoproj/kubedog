package unstructured

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
)

func validateDynamicClient(dynamicClient dynamic.Interface) error {
	if dynamicClient == nil {
		return errors.Errorf("'k8s.io/client-go/dynamic.Interface' is nil.")
	}
	return nil
}
