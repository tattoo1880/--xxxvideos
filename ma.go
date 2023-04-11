package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

func HandleErrors(err error, why string) {
	if err != nil {
		fmt.Println(why, err)
		return
	}
}

// func parseUrl(url string) string {
// 	resp1, err := http.Get(url)
// 	HandleErrors(err, "获取相应失败")
// 	defer resp1.Body.Close()
// 	code := resp1.StatusCode
// 	fmt.Println("响应的返回代码为: ", code)
// 	body, err := ioutil.ReadAll(resp1.Body)
// 	HandleErrors(err, "获取源代码失败:")
// 	content := string(body)
// 	fmt.Println(content)
// 	return content
// }

func parseUrl(baseUrl string) string {
	client := &http.Client{}
	url := baseUrl
	req, err := http.NewRequest("GET", url, nil)
	HandleErrors(err, "NewRequest")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	HandleErrors(err, "client.do")
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	HandleErrors(err, "readall")
	content := string(body)
	// fmt.Println(content)
	ioutil.WriteFile("./course.txt", body, 0666)
	return content
}

func parseHls(content string) string {
	html := strings.Replace(content, "\n", "", -1)
	re1 := regexp.MustCompile(`html5player.setVideoHLS\('.*?\.m3u8'\)`)
	rec1 := re1.FindString(html)
	// fmt.Println(rec1)
	relink := regexp.MustCompile(`https\:.*?\.m3u8`)
	hlsUrl := relink.FindString(rec1)
	// fmt.Println(hlsUrl)
	hlsUrl = strings.TrimSpace(hlsUrl)
	return hlsUrl
}

func parseM3u8Url(hlsUrl string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", hlsUrl, nil)
	HandleErrors(err, "m3u8请求")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	HandleErrors(err, "获取m3u8响应")
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	HandleErrors(err, "读取响应数据失败")
	ioutil.WriteFile("./m3u8.txt", body, 0644)
	// fmt.Println(string(body))
	downloadBaseUrl := strings.Replace(hlsUrl, "hls.m3u8", "", -1)
	// fmt.Println(downloadBaseUrl)
	// reader, err := ioutil.ReadFile("./m3u8.txt")
	// HandleErrors(err, "读取m3u8列表")
	// fmt.Println(string(reader))
	f1, err := os.Open("./m3u8.txt")
	HandleErrors(err, "打开文件失败")
	defer f1.Close()
	br := bufio.NewReader(f1)
	list := []string{}
	for {
		line, _, err := br.ReadLine()
		lin := strings.TrimSpace(string(line))
		if strings.Contains(lin, "#") {
			continue
		}
		list = append(list, lin)
		if err == io.EOF {
			break
		}
	}
	fmt.Println(list[:len(list)-1])
	for k, v := range list[:len(list)-1] {
		fmt.Println("序号:", k+1, "-----", "分辨率:", v)
	}
	fmt.Println("请选择您要的选择分辨率的序号:")
	num := 0
	fmt.Scanln(&num)
	resUrl := downloadBaseUrl + list[num-1]
	fmt.Println(resUrl)
	return resUrl

}

func parseTs(url string) (nameList, tslit []string) {
	r, e := http.Get(url)
	HandleErrors(e, "get方法")
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	HandleErrors(err, "readall")
	// fmt.Println(string(body))
	ioutil.WriteFile("./list.txt", body, 0666)
	f, e := os.Open("./list.txt")
	defer f.Close()
	HandleErrors(e, "open list txt:")
	tslist := []string{}
	br := bufio.NewReader(f)
	for {
		str, err := br.ReadString('\n')
		lin := strings.TrimSpace(str)
		if strings.Contains(lin, "#") {
			continue
		}
		tslist = append(tslist, lin)
		if err == io.EOF {
			break
		}
	}
	// fmt.Println(tslist)
	lis := strings.Split(url, "/")
	// fmt.Println(lis)
	newUrl := "http:/"
	for i := 1; i < len(lis)-1; i++ {
		newUrl += lis[i] + "/"
	}
	// fmt.Println(newUrl)
	tslist = tslist[:len(tslist)-1]
	finalList := []string{}
	for _, v := range tslist {
		finalList = append(finalList, newUrl+v)
		// ioutil.WriteFile("./videos/file_list.txt",v,0666)
	}
	// fmt.Println(finalList)
	return tslist, finalList
}

func Dl(url, filename string) {
	defer wg.Done()
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	HandleErrors(err, "获取ts请求失败")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	HandleErrors(err, "ts获取失败")
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	HandleErrors(err, "ReadAll")
	fileName := "./videos/" + filename
	ioutil.WriteFile(fileName, body, 0666)
	fmt.Println(filename, "下载完成")
}

func fileList(namelist []string) {

	filePath := "./videos/file_list.txt"
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	HandleErrors(err, "openfile")
	defer file.Close()
	writer := bufio.NewWriter(file)
	for i := 0; i < len(namelist); i++ {
		writer.WriteString("file " + "'" + namelist[i] + "'" + "\n")
	}
	writer.Flush()
}

func concatTS(url string) {
	name := reName(url)
	cmd := exec.Command("ffmpeg", "-f", "concat", "-i", "./videos/file_list.txt", "-c", "copy", name)
	err := cmd.Run()
	HandleErrors(err, "cmd.run")

	cmd2 := exec.Command("mv", name, "/gd/gd/movie/")
	err2 := cmd2.Run()
	HandleErrors(err2, "cmd2.run")
	fmt.Println("您下载的文件名为: ", name)
}

func deleteDir() {
	dir, err := ioutil.ReadDir("./videos")
	HandleErrors(err, "readdir")
	for _, d := range dir {
		os.RemoveAll(path.Join([]string{"videos", d.Name()}...))
	}
}

func reName(url string) string {
	n := strings.Split(url, "/")
	n1 := n[len(n)-1:]
	newname := n1[0] + ".mp4"

	return newname

}

//productors

func productors(kl []string, in chan<- string) {
	defer wg.Done()
	for _, url := range kl {
		in <- url
		time.Sleep(time.Millisecond * 10)
	}
	close(in)
	// return in
}

func customers(out <-chan string) {
	defer wg.Done()
	for v := range out {

		url := v
		pp := strings.Split(v, "/")
		name := pp[len(pp)-1:]
		rename := name[0]
		wg.Add(1)
		go Dl(url, rename)
	}
}

func main() {
	var mainURL string = ""
	fmt.Println("请输入您要下载的视频地址")
	fmt.Scanln(&mainURL)

	content := parseUrl(mainURL)
	hls := parseHls(content)
	finalURL := parseM3u8Url(hls)
	nameList, finaUrl := parseTs(finalURL)
	// // wg.Add(len(finaUrl))
	// for _, v := range finaUrl {
	// 	url := v
	// 	pp := strings.Split(v, "/")
	// 	name := pp[len(pp)-1:]
	// 	rename := name[0]
	// 	// fmt.Println(url)
	// 	// fmt.Println(name)
	// 	go Dl(url, rename)
	// 	// break
	// }
	// wg.Wait()
	var ch chan string = make(chan string, 25)
	wg.Add(2)
	go productors(finaUrl, ch)
	go customers(ch)
	wg.Wait()
	fmt.Println("下载完成===========================")
	fileList(nameList)
	concatTS(mainURL)
	// time.Sleep(time.Second * 5)

	deleteDir()
	fmt.Println("清空文件夹完成")
}
