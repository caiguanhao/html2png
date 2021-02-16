# html2png

Uses chromedp's screenshot function to convert HTML to PNG image.

```
go get -v -u github.com/caiguanhao/html2png
```

You must install Chrome first. If you don't want to install Chrome, you can use
the [headless-shell](https://github.com/chromedp/docker-headless-shell) Docker
image.

```
echo '<body bgcolor=black><font color=white>Hello, <b>World</b>!</font></body>' | \
    html2png -d "width=120&height=34&mobile=false" -o hello.png
```

Will generate a temporary image file like this:

<img src="./hello.png" width="120" height="34">

## Usage

```
Usage of html2png:
  -d value
        short version of --device
  -device value
        device number or name from --devices or custom device properties like:
        -d name=string -d useragent=string -d width=int -d height=int -d scale=float -d landscape=bool -d mobile=bool -d touch=bool
  -devices
        list all mobile devices and exit
  -full
        capture full web page
  -i string
        input HTML file name or HTTP URL, default is stdin
  -o string
        output file name, default is stdout, "-" to use temporary file
  -open
        open the output file
  -v    verbose
  -ws string
        WebSocket debugger URL
```

## headless-shell

```
# start headless-shell container
docker run -d -p 127.0.0.1:9222:9222 --rm --name headless-shell --shm-size 2G chromedp/headless-shell

# use the -ws option
html2png -i https://en.wikipedia.org/wiki/Golang -o page.png -full -ws "$(curl -s http://localhost:9222/json/version | jq -r .webSocketDebuggerUrl)"
```

To make headless-shell display CJK (Chinese, Japanese, Korean) characters
correctly, you can install the fonts-noto-cjk package. To hide the scrollbars,
you can add the --hide-scrollbars flag. Example Dockerfile:

```Dockerfile
FROM chromedp/headless-shell

RUN \
    apt-get update -y \
    && apt-get install -y fonts-noto-cjk \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ENTRYPOINT [ "/headless-shell/headless-shell", "--no-sandbox", "--hide-scrollbars", "--remote-debugging-address=0.0.0.0", "--remote-debugging-port=9222" ]
```
