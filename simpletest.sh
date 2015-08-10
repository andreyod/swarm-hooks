#Get user token
echo "Getting user token from keystone..."
UserToken=$(curl -d '{"auth":{"passwordCredentials":{"username": "doron", "password": "1234"}}}'  -H "Content-type: application/json"  http://127.0.0.1:35357/v2.0/tokens -v | jq -r '.access.token.id')
echo $UserToken
sleep 1
echo "Asking Info with valid token"
curl -H "User-token: $UserToken"  http://127.0.0.1:2379/info
sleep 1
echo "Creating a container..."
ContainerId=$(curl --data-binary @redis1.json -H "User-token:$UserToken" -H "Content-type: application/json" http://127.0.0.1:2379/containers/create | jq -r '.Id')
echo $ContainerId


#sleep 1
#echo "Asking Info with invalid token"
#curl -H "User-token: Invalid"  http://127.0.0.1:2379/info

