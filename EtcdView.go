package main

import (
	"context"
	"encoding/json"
	"errors"
	"etcdviewer/utils"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"go.etcd.io/etcd/clientv3"
	"log"
	"strings"
	"time"
)

var client *clientv3.Client
var kvs map[string]string
var keylist []string
var w fyne.Window
var myApp fyne.App
var selectEntry *widget.SelectEntry

const selectnum = 300
const etcdtimeout = 15
const textsize = 20

func RefreshData() {
	kvs = make(map[string]string)
	keylist = keylist[:0]
	if client == nil {
		fmt.Println("abnormal etcd client")
		return
	}

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Start()
	size := fyne.NewSize(50, 10)
	progressBar.Resize(size)
	progressBar.Start()
	loadProgress := myApp.NewWindow("数据加载中")

	loadProgress.SetContent(progressBar)
	loadProgress.Resize(size)
	go func() {
		loadProgress.FullScreen()
		loadProgress.CenterOnScreen()
		loadProgress.Show()
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		resp, err := client.Get(ctx, "", clientv3.WithPrefix())
		cancel()
		if err != nil {
			msg := "etcd connect err"
			fmt.Println(msg)
			myDialog := dialog.NewError(errors.New(msg), w)
			myDialog.Show()
			return
		}
		if kvs == nil {
			kvs = make(map[string]string, len(resp.Kvs))
		}
		for _, kv := range resp.Kvs {
			kvs[string(kv.Key)] = string(kv.Value)
			keylist = append(keylist, string(kv.Key))
		}
		fmt.Println("kvs size:", len(resp.Kvs))
		if len(keylist) != 0 {
			selectEntry.SetOptions(func() []string {
				if len(keylist) > selectnum {
					return keylist[:selectnum]
				} else {
					return keylist
				}
			}())
		}
		loadProgress.Close()
	}()

}

func PrettyJsonStr(raw []byte) (pretty []byte, err error) {
	var data map[string]interface{}
	err = json.Unmarshal(raw, &data)
	if err != nil {
		return nil, err
	}

	// Marshal map as pretty JSON string
	pretty, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}
	return pretty, nil
}

func Init(endpoints []string) bool {
	var eps []string
	for _, ep := range endpoints {
		vec := strings.Split(ep, ":")
		if len(vec) != 2 {
			continue
		}
		if !utils.CheckNet(vec[0], vec[1]) {
			continue
		} else {
			eps = append(eps, ep)
		}
	}
	if len(eps) == 0 {
		return false
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   eps,
		DialTimeout: etcdtimeout * time.Second,
	})
	if err != nil {
		fmt.Println(err)
		msg := "etcd connect err"
		fmt.Println(msg)
		myDialog := dialog.NewError(errors.New(msg), w)
		myDialog.Show()
		return false
	}
	client = cli
	RefreshData()
	return true
}

func Release() {
	if client != nil {
		err := client.Close()
		if err != nil {
			return
		}
	}
}

type EnterShort struct {
}

func (e EnterShort) ShortcutName() string {
	return string(fyne.KeyEnter)
}

