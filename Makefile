testenv:
	mkdir -p tmpdb/data{0,1,2}
	mkdir -p tmpdb/log
	mongod --dbpath tmpdb/data0 --port 30000 --replSet=testRs --noprealloc --oplogSize=5 --fork --logpath tmpdb/log/0.log && sleep 2
	mongod --dbpath tmpdb/data1 --port 30001 --replSet=testRs --noprealloc --oplogSize=5 --fork --logpath tmpdb/log/1.log && sleep 2
	mongod --dbpath tmpdb/data2 --port 30002 --replSet=testRs --noprealloc --oplogSize=5 --fork --logpath tmpdb/log/2.log && sleep 2
	mongo --port 30000 --eval 'rs.initiate({"_id" : "testRs", members: [ {"host" : "127.0.0.1:30000", "_id" : 1}, {"host" : "127.0.0.1:30001", "_id" : 2}, {"host" : "127.0.0.1:30002", "_id" : 3 } ]})'

clean:
	@-`ps ax | grep tmpdb | grep -v grep | awk   '{print $1}'  | xargs -n1 kill`
	@-rm -rf tmpdb/
