package mutant

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
)

// Status is a struct required to determine whether the specified
// Kubernetes workload is mutant or not.
type Status map[string][]ContainerImageID

type ContainerImageID struct {
	Name    string
	ImageID string
}

func (s Status) IsMutant() bool {
	c := make(map[string]map[string]bool)
	for _, containerImages := range s {
		for _, containerImage := range containerImages {
			if _, ok := c[containerImage.Name]; !ok {
				c[containerImage.Name] = make(map[string]bool)
			}

			c[containerImage.Name][containerImage.ImageID] = true
		}
	}

	for _, digests := range c {
		if len(digests) > 1 {
			return true
		}
	}

	return false
}

func GetStatus(pods []corev1.Pod) Status {
	status := make(map[string][]ContainerImageID)
	for _, pod := range pods {
		var cii []ContainerImageID
		for _, cs := range pod.Status.ContainerStatuses {
			cii = append(cii, ContainerImageID{Name: cs.Name, ImageID: cs.ImageID})
		}
		status[pod.Name] = cii
	}
	return status
}

func (s Status) AsJson() (string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
