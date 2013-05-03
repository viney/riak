#! /bin/bash

export PATH=$PATH:/var/lib/jshon
PWDDIR="$(dirname "`readlink -f $0`")"

which jshon
if [ "$?" == "1" ];then
	echo "jshon uninstalled!"
	sudo apt-get install libjansson-dev
	wget http://kmkeen.com/jshon/jshon.tar.gz 
    tar zxvf jshon.tar.gz
	cd $PWDDIR/jshon-20120914 && make
    mv $PWDDIR/jshon-20120914 $PWDDIR/jshon
    sudo cp -rf $PWDDIR/jshon/ /var/lib/
	echo "jshon installed!"
	exit 1
fi

URI=127.0.0.1:8098
FLAGS=

if [ "$1" == "DEBUG" ];then
	FLAGS=-v
fi

function deleteByBucket() {
	keys=`curl -s http://$URI/buckets/$1/keys?keys=true | jshon -e keys`
	for k in $keys
	do
		if [ "$k" == "[" -o "$k" == "]" -o "$k" == "[]" ];then
			continue
		fi
		tmp=${k#\"}
		key=${tmp%\"*}
		key=${key/\%/\%25}
		echo "deleting bucket: $1, key: $key..."
		curl $FLAGS -X DELETE http://$URI/buckets/$1/keys/$key
	done
}

buckets="loc_reg loc_login"
for bucket in $buckets
do
	deleteByBucket $bucket
done
