package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"syscall"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/term"
)

func die(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

func main() {
	var inputFilename string
	var outputFilename string
	var openFile bool
	var fullPage bool
	var verbose bool

	showDevices := flag.Bool("devices", false, "list all mobile devices and exit")
	deviceToUse := devices.NewDeviceByName("iPhone 8")
	flag.Var(deviceToUse, "device", "device number or name from --devices or custom device properties like:\n"+
		"-d name=string -d useragent=string -d width=int -d height=int -d scale=float -d landscape=bool -d mobile=bool -d touch=bool")
	flag.Var(deviceToUse, "d", "short version of --device")
	flag.BoolVar(&fullPage, "full", false, "capture full web page")
	flag.StringVar(&inputFilename, "i", "", "input HTML file name or HTTP URL, default is stdin")
	flag.StringVar(&outputFilename, "o", "", `output file name, default is stdout, "-" to use temporary file`)
	flag.BoolVar(&openFile, "open", false, "open the output file")
	flag.BoolVar(&verbose, "v", false, "verbose")
	webSocketDebuggerUrl := flag.String("ws", "", "WebSocket debugger URL")
	flag.Parse()

	if *showDevices {
		fmt.Println(devices)
		return
	}

	if verbose {
		fmt.Fprintln(os.Stderr, "Device:")
		fmt.Fprintln(os.Stderr, deviceToUse.MultilineStringIndent(4))
	}

	var inputUrl string
	if inputFilename == "" {
		f, err := ioutil.TempFile("", "stdin.*.html")
		if err != nil {
			die(err)
		}
		defer os.Remove(f.Name())
		if _, err := io.Copy(f, os.Stdin); err != nil {
			die(err)
		}
		if err := f.Close(); err != nil {
			die(err)
		}
		inputUrl = "file://" + f.Name()
	} else {
		if _, err := url.ParseRequestURI(inputFilename); err == nil {
			inputUrl = inputFilename
		} else {
			p, err := filepath.Abs(inputFilename)
			if err != nil {
				die(err)
			}
			_, err = os.Stat(p)
			if err != nil {
				die(err)
			}
			inputUrl = "file://" + p
		}
	}

	var target io.WriteCloser
	var targetFileName string
	if outputFilename == "" {
		target = os.Stdout
		if term.IsTerminal(syscall.Stdout) {
			die("Error: Binary output can mess up your terminal. Please use -o <FILE>.")
		}
	} else if outputFilename == "-" {
		f, err := ioutil.TempFile("", "screenshot.*.png")
		if err != nil {
			die(err)
		}
		target = f
		targetFileName = f.Name()
	} else {
		f, err := os.OpenFile(outputFilename, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			die(err)
		}
		target = f
		targetFileName = f.Name()
	}
	defer target.Close()

	var cancel func()
	ctx := context.Background()

	if *webSocketDebuggerUrl != "" {
		if verbose {
			fmt.Fprintln(os.Stderr, "websocket", *webSocketDebuggerUrl)
		}
		ctx, cancel = chromedp.NewRemoteAllocator(ctx, *webSocketDebuggerUrl)
		defer cancel()
	}

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	var buf []byte
	err := chromedp.Run(
		ctx,
		chromedp.Emulate(deviceToUse),
		chromedp.Navigate(inputUrl),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// https://github.com/chromedp/examples/blob/master/screenshot/main.go
			_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}
			if fullPage {
				width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))
				err = emulation.SetDeviceMetricsOverride(width, height, 1, false).Do(ctx)
				if err != nil {
					return err
				}
				buf, err = page.CaptureScreenshot().WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  deviceToUse.info.Scale,
				}).Do(ctx)
			} else {
				buf, err = page.CaptureScreenshot().Do(ctx)
			}
			return err
		}),
	)
	if err != nil {
		die(err)
	}
	_, err = target.Write(buf)
	if err != nil {
		die(err)
	}
	if targetFileName != "" {
		fmt.Println(targetFileName)
		if openFile {
			open.Start(targetFileName)
		}
	}
}
