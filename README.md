![alt tag](https://upload.wikimedia.org/wikipedia/commons/2/23/Golang.png)

[![License](https://img.shields.io/github/license/Massad/gin-boilerplate)](https://github.com/Massad/gin-boilerplate/blob/master/LICENSE) [![GitHub release (latest by date)](https://img.shields.io/github/v/release/Massad/gin-boilerplate)](https://github.com/Massad/gin-boilerplate/releases) [![Go Version](https://img.shields.io/github/go-mod/go-version/Massad/gin-boilerplate)](https://github.com/Massad/gin-boilerplate/blob/master/go.mod) [![DB Version](https://img.shields.io/badge/DB-PostgreSQL--latest-blue)](https://github.com/Massad/gin-boilerplate/blob/master/go.mod) [![DB Version](https://img.shields.io/badge/DB-Redis--latest-blue)](https://github.com/Massad/gin-boilerplate/blob/master/go.mod)

[![Build Status](https://travis-ci.org/Massad/gin-boilerplate.svg?branch=master)](https://travis-ci.org/Massad/gin-boilerplate) [![Go Report Card](https://goreportcard.com/badge/github.com/Massad/gin-boilerplate)](https://goreportcard.com/report/github.com/Massad/gin-boilerplate)

[![Join the chat at https://gitter.im/Massad/gin-boilerplate](https://badges.gitter.im/Massad/gin-boilerplate.svg)](https://gitter.im/Massad/gin-boilerplate?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Welcome to **Golang Gin boilerplate** v2

The fastest way to deploy a restful api's with [Gin Framework](https://gin-gonic.github.io/gin/) with a structured project that defaults to **PostgreSQL** database and **JWT** authentication middleware stored in **Redis**

## Configured with

- [go-gorp](https://github.com/go-gorp/gorp): Go Relational Persistence
- [jwt-go](https://github.com/dgrijalva/jwt-go): JSON Web Tokens (JWT) as middleware
- [go-redis](https://github.com/go-redis/redis): Redis support for Go
- Go Modules
- Built-in **CORS Middleware**
- Built-in **RequestID Middleware**
- Feature **PostgreSQL 12** with JSON/JSONB queries & trigger functions
- SSL Support
- Enviroment support
- Unit test
- And few other important utilties to kickstart any project

### Installation

```
$ go get github.com/Massad/gin-boilerplate
```

```
$ cd $GOPATH/src/github.com/Massad/gin-boilerplate
```

```
$ go mod init
```

```
$ go install
```

You will find the **database.sql** in `db/database.sql`

And you can import the postgres database using this command:

```
$ psql -U postgres -h localhost < ./db/database.sql
```

Tip:

You will find that we added 2 trigger functions to the dabatase:

- `public.created_at_column()`
- `public.update_at_column()`

Those are added to the `updated_at` and `created_at` columns to update the latest timestamp automatically in both **user** and **article** tables. You can explore the tables and public schema for more info.

## Running Your Application

Rename .env_rename_me to .env and place your credentials

```
$ mv .env_rename_me .env
```

Generate SSL certificates (Optional)

> If you don't SSL now, change `SSL=TRUE` to `SSL=FALSE` in the `.env` file

```
$ mkdir cert/
```

```
$ sh generate-certificate.sh
```

> Make sure to change the values in .env for your databases

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

Download [Postman](https://www.getpostman.com/) -> Import -> Import From Link

https://www.postman.com/collections/7f941b400a88ddd9c137

Includes the following:

- User
  - Login
  - Register
  - Logout
- Article
  - Create
  - Update
  - Get Article
  - Get Articles
  - Delete
- Auth
  - Refresh Token

> In Login request in Tests tab:

```
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);

    var jsonData = JSON.parse(responseBody);
    pm.globals.set("token", jsonData.token.access_token);
    pm.globals.set("refresh_token", jsonData.token.refresh_token);

});
```

It captures the `access_token` from the success login in the **global variable** for later use in other requests.

Also, you will find in each request that needs to be authenticated you will have the following:

    Authorization -> Bearer Token with value of {{token}}

It's very useful when you want to test the APIs in Postman without copying and pasting the tokens.

## On You

You will need to implement the `refresh_token` mechanism in your application (Frontend).

> We have the `/v1/token/refresh` API here to use it.

_For example:_

If the API sends `401` Status Unauthorized, then you can send the `refresh_token` that you stored it before from the Login API in POST `/v1/token/refresh` to receive the new `access_token` & `refresh_token` and store them again. Now, if you receive an error in refreshing the token, that means the user will have to Login again as something went wrong.

That's just an example, of course you can implement your own way.

## Version 1

    No longer supported

You will find the last update on v1 in [v1-session-cookies-auth](https://github.com/Massad/gin-boilerplate/tree/v1-session-cookies-auth) branch or [v1.0.5 release](https://github.com/Massad/gin-boilerplate/releases/tag/1.05) that supported the authentication using the **session** and **cookies** stored in **Redis** if needed.

- [RedisStore](https://github.com/gin-gonic/contrib/tree/master/sessions): Gin middleware for session management with multi-backend support (currently cookie, Redis).

## Contribution

You are welcome to contribute to keep it up to date and always improving!

If you have any question or need help, drop a message at [https://gitter.im/Massad/gin-boilerplate](https://gitter.im/Massad/gin-boilerplate)

## Credit

The implemented JWT inspired from this article: [Using JWT for Authentication in a Golang Application](https://www.nexmo.com/blog/2020/03/13/using-jwt-for-authentication-in-a-golang-application-dr) worth reading it, thanks [Victor Steven](https://medium.com/@victorsteven)

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
