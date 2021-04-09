# jetis

A trivial http proxy that throws off https encryption. This allows for payload packet capture of https connections with tools like `tcpdump`.

## Usage

To sniff a request to a `https://` url with `tcpdump`:

- start `jetis`

```
$ jetis
starting proxy server on localhost:8888 ...
```

- point your http client to the http proxy `http://localhost:8888`
- in your request replace the `https://` url you want to request with `http://` (jetis will modify it to a `https://` url in flight)

```
curl --proxy http://localhost:8888 http://<url>
```


- sniff the traffic between client and proxy - it's unencrpyted!

```
tcpdump -i lo "host localhost and port 8888"
```

## Usage example

Let's try to see the contents of `https://github.com/robots.txt` with `tcpdump`.

If we would run `curl https://github.com/robots.txt` and sniff with `tcpdump` on the interface we would usually only see encrypted traffic because:

- this URL is https only (so tcpdump only sees encrypted traffic)
- and the `http://...` URL returns a 301 redirect to https

As a workaround we can use `jetis` as a proxy server in between. It converts any requested `http://` URL to a `https://` URL in flight. Its proxy server port is plain http, so we can sniff the traffic between our client (`curl`) and the proxy with `tcpdump`.

Start `jetis`:
```
$ jetis
starting proxy server on localhost:8888 ...
```

In another terminal start tcpdump on the local loopback interface:
```
tcpdump -i lo -l -w - port 8888 | tcpflow -C -r -
```

In another terminal we can make our curl request:


```
curl --proxy http://localhost:8888 http://github.com/robots.txt
```

Note that we use a `http://` URL in our request. In the `jetis` output, we can see that it automatically modified it to `https://` before making the request against the actual server:

```
2021/04/09 17:15:13 📨	original url: 	 http://github.com/robots.txt
2021/04/09 17:15:13 ✏️	new url: 	 https://github.com/robots.txt

```

And therefore `tcpdump` has seen the response data in plain text:

```
<...>
User-agent: *

Disallow: /*/pulse
<...>
```

## Attribution

This is a quick hack on top of the great [work](https://medium.com/@mlowicki/http-s-proxy-in-golang-in-less-than-100-lines-of-code-6a51c2f2c38c) by Michał Łowicki, licensed under [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/). 



## FAQ

### what happens if I request a `https://` URL?

it gets passed through, but the traffic is encrypted

### are there limitations?

Lots. (301 redirects to https are encrypted (don't rely on `curl -L`), no http/2, no hbh headers, etc etc). This is a convenience toy tool for simple use cases. When in doubt you probably want to use a proper "inspection" proxy like [Charles](https://www.charlesproxy.com/).

### why this name?

> "to jettison (aviation): to throw off from a moving aircraft."