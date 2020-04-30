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
