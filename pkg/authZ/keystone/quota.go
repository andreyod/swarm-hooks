package keystone

import (
	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	"io/ioutil"
	"github.com/docker/swarm/pkg/authZ/utils"
	"errors"
)

type QuotaImpl struct{
	tenantMemoryQuota int64
}

var tenancyLabel = "com.swarm.tenant.0"
var CONFIG_FILE_PATH = "/root/.docker/config.json"
var DEFAULT_MEMORY_QUOTA int64 = 1024 * 1024 * 100 //100MB (Currently hardcoded for all tenant)

/*
ValidateQuota - checks if tenant quota satisfies container create request
*/
func (this QuotaImpl) ValidateQuota(myCluster cluster.Cluster, tenant string, memory float64) error{
	log.Debugf("In ValidateQuota with tenant %v and quota limit %v", tenant, this.tenantMemoryQuota)
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

	log.Debugf("tenantMemoryTotal: %v, memory: %v, tenantMemoryQuota: %v", tenantMemoryTotal, int64(memory), this.tenantMemoryQuota)
	if ((tenantMemoryTotal + int64(memory)) > this.tenantMemoryQuota){
		return errors.New("Tenant memory quota limit reached!")
	}

	return nil
}

/*
Initializing tenant quotas from config file
Example of config file (located at /root/.docker/config.json):
{
    "auths": {
                    "test1" : {
                    "auth": "TXlVc2VybmFtZTpNeVBhc3HUhJKhK",
                    "email": "myemail@gmai.com"
		            }
             },
    "HttpHeaders": {
            "X-Auth-Token": "77c2492a64c743b0b0ee9fdsdasdsadas",
            "X-Auth-TenantId": "05f44f172b0e42dabsdsadsfdewfef"
            },
    "quotas": {
            "Memory": 128
    }
}
*/
func (this *QuotaImpl) Init() error {
	log.Debugf("Initializing quotas")

    file, e := ioutil.ReadFile(CONFIG_FILE_PATH)
    if e != nil {
        log.Errorf("Failed to read tenant memory quota config from %v. Using default quota limit of %vB", CONFIG_FILE_PATH, DEFAULT_MEMORY_QUOTA)
        this.tenantMemoryQuota = DEFAULT_MEMORY_QUOTA
        return nil
    }

	var fieldType float64
	res, err := utils.ParseField("quotas.Memory", fieldType, file)
	if err != nil{
		log.Errorf("Failed to parse memory quota config from %v. Using default quota limit of %vB", CONFIG_FILE_PATH, DEFAULT_MEMORY_QUOTA)
		this.tenantMemoryQuota = DEFAULT_MEMORY_QUOTA
	}else{
		memory := res.(float64)
		log.Debugf("Setting tenant memory quota to quota limit to %vMB", memory)
		this.tenantMemoryQuota = int64(memory) * 1024 * 1024
	}

	return nil
}
