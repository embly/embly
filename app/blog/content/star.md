+++
title = "Star: Go but in Python?"
date = 2020-02-01T22:32:00Z
[extra]
author = "Max McDonnell"
+++

[Starlark](https://github.com/bazelbuild/starlark) is Google's custom subset of Python that it uses as a configuration language with Bazel. I started looking into it a little because it has some interesting characteristics. It's not turing complete, it doesn't have classes or higher level abstractions, it's also deterministic and is somewhat safe to run untrusted user input.

Although somewhat limited it felt like it might make a good candidate for a serverless function runtime. Easier to sandbox, could mock out pieces of the python standard library, maybe provide just enough functionality for it to be useful. I also thought you might be able to identify the deterministic blocks of code and replace them with the static result on subsequent runs. Potentially exciting stuff!

I used [starlark-go](https://github.com/google/starlark-go) to get started and pretty quickly had this demo up and running:

```python
def hello(w, req):
    w.write(req.content_type)
    w.write(req.path)
    w.write("\nHello World\n" + str([x for x in range(10)]) + "\n")

    return
```
The server parses the file, takes the `hello` function and passes the `http.ResponseWriter` and `http.Request` over to the ~~python~~ starlark side of things. From there I could add attributes to pass over things like the `path` and `content_type`.

Starlark doesn't have exception handling, so I add a Go error type and passed that over as well. This ends up being almost as verbose as Go, but without typed functions it does allow for a little more flexibility:
```python
def handle_err(args):
    for a in args:
        if type(a) == "error" and a:
            print(a, a.stacktrace())
    return args

foo, err = handle_err(function_call())
```

At this point I got a little more interested in continuing to cram Go functionality into this scripting language. The [starlark-go](https://github.com/google/starlark-go) implementation is really wonderful and it was quite easy to extend.

In talking to my coworker I wondered if you could pull in lots of Go functionality, we sketched up something like this:

```python
import time
from net import http
from star import chan, go

def foo(c):
    time.sleep(5)
    c.push("hello")

def main():
    resp, err = http.get("http://www.embly.run")
    if err:
        return print(err)

    c = chan(1)
    go(foo, c)
    c.next()
```

From a syntax standpoint it seems like this would be possible. Starlark also has first-class support for running separate chunks of code in parallel (globals are immutable, which helps), so leveraging Go's concurrency model didn't seem too difficult either.

After a little more messing around, most of that sketch has been implemented: https://github.com/embly/star

[**star**](https://github.com/embly/star) is a python(ish) programming environment that lets you call Go library functions. It is very fragile and shouldn't be taken seriously, but it's very fun to play with.

Here's what that sketch ended up looking like:

```python
http = require("net/http")
ioutil = require("io/ioutil")
sync = require("sync")
star = require("star")
time = require("time")

def get_url(url, wg):
    resp, err = http.Get(url)
    if err:
        return print(err)
    b, err = ioutil.ReadAll(resp.Body)
    if err:
        return print(err)
    body, err = star.bytes_to_string(b)
    if err:
        return print(err)
    time.Sleep(time.Second * 2)
    wg.Done()


def main():
    wg = sync.WaitGroup()
    wg.Add(3)
    urls = [
        "https://api.exchangeratesapi.io/latest",
        "https://api.exchangeratesapi.io/latest",
        "https://api.exchangeratesapi.io/latest",
    ]
    for url in urls:
        star.go(get_url, url, wg)
    wg.Wait()
```

Really trippy to look at if you've written python and Go. A few notes:
 -  starlark doesn't support imports, so I went with a `require` function
 - `star.go` spawns a goroutine and it works! adding channels wouldn't be too hard either
 -  `star.bytes_to_string` shows some of the cracks, I wasn't too sure how to add type conversion


The coolest part is that this is all pure Go, so we can also run all of it in webassembly. Here's a repl, script away!:
