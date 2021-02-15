package main

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/chromedp/chromedp/device"
)

var (
	devices Devices
)

type (
	Device struct {
		id   int
		info device.Info
	}

	Devices []Device
)

func init() {
	t := reflect.TypeOf(device.Reset)
	defer func() { recover() }() // exit loop if "index-out-of-range"
	for i := 1; ; i++ {
		infoType := reflect.New(t)
		infoType.Elem().SetInt(int64(i))
		devices = append(devices, Device{
			id:   i,
			info: infoType.MethodByName("Device").Interface().(func() device.Info)(),
		})
	}
}

func (d Device) Device() device.Info {
	return d.info
}

func (d *Device) Set(i string) error { // flag.Value
	_d := devices.NewDeviceByName(i)
	if _d == nil {
		n, _ := strconv.Atoi(i)
		_d = devices.NewDeviceById(n)
	}
	if _d == nil {
		query, _ := url.ParseQuery(i)
		for key := range query {
			value := query.Get(key)
			switch strings.ToLower(key) {
			case "name":
				d.info.Name = value
			case "useragent":
				d.info.UserAgent = value
			case "width":
				n, _ := strconv.Atoi(value)
				d.info.Width = int64(n)
			case "height":
				n, _ := strconv.Atoi(value)
				d.info.Height = int64(n)
			case "scale":
				n, _ := strconv.ParseFloat(value, 64)
				d.info.Scale = n
			case "landscape":
				d.info.Landscape = value == "true"
			case "mobile":
				d.info.Mobile = value == "true"
			case "touch":
				d.info.Touch = value == "true"
			}
		}
	} else {
		d.updateFrom(_d)
	}
	return nil
}

func (d Device) String() string { // flag.Value
	return ""
}

func (d Device) MultilineString() string {
	return "Name: " + d.info.Name + "\n" +
		"User-Agent: " + d.info.UserAgent + "\n" +
		"Width: " + strconv.Itoa(int(d.info.Width)) + "\n" +
		"Height: " + strconv.Itoa(int(d.info.Height)) + "\n" +
		"Scale: " + strconv.FormatFloat(d.info.Scale, 'f', -1, 64) + "\n" +
		"Landscape: " + strconv.FormatBool(d.info.Landscape) + "\n" +
		"Mobile: " + strconv.FormatBool(d.info.Mobile) + "\n" +
		"Touch: " + strconv.FormatBool(d.info.Touch)
}

func (d Device) MultilineStringIndent(n int) string {
	return regexp.MustCompile("(?m)^").ReplaceAllString(d.MultilineString(), strings.Repeat(" ", n))
}

func (n *Device) updateFrom(o *Device) {
	n.info.Name = o.info.Name
	n.info.UserAgent = o.info.UserAgent
	n.info.Width = o.info.Width
	n.info.Height = o.info.Height
	n.info.Scale = o.info.Scale
	n.info.Landscape = o.info.Landscape
	n.info.Mobile = o.info.Mobile
	n.info.Touch = o.info.Touch
}

func (dd Devices) NewDeviceByName(name string) *Device {
	for _, _d := range dd {
		if _d.info.Name == name {
			d := &Device{}
			d.updateFrom(&_d)
			return d
		}
	}
	return nil
}

func (dd Devices) NewDeviceById(i int) *Device {
	for _, _d := range dd {
		if _d.id == i {
			d := &Device{}
			d.updateFrom(&_d)
			return d
		}
	}
	return nil
}

func (dd Devices) String() string {
	ss := []string{}
	for _, d := range dd {
		s := fmt.Sprintf("%2d  %-40s  %-10s  %0.2fx", d.id, d.info.Name, fmt.Sprintf("%dx%d", d.info.Width, d.info.Height), d.info.Scale)
		ss = append(ss, s)
	}
	return strings.Join(ss, "\n")
}
