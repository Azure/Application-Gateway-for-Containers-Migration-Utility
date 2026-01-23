package resources

import (
	core_v1 "k8s.io/api/core/v1"
)

type ServiceContext struct {
	core_v1.Service
}
