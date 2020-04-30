~~~
$ ./TunnelXHTTP -tls -domain httpbin.org:443 -mode 2
~~~

~~~
$ ./TunnelXHTTP -help

Usage of ./TunnelXHTTP:
  -domain string
    	 (default "example.com:443")
  -k	TLS Insecure Skip Verify
  -listen string
    	 (default ":8000")
  -mode int
    	0 [conn timeout] | 1 [http 500] | 2 [proxy]
  -tls
    	With TLS
~~~

~~~
$ curl https://0.0.0.0:8000/vinicio/jose/valbuena -k -H "name:vincio" -d 'hola mundo'
$ curl http://0.0.0.0:8000/vinicio/jose/valbuena     -H "name:vincio" -d 'hola mundo'
~~~

~~~
### multipart/form-data
$ curl 0.0.0.0:8000 -F "image=@/home/formatcom/image.png"
~~~

~~~
### dump
$ sudo tcpdump -A -i lo dst port 8000
~~~