func EtcdView() {
	//err := os.Setenv("FYNE_SCALE", "1.2")
	//if err != nil {
	//	return
	//}
	myApp = app.New()
	t := &TestTheme{}
	t.SetFonts("./simhei.ttf", "")
	myApp.Settings().SetTheme(t)

	w = myApp.NewWindow("Etcd可视化工具")
	w.Resize(fyne.NewSize(500, 300))
	endpoints := []string{"10.242.100.33:2379"}
	if !Init(endpoints) {
		msg := "etcd connect err"
		fmt.Println(msg)
		myDialog := dialog.NewError(errors.New(msg), w)
		myDialog.Show()
	}
	defer Release()

	selectEntry = &widget.SelectEntry{}
	etcd := widget.NewEntry()
	etcd.SetText("10.242.100.33:2379")
	etcd.OnSubmitted = func(s string) {
		endpoints = []string{etcd.Text}
		fmt.Println("etcd change to ", etcd.Text)
		if !Init([]string{s}) {
			msg := "etcd connect err"
			fmt.Println(msg)
			myDialog := dialog.NewError(errors.New(msg), w)
			myDialog.Show()
			return
		}
		if selectEntry != nil {
			if len(keylist) > selectnum {
				msg := "too many data"
				fmt.Println(msg)
				myDialog := dialog.NewInformation("Notify", msg, w)
				myDialog.Show()
				selectEntry.SetOptions(keylist[:selectnum])
			} else {
				selectEntry.SetOptions(keylist)
			}
		}
	}
	keyEntry := widget.NewEntry()
	//valueEntry := widget.NewEntry()
	prefixEntry := widget.NewEntry()
	prefixEntry.SetText("Taurus/SR/TaskParam")
	label := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{
		Bold:      false,
		Italic:    true,
		Monospace: true,
		Symbol:    false,
		TabWidth:  4,
	})
	labelScroll := container.NewVScroll(label)
	labelScroll.SetMinSize(fyne.NewSize(800, 700))
	//labelBox := container.NewWithoutLayout(labelContainer)
	//addButton := widget.NewButton("put", func() {
	//	key := keyEntry.Text
	//	value := valueEntry.Text
	//	if key == "" || value == "" {
	//		return
	//	}
	//	cli, err := clientv3.New(clientv3.Config{
	//		Endpoints:   endpoints,
	//		DialTimeout: etcdtimeout * time.Second,
	//	})
	//	if err != nil {
	//		panic(err)
	//	}
	//	defer cli.Close()
	//
	//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//	_, err = cli.Put(ctx, key, value)
	//	cancel()
	//	if err != nil {
	//		panic(err)
	//	}
	//	keyEntry.SetText("")
	//	valueEntry.SetText("")
	//})

	selectEntry = widget.NewSelectEntry(func() []string {
		if len(keylist) >= selectnum {
			return keylist[:selectnum]
		} else {
			return keylist
		}
	}())
	selectEntry.OnChanged = func(value string) {
		v := kvs[value]
		log.Printf("%s:\n%s\n", value, v)
		keyEntry.SetText(value)
		pretty, _ := PrettyJsonStr([]byte(v))
		label.SetText(string(pretty))
	}

	scrolledContainer := container.NewVScroll(selectEntry)
	//scrolledContainer.Resize(fyne.NewSize(100, 50))

	keyEntry.OnSubmitted = func(s string) {

		key := keyEntry.Text
		selectEntry.Text = ""
		if key == "" {
			msg := "empty key"
			fmt.Println(msg)
			myDialog := dialog.NewError(errors.New(msg), w)
			myDialog.Show()
			return
		}
		cli, err := clientv3.New(clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: etcdtimeout * time.Second,
		})
		if err != nil {
			msg := "etcd connect err"
			fmt.Println(msg)
			myDialog := dialog.NewError(errors.New(msg), w)
			myDialog.Show()
			return
		}
		defer cli.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		resp, err := cli.Get(ctx, key)
		cancel()
		if err != nil {
			panic(err)
		}
		if len(resp.Kvs) == 0 {
			fmt.Println("键不存在")
			label.SetText("")
			return
		}
		for _, kv := range resp.Kvs {
			//fmt.Printf("键：%s，值：%s\n", kv.Key, kv.Value)
			selectEntry.SetText(string(kv.Key))
			pretty, _ := PrettyJsonStr(kv.Value)
			label.SetText(string(pretty))
		}
	}
	//deleteButton := widget.NewButton("delete", func() {
	//	key := keyEntry.Text
	//	if key == "" {
	//		return
	//	}
	//	cli, err := clientv3.New(clientv3.Config{
	//		Endpoints:   endpoints,
	//		DialTimeout: etcdtimeout * time.Second,
	//	})
	//	if err != nil {
	//		panic(err)
	//	}
	//	defer cli.Close()
	//
	//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//	_, err = cli.Delete(ctx, key)
	//	cancel()
	//	if err != nil {
	//		panic(err)
	//	}
	//	keyEntry.SetText("")
	//	valueEntry.SetText("")
	//})

	//listButton := widget.NewButton("list", func() {
	prefixEntry.OnSubmitted = func(s string) {
		selectEntry.Text = ""
		cli, err := clientv3.New(clientv3.Config{
			Endpoints:   []string{etcd.Text},
			DialTimeout: etcdtimeout * time.Second,
		})
		if err != nil {
			msg := "etcd connect err"
			fmt.Println(msg)
			myDialog := dialog.NewError(errors.New(msg), w)
			myDialog.Show()
			return
		}
		defer cli.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		resp, err := cli.Get(ctx, prefixEntry.Text, clientv3.WithPrefix())
		cancel()
		if err != nil {
			log.Println(err)
			return
		}
		if len(kvs) == 0 {
			return
		}
		fmt.Println("kvs size:", len(resp.Kvs))
		if len(resp.Kvs) == 0 {
			return
		}
		options := []string{}
		for _, kv := range resp.Kvs {
			//fmt.Printf("键：%s，值：%s\n", kv.Key, kv.Value)
			options = append(options, string(kv.Key))
			kvs[string(kv.Key)] = string(kv.Value)
		}
		if len(options) > selectnum {
			msg := "too many data"
			fmt.Println(msg)
			myDialog := dialog.NewInformation("Notify", msg, w)
			myDialog.Show()
			selectEntry.SetOptions(options[:selectnum])
		} else {
			selectEntry.SetOptions(options)
		}
		selectEntry.SetText(string(options[0]))
	}

	w.SetContent(container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewLabel("Etcd:"),
			etcd,
			widget.NewLabel("Key:"),
			keyEntry,
			//widget.NewLabel("value:"),
			//valueEntry,
			widget.NewLabel("Prefix:"),
			prefixEntry,
		),
		scrolledContainer,
		labelScroll,
		//labelContainer,
		//labelBox,
	))
	w.CenterOnScreen()
	w.ShowAndRun()
}
