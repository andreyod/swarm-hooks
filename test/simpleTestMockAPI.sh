echo "Count all containers + 1 ..."
docker -H 127.0.0.1:2375 ps -a | wc -l
#Get user token
echo "Getting user token..."
UserToken=Space1
echo $UserToken
sleep 1
echo "Asking Info with valid token..."
curl -H "X-Auth-Token: $UserToken"  http://127.0.0.1:2379/info
sleep 1
echo "Creating a container..."
#curl -v --data-binary @redis1.json -H "X-Auth-Token:$UserToken" -H "Content-type: application/json" http://127.0.0.1:2379/containers/create?name=Doron
ContainerId=$(curl --data-binary @redis1.json -H "X-Auth-Token:$UserToken" -H "Content-type: application/json" http://127.0.0.1:2379/containers/create | jq -r '.Id')
echo $ContainerId
sleep 1
echo "Starting the container..."
curl -v -X POST -H "X-Auth-Token:$UserToken" http://127.0.0.1:2379/v1.18/containers/$ContainerId/start
sleep 1
echo "Listing containers..."
curl -H "X-Auth-Token: $UserToken"  http://127.0.0.1:2379/containers/json?all=1 | jq '.'
sleep 1

#echo "showing the underline lable mechanisem..."
#docker -H 127.0.0.1:2375 inspect $ContainerId
#sleep 1

echo "Getting Another user token..."
UserToken2=Space2
echo $UserToken2

echo "Listing containers with the other user token..."
curl -H "X-Auth-Token: $UserToken2"  http://127.0.0.1:2379/containers/json?all=1 | jq '.'
sleep 1

echo "Listing containers with the admin user token..."
curl -H "X-Auth-Token: admin"  http://127.0.0.1:2379/containers/json?all=1 | jq '.'
sleep 1

echo "Inspecting the container..."
curl -H "X-Auth-Token:$UserToken" http://127.0.0.1:2379/containers/$ContainerId/json | jq '.'
sleep 2

echo "Getting the container logs..."
curl -H "X-Auth-Token:$UserToken" "http://127.0.0.1:2379/containers/$ContainerId/logs?stderr=1&stdout=1&timestamps=1&follow=0&tail=10"
sleep 2

echo "Stopping the container..."
curl -X POST -H "X-Auth-Token:$UserToken" http://127.0.0.1:2379/containers/$ContainerId/stop | jq '.'
sleep 1

echo "Inspecting the container again..."
curl -H "X-Auth-Token:$UserToken" http://127.0.0.1:2379/containers/$ContainerId/json | jq '.'
sleep 1

echo "Trying to Delete the container with the other user token..."
curl -v -X DELETE -H "X-Auth-Token:$UserToken2" http://127.0.0.1:2379/containers/$ContainerId
sleep 1

echo "Deleting the container..."
curl -X DELETE -H "X-Auth-Token:$UserToken" http://127.0.0.1:2379/containers/$ContainerId
sleep 1

echo "Listing containers..."
curl -H "X-Auth-Token: $UserToken"  http://127.0.0.1:2379/containers/json?all=1 | jq '.'
echo "Count all containers + 1 ..."
docker -H 127.0.0.1:2375 ps -a | wc -l

