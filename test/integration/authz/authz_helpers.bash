#!/bin/bash

#FiWare credentials
export OS_TENANT_NAME=''
export OS_USERNAME=''
export OS_PASSWORD=''	
export TENANT1=''
export TENANT2=''
export KEYSTONE_IP='cloud.lab.fi-ware.org:4730'

#export DISCOVERY_FILE="/root/work/src/github.com/docker/swarm/my_cluster"
#export DISCOVERY="--multiTenant file://$DISCOVERY_FILE"

export DOCKER_IMAGE=${DOCKER_IMAGE:-dockerswarm/dind}
export DOCKER_VERSION=${DOCKER_VERSION:-1.9.0}
export STORAGE_DRIVER='aufs'

load ../helpers

function loginToKeystoneTenant1(){
	export OS_TENANT_NAME=$TENANT1
	../../../tools/set_docker_conf.sh
}

function loginToKeystoneTenant2(){
	export OS_TENANT_NAME=$TENANT2
	../../../tools/set_docker_conf.sh
}

# Start the swarm manager in background.
function swarm_manage_multy_tenant() {
	local discovery
	discovery=`join , ${HOSTS[@]}`
	
	local i=${#SWARM_MANAGE_PID[@]}
	local port=$(($SWARM_BASE_PORT + $i))
	local host=127.0.0.1:$port
	
	"$SWARM_BINARY" -l debug manage -H "$host" --heartbeat=1s --multiTenant $discovery &
	SWARM_MANAGE_PID[$i]=$!
	SWARM_HOSTS[$i]=$host
	wait_until_reachable "$host"
}
