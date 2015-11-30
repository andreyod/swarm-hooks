#!/bin/bash

#KEYSTONE_IP='cloud.lab.fi-ware.org:4730'

DOCKER_CONF=~/.docker/config.json   #default file is docker-compose config file

verbose=false
print_env() {
    echo '---------------------------'
    echo 'Docker conf file:         '$DOCKER_CONF
    echo 'OpenStack Tenant name:    '$OS_TENANT_NAME
    echo 'OpenStack Username:       '$OS_USERNAME
    echo 'OpenStack Password:       '$OS_PASSWORD
    echo 'Keystone IP:              '$KEYSTONE_IP
    echo '---------------------------'
}

display_usage() { 
    echo -e "\nThis script updates docker config file with Keystone"
	echo -e "tenant/token variables Keystone server IP must be specified"
    echo -e "either as script input or added to environment as KEYSTONE_IP"
    echo -e "variable. The rest (OS_USERNAME, OS_PASSWORD...etc.) the script"
    echo -e "may get from environment, so in most cases it's enough to"
    echo -e "source OpenStack openrc file"
    echo -e "\nIn case environment missing those variables those must be supplied as script arguments"
    echo -e "If no arguments specified will try to use defaults below:"
    print_env
    echo -e "\nUsage:\n$0 [-f CONFIG_FILE] [-t TENANT_NAME] [-u USER_NAME] [-p PASSWORD] [-a KEYSTONE_IP] [-v|-verbose] [-h|-help]\n"
    echo -e  "\nExample:\n$0 -f ~/.docker/config.json -d tcp://0.0.0.0:2376 -t \"my cloud\" -u myfiwareuser -p myfiwarepassword -a cloud.lab.fi-ware.org:4730 \n"
} 

validate_env() {
    [[ $DOCKER_CONF && $OS_TENANT_NAME && $OS_USERNAME && $OS_PASSWORD && $KEYSTONE_IP ]] || { echo -e 'ERROR! Missing one or more requiered variables\n\n'; print_env; exit 1; }
}

while getopts ":hhelp:f:d:t:u:p:a:vverbose" opt; do
      case $opt in
          h|help )
                display_usage >&2
                exit 1
                ;;
          f)
                DOCKER_CONF=$OPTARG
                ;;                
          t)
                OS_TENANT_NAME=$OPTARG
                ;;
          u)
                OS_USERNAME=$OPTARG
                ;;
          p)
                OS_PASSWORD=$OPTARG
                ;;
          a)
                KEYSTONE_IP=$OPTARG
                ;;
          v)
                verbose=true
                ;;
          \?)
                echo "Invalid option: -$OPTARG" >&2
                exit 1
                ;;
          :)
                echo "Option -$OPTARG requires an argument." >&2
                exit 1
                ;;
      esac
done

validate_env

[[ $KEYSTONE_IP != *:* ]] && KEYSTONE_IP=$KEYSTONE_IP:5000
[[ $KEYSTONE_IP != *:*/ ]] && KEYSTONE_IP=$KEYSTONE_IP/
[[ $KEYSTONE_IP != *:*/*/ ]] && KEYSTONE_IP=${KEYSTONE_IP}v2.0/
[[ $KEYSTONE_IP != http:*:*/*/ ]] && KEYSTONE_IP=http://$KEYSTONE_IP

$verbose && echo -e '\n---------------------------'
$verbose && echo 'Using following environment'
$verbose && print_env

out=`curl -s -X POST {$KEYSTONE_IP}tokens -H "Content-Type: application/json" -d '{"auth": {"tenantName": "'"$OS_TENANT_NAME"'", "passwordCredentials":{"username": "'"$OS_USERNAME"'", "password": "'"$OS_PASSWORD"'"}}}'| python -m json.tool|grep id|tail -3|head -2|awk -F"\"id\":" '{print $1,$2}'|awk -F"," '{print $1,$2}'`

test=( $out )
token=${test[0]}
tenant=${test[1]}

$verbose && echo -e "\nTOKEN: $token"
$verbose && echo -e "TENANT: $tenant\n"

sed -i '/X-Auth-Token/c\            \"X-Auth-Token\": \'${token}'\,' $DOCKER_CONF
sed -i '/X-Auth-TenantId/c\            "X-Auth-TenantId": '$tenant'' $DOCKER_CONF

$verbose && echo -e '\n\n---------------------------'
$verbose && echo 'New config file: '
$verbose && echo '---------------------------\n'
$verbose && cat $DOCKER_CONF