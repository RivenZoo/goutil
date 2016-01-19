#### Gurl
	A http query tool for test. It can be used in single machine or be deployed to multi machine and run query at the same time.

##### Build
	make
##### Usage
	gurl use three modes 
	mode query, run query test, default mode
	mode agent, receive command and report result, set addrs to agent address
	mode command, send command to agent, set addrs to receive address
	Usage of ./build/gurl:
	  -H string
        	set request header, multi split by';', eg. "Host:abc.com;Agent:Firefox"
	  -addrs string
	    	address 'ip:port' concatenated by ';'
	  -cli int
	    	query client number (default 1)
	  -cmd string
	    	send cmd to agent, support command:[query|exit]
	  -d string
	    	post data
	  -en string
	    	post data encode type, support:base64
	  -m string
	    	query method, GET or POST (default "GET")
	  -mode string
	    	mode:[agent|cmd|query] (default "query")
	  -n int
	    	query times per client (default 1)
	  -url string
	    	url to query
##### Example
	Mode Query
		$ ./build/gurl -url=http://github.com/ # GET with default query mode
		$ ./build/gurl -d='{"key":"abc"}' -m=POST -url='http://yourweb.com' # POST string to url
		$ ./build/gurl -d='CAACBA==' -en=base64 -url='http://yourweb.com' # decode base64 param and POST data to url
		$ ./build/gurl -d='{"key":"abc"}' -m=POST -url='http://yourweb.com' -cli=10 -n=10 # 10 client and every client POST 10 times to url
		
	Mode Agent
		$ ./build/gurl -mode=agent # start agent and listen 0.0.0.0:9012
		$ ./build/gurl -mode=agent -addrs="0.0.0.0:9090" # set listen address
		
	Mode Command
		$ ./build/gurl -mode=cmd -addrs="127.0.0.1:9012" -cmd=exit # send exit command to agent
		$ ./build/gurl -mode=cmd -addrs="10.10.60.10:9012;10.10.60.11:9012" -cmd=query -url="http://github.com" -d='CAACBA==' -en=base64 -cli=10 -n=10 # send query to 2 agent and every agent run 10 client and every client post 10 request
