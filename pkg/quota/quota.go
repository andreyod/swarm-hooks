package quota

import (
	"errors"
//	"io/ioutil"
	"os"
	log "github.com/Sirupsen/logrus"
//	"github.com/docker/swarm/pkg/authZ/utils"
)

type Quota struct {
	tenantMemoryLimit int64
	tenantMemoryAvailable int64
	containers map[string]int64
}

var quotas = make(map[string]Quota)
var tenancyLabel = "com.swarm.tenant.0"
var CONFIG_FILE_PATH = os.Getenv("SWARM_CONFIG")
var DEFAULT_MEMORY_QUOTA int64 = 1024 * 1024 * 150 //150MB (Currently hardcoded for all tenant)
var DEFAULT_MEMORY int64 = 1024 * 1024 * 64        //64MB (Currently hardcoded for all tenant)


/*
ValidateQuota - checks if tenant quota satisfies container create request
*/
//func (*Quota) ValidateQuota(reqBody []byte, tenant string) error {
func (*Quota) ValidateQuota(resource int64, tenant string) error {
//	log.Debug("Going to validate quota")
//	log.Debug("Parsing requiered memory field")
//	var fieldType float64
//	var memory float64
//	res, err := utils.ParseField("HostConfig.Memory", fieldType, reqBody)
//	if err != nil {
//		log.Debugf("Failed to parse mandatory memory limit in container config, using default memory limit of %vB", DEFAULT_MEMORY)
//		memory = DEFAULT_MEMORY
//	}else{
//		memory = res.(float64)
//		
//		if memory == 0{
//			log.Debugf("Parsed memory limit is 0, using default memory limit of %vB", DEFAULT_MEMORY)
//			memory = DEFAULT_MEMORY
//		}
//	}
	if resource == 0{
		log.Debugf("Parsed memory limit is 0, using default memory limit of %vB", DEFAULT_MEMORY)
		resource = DEFAULT_MEMORY
	}
	
	
	//tenantQuota := new(Quota)
	// check if tenant in quotas key
	if tenantQuota, ok := quotas[tenant]; ok {
    	//update quota
    	log.Debug("Existing tenant")
    	log.Debug("Current action memory add: ", int64(resource))
		tenantQuota.tenantMemoryAvailable = tenantQuota.tenantMemoryAvailable - resource
		quotas[tenant] = tenantQuota
		log.Debug("New available: ", tenantQuota.tenantMemoryAvailable)
	}else{// if not create entry with tenant=defaultQuota
		log.Debug("New tenant")
		tenantQuota.tenantMemoryLimit = DEFAULT_MEMORY_QUOTA
		tenantQuota.tenantMemoryAvailable = DEFAULT_MEMORY_QUOTA - resource
        quotas[tenant] = tenantQuota
	}
	
	//sanaty
	for key, value := range quotas {
	    log.Debug("Tenant: ", key, " Limit: ", value.tenantMemoryLimit," Available: ", value.tenantMemoryAvailable)
	}

	if (quotas[tenant].tenantMemoryAvailable < 0) {
		// need temp var. bug in go. https://github.com/golang/go/issues/3117
		revertQuota := quotas[tenant]
		revertQuota.tenantMemoryAvailable = revertQuota.tenantMemoryAvailable + resource
		quotas[tenant]= revertQuota
		//quotas[tenant].tenantMemoryAvailable = quotas[tenant].tenantMemoryAvailable + int64(memory)
		return errors.New("Tenant memory quota limit reached!")
	}
	return nil
}

func (*Quota) UpdateQuota(tenant string, toFree bool) error {
	if tenantQuota, ok := quotas[tenant]; ok {
		log.Debug("Quota exists for this tenant")
		if toFree{
			log.Debug("Free resources. Current availible:")
			log.Debug(tenantQuota.tenantMemoryAvailable)
			tenantQuota.tenantMemoryAvailable = tenantQuota.tenantMemoryAvailable + DEFAULT_MEMORY
		} else{
			tenantQuota.tenantMemoryAvailable = tenantQuota.tenantMemoryAvailable - DEFAULT_MEMORY
		}
		
		log.Debug(tenantQuota.tenantMemoryLimit)
		log.Debug(tenantQuota.tenantMemoryAvailable)
		quotas[tenant] = tenantQuota
	} else {
		log.Debug("No quota exists for this tenant")
	}
	return nil
}

//Init - Any required initialization
func (*Quota) Init() error {
	//start event listener in new thread
	//startEventListener()
	//log.Debug("Init QUOTA INIT -----------------------------------")
	//go RunServer()
	//RunServer()
	log.Debug("Init QUOTA INIT -----------------------------------")
	//go RunListener()
	return nil
}

//func CreateDefaultLimits(string tenant){
	
//}

//func UpdateAvailable(string tenant){
	
//}
