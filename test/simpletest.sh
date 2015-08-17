echo "Count all containers..."
docker -H 127.0.0.1:2375 ps -a | wc -l
#Get user token
echo "Getting user token from keystone..."
UserToken=$(curl -d '{"auth":{"passwordCredentials":{"username": "doron", "password": "1234"}}}'  -H "Content-type: application/json"  http://127.0.0.1:35357/v2.0/tokens -v | jq -r '.access.token.id')
echo $UserToken
sleep 1
echo "Asking Info with valid token"
curl -H "X-Auth-Token: $UserToken"  http://127.0.0.1:2379/info
sleep 1
echo "Creating a container..."
ContainerId=$(curl --data-binary @redis1.json -H "X-Auth-Token:$UserToken" -H "Content-type: application/json" http://127.0.0.1:2379/containers/create | jq -r '.Id')
echo $ContainerId
sleep 1
echo "Starting the container..."
curl -X POST -H "X-Auth-Token:$UserToken" http://127.0.0.1:2379/containers/$ContainerId/start
sleep 1
echo "Listing containers..."
curl -H "X-Auth-Token: $UserToken"  http://127.0.0.1:2379/containers/json?all=1 | jq '.'
sleep 1
echo "Stopping the container..."
curl -X POST -H "X-Auth-Token:$UserToken" http://127.0.0.1:2379/containers/$ContainerId/stop
sleep 1
echo "Deleting the container..."
curl -X DELETE -H "X-Auth-Token:$UserToken" http://127.0.0.1:2379/containers/$ContainerId
sleep 1
echo "Listing containers..."
curl -H "X-Auth-Token: $UserToken"  http://127.0.0.1:2379/containers/json?all=1 | jq '.'
echo "Count all containers..."
docker -H 127.0.0.1:2375 ps -a | wc -l

