package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"go.etcd.io/etcd/clientv3"
	"log"
	"time"
)

var client *clientv3.Client
var kvs map[string]string
var keylist []string

func RefreshData() {
	kvs = make(map[string]string)
	keylist = keylist[:0]
	if client == nil {
		fmt.Println("abnormal etcd client")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	resp, err := client.Get(ctx, "", clientv3.WithPrefix())
	cancel()
	if err != nil {
		panic(err)
	}
	if kvs == nil {
		kvs = make(map[string]string, len(resp.Kvs))
	}
	for _, kv := range resp.Kvs {
		kvs[string(kv.Key)] = string(kv.Value)
		keylist = append(keylist, string(kv.Key))
	}
	fmt.Println("kvs size:", len(resp.Kvs))
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
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println(err)
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

func EtcdView() {
	myApp := app.New()
	// 获取当前主题
	//myApp.Settings().SetTheme(&MyTheme{})
	w := myApp.NewWindow("Etcd可视化工具")
	endpoints := []string{"10.242.100.33:2379"}
	if !Init(endpoints) {
		msg := "etcd connect err"
		fmt.Println(msg)
		myDialog := dialog.NewError(errors.New(msg), w)
		myDialog.Show()
	}
	defer Release()

	combo := &widget.Select{}
	etcd := widget.NewEntry()
	etcd.SetText("10.242.100.33:2379")
	etcd.OnChanged = func(s string) {
		fmt.Println("etcd change to ", etcd.Text)
		if !Init([]string{s}) {
			msg := "etcd connect err"
			fmt.Println(msg)
			myDialog := dialog.NewError(errors.New(msg), w)
			myDialog.Show()
		}
		if combo != nil {
			if len(keylist) > 200 {
				combo.Options = keylist[:200]
			} else {
				combo.Options = keylist
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
	//		DialTimeout: 5 * time.Second,
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

	combo = widget.NewSelect(func() []string {
		if len(keylist) >= 200 {
			return keylist[:200]
		} else {
			return keylist
		}
	}(), func(value string) {
		v := kvs[value]
		log.Printf("%s:\n%s\n", value, v)
		keyEntry.SetText(value)
		pretty, _ := PrettyJsonStr([]byte(v))
		label.SetText(string(pretty))
	})
	scrolledContainer := container.NewVScroll(combo)
	//scrolledContainer.Resize(fyne.NewSize(100, 50))

	getButton := widget.NewButton("get", func() {
		key := keyEntry.Text
		if key == "" {
			return
		}
		cli, err := clientv3.New(clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: 5 * time.Second,
		})
		if err != nil {
			panic(err)
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
			return
		}
		for _, kv := range resp.Kvs {
			//fmt.Printf("键：%s，值：%s\n", kv.Key, kv.Value)
			combo.ClearSelected()
			combo.SetSelected(string(kv.Key))
			pretty, _ := PrettyJsonStr(kv.Value)
			label.SetText(string(pretty))
		}
	})
	//deleteButton := widget.NewButton("delete", func() {
	//	key := keyEntry.Text
	//	if key == "" {
	//		return
	//	}
	//	cli, err := clientv3.New(clientv3.Config{
	//		Endpoints:   endpoints,
	//		DialTimeout: 5 * time.Second,
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

	listButton := widget.NewButton("list", func() {
		cli, err := clientv3.New(clientv3.Config{
			Endpoints:   []string{etcd.Text},
			DialTimeout: 5 * time.Second,
		})
		if err != nil {
			panic(err)
		}
		defer cli.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		resp, err := cli.Get(ctx, prefixEntry.Text, clientv3.WithPrefix())
		cancel()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println("kvs size:", len(resp.Kvs))
		options := []string{}
		for _, kv := range resp.Kvs {
			//fmt.Printf("键：%s，值：%s\n", kv.Key, kv.Value)
			options = append(options, string(kv.Key))
		}
		if len(options) > 200 {
			combo.Options = options[:200]
		} else {
			combo.Options = options
		}
	})
	w.SetContent(container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewLabel("etcd:"),
			etcd,
			widget.NewLabel("key:"),
			keyEntry,
			//widget.NewLabel("value:"),
			//valueEntry,
			widget.NewLabel("prefix:"),
			prefixEntry,
		),
		container.NewHBox(
			//addButton,
			getButton,
			//deleteButton,
			listButton,
		),
		scrolledContainer,
		labelScroll,
		//labelContainer,
		//labelBox,
	))
	w.Resize(fyne.NewSize(1000, 800))
	w.ShowAndRun()
}
