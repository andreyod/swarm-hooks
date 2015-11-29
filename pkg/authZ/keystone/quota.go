package keystone

import (
	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
)

type QuotaImpl struct{}

var tenancyLabel = "com.swarm.tenant.0"

var tenantMemoryQuota int64 = 1024 * 1024 * 100 //100MB (Currently hardcoded for all tenant)

/*
ValidateQuota - checks if tenant quota satisfies container create request
*/
func (QuotaImpl) ValidateQuota(myCluster cluster.Cluster, tenant string, memory float64) bool{
	log.Debugf("In ValidateQuota with tenant %v and quota limit %v", tenant, tenantMemoryQuota)
	containers := myCluster.Containers()

	var tenantMemoryTotal int64 = 0
	for _, container := range containers {
		log.Debugf("Container %v tenant Label: %v", container.Id, container.Labels[tenancyLabel])
		log.Debugf("Container name: %v", container.Names[0])
		if container.Labels[tenancyLabel] == tenant {
			memory := container.Config.Memory
			log.Debugf("Incrementing total memory %v by %v", tenantMemoryTotal, memory)
			tenantMemoryTotal += memory
		}
	}

	log.Debugf("tenantMemoryTotal: %v, memory: %v, tenantMemoryQuota: %v", tenantMemoryTotal, int64(memory), tenantMemoryQuota)
	return ((tenantMemoryTotal + int64(memory)) < tenantMemoryQuota)
}

//Init - Here can be default initialization from file
func (*QuotaImpl) Init() error {
	return nil
}
