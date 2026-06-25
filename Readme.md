# ID as a Service

Through some hearsay, I came across [this post](https://news.ycombinator.com/item?id=48061235):

> Funny story no one will believe, but it’s true. A good friend of mine joined a
> startup as CTO 10 years ago, high growth phase, maybe 200 devs… In his first
> week he discovered the company had a microservice for generating new UUIDs.
> One endpoint with its own dedicated team of 3 engineers …including a database
> guy (the plot thickens). Other teams were instructed to call this service
> every time they needed a new ‘safe’ UUID. My pal asked wtf. It turned out this
> service had its own DB to store every previously issued UUID. Requests were
> handled as follows: it would generate a UUID, then ‘validate’ it by checking
> its own database to ensure the newly generated UUID didn’t match any
> previously generated UUIDs, then insert it, then return it to the client.
> Peace of mind I guess. The team had its own kanban board and sprints.

One day, I was wondering if I could implement something like this in under 100
lines of Go code? The answer is yes.

## Features

The program creates an HTTP server (listening on 8080) which, for each request,
returns a unique ID in its body as `text/plain`.

```shell
curl localhost:8080 -v
*   Trying 127.0.0.1:8080...
* Connected to localhost (127.0.0.1) port 8080 (#0)
> GET / HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.81.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Content-Type: text/plain
< Date: Thu, 11 Jun 2026 06:35:19 GMT
< Content-Length: 19
<
* Connection #0 to host localhost left intact
1781159719417711419
```

The ID itself is (roughly) the current timestamp with nanosecond precision.
By using atomics, the generator can ensure unique IDs while also being wait free.

### v1.0

The API returns the timestamp itself.

### v1.1

To avoid exposing the creation time of an ID and have it more "random" instead (similar to v4-UUIDs), `v1.1` hashes the created ID with [XXH3](https://github.com/cyan4973/xxhash) and a secret seed.
Note that the seed is hard-coded and you need to replace it with your own.
Since XXH3 is bijective for 64-bit inputs (see https://github.com/Cyan4973/xxHash/issues/236#issuecomment-522051621), it is still guaranteed that no ID is produced twice.
