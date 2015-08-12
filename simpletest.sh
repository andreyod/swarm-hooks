#Get user token
echo "Getting user token from keystone..."
UserToken=$(curl -d '{"auth":{"passwordCredentials":{"username": "doron", "password": "1234"}}}'  -H "Content-type: application/json"  http://127.0.0.1:35357/v2.0/tokens -v | jq -r '.access.token.id')
echo $UserToken
#sleep 1
#echo "Asking Info with valid token"
#curl -H "X-Auth-Token: $UserToken"  http://127.0.0.1:2379/info
#sleep 1
#echo "Creating a container..."
#ContainerId=$(curl --data-binary @redis1.json -H "X-Auth-Token:$UserToken" -H "Content-type: application/json" http://127.0.0.1:2379/containers/create | jq -r '.Id')
#echo $ContainerId
#sleep 1
#echo "Starting the container..."
#curl -X POST -H "X-Auth-Token:$UserToken" http://127.0.0.1:2379/containers/$ContainerId/start
#sleep 1
#echo "Listing containers..."
#curl -H "X-Auth-Token: $UserToken"  http://127.0.0.1:2379/containers/json?all=1 | jq '.'
#sleep 1
#echo "Deleting the container..."
#curl -X DELETE -H "X-Auth-Token:$UserToken" http://127.0.0.1:2379/containers/$ContainerId
#sleep 1
#echo "Listing containers..."
#curl -H "X-Auth-Token: $UserToken"  http://127.0.0.1:2379/containers/json?all=1 | jq '.'
echo "Checking Patch"
curl -H "X-Auth-Token: $UserToken" http://127.0.0.1:2379/containers/json?all=1 | jq '.'



#GET /containers/json?all=1&filters={%22com.ibm.tenant.0%22:[%225d69bc47d55c412d92e93e1923c35bee%22]}


# %22label%22%3Acom.ibm.tenant.0%3D5d69bc47d55c412d92e93e1923c35bee
#   "label":com.ibm.tenant.0=5d69bc47d55c412d92e93e1923c35bee

# localhost:5555/containers/json?all=1&filters={%22status%22:%5B%22exited%2‌​2%5D} 
