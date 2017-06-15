![alt tag](https://upload.wikimedia.org/wikipedia/commons/2/23/Golang.png)

[![Build Status](https://travis-ci.org/Massad/gin-boilerplate.svg?branch=master)](https://travis-ci.org/Massad/gin-boilerplate)
[![Join the chat at https://gitter.im/Massad/gin-boilerplate](https://badges.gitter.im/Massad/gin-boilerplate.svg)](https://gitter.im/Massad/gin-boilerplate?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Welcome to **Golang Gin boilerplate**!

The fastest way to deploy a restful api's with [Gin Framework](https://gin-gonic.github.io/gin/) with a structured project that defaults to **PostgreSQL** database and **Redis** as the session storage.

## Configured with

* [go-gorp](https://github.com/go-gorp/gorp): Go Relational Persistence
* [RedisStore](https://github.com/gin-gonic/contrib/tree/master/sessions): Gin middleware for session management with multi-backend support (currently cookie, Redis).
* Built-in **CORS Middleware**
* Feature **PostgreSQL 9.6** JSON queries
* Unit test

### Installation

```
$ go get github.com/Massad/gin-boilerplate
```

```
$ cd $GOPATH/src/github.com/Massad/gin-boilerplate
```

```
$ go get -t -v ./...
```

> Sometimes you need to get this package manually
```
$ go get github.com/bmizerany/assert
```

You will find the **database.sql** in `db/database.sql`

And you can import the postgres database using this command:
```
$ psql -U postgres -h localhost < ./db/database.sql
```

## Running Your Application

```
$ go run *.go
```

## Building Your Application

```
$ go build -v
```

```
$ ./gin-boilerplate
```

## Testing Your Application

```
$ go test -v ./tests/*
```


## Import Postman Collection (API's)
You can import from this [link](https://www.getpostman.com/collections/ac0680f90961bafd5de7). If you don't have **Postman**, check this link [https://www.getpostman.com](https://www.getpostman.com/)

## Contribution

You are welcome to contribute to keep it up to date and always improving!

If you have any question or need help, drop a message at [https://gitter.im/Massad/gin-boilerplate](https://gitter.im/Massad/gin-boilerplate)

---

## License
(The MIT License)

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
'Software'), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
