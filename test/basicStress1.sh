#Get user token
echo "Getting user token from keystone..."
UserToken=$(curl -H "X-Auth-Token:abfe604f65e6b337c0f2" -H "User:red_user1" -H "Password:passw0rd" http://127.0.0.1:3000/authenticateUser)
echo $UserToken
sleep 1

Time1=$(date +%s )
num=100

for i in `seq 1 $num`;
        do

#Create container
echo "Creating a container..."
#ContainerId=$(curl --data-binary @redis1.json -H "User-token:$UserToken" -H "Content-type: application/json" http://127.0.0.1:3000/containers/create | jq -r '.Id')
ContainerId=$(curl --data-binary @redis1.json -H "Content-type: application/json" http://127.0.0.1:2377/containers/create | jq -r '.Id')
echo $ContainerId

#Start container
echo "Starting the container..."
#curl -X POST -H "User-token:$UserToken" http://127.0.0.1:3000/containers/$ContainerId/start
curl -X POST http://127.0.0.1:2377/containers/$ContainerId/start

done

Time2=$(date +%s )

Count=`expr $Time2 - $Time1`
echo ""
echo "Took: $Count Seconds"

echo "Created and started $num containers"

up=$(docker -H 127.0.0.1:2377 ps | wc -l)
#up=$(docker -H 10.120.39.34:2375 ps | wc -l)
up=$((up-1))

echo "$up are up"
#List containers (see status is up if need be)
#echo "Listing containers..."
#curl -H "User-token: $UserToken"  http://127.0.0.1:3000/containers/json?all=1 | jq '.'
#sleep 1
#Delete container
#echo "Deleting the container..."
#curl -X DELETE -H "User-token:$UserToken" http://127.0.0.1:3000/containers/$ContainerId
#sleep 1
#List again (Container should be gone)
#echo "Listing containers..."
#curl -H "User-token: $UserToken"  http://127.0.0.1:3000/containers/json?all=1 | jq '.'

