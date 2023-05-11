package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type flagInfo struct {
	Silence string
	File    string
}

type result struct {
	State      int
	StatusCode int64
	Url        string
	Title      string
	Length     int64
}

const (
	Survival = 1
	Reject   = -1
)

// rand UA header from https://github.com/boy-hack/goWhatweb/blob/master/fetch/fetch.go
func getRandUa() string {
	UserAgents := []string{"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1; AcooBrowser; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.0; Acoo Browser; SLCC1; .NET CLR 2.0.50727; Media Center PC 5.0; .NET CLR 3.0.04506)",
		"Mozilla/4.0 (compatible; MSIE 7.0; AOL 9.5; AOLBuild 4337.35; Windows NT 5.1; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		"Mozilla/5.0 (Windows; U; MSIE 9.0; Windows NT 9.0; en-US)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Win64; x64; Trident/5.0; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 2.0.50727; Media Center PC 6.0)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 1.0.3705; .NET CLR 1.1.4322)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.2; .NET CLR 1.1.4322; .NET CLR 2.0.50727; InfoPath.2; .NET CLR 3.0.04506.30)",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; zh-CN) AppleWebKit/523.15 (KHTML, like Gecko, Safari/419.3) Arora/0.3 (Change: 287 c9dfb30)",
		"Mozilla/5.0 (X11; U; Linux; en-US) AppleWebKit/527+ (KHTML, like Gecko, Safari/419.3) Arora/0.6",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.8.1.2pre) Gecko/20070215 K-Ninja/2.1.1",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; zh-CN; rv:1.9) Gecko/20080705 Firefox/3.0 Kapiko/3.0",
		"Mozilla/5.0 (X11; Linux i686; U;) Gecko/20070322 Kazehakase/0.4.5",
		"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.8) Gecko Fedora/1.9.0.8-1.fc10 Kazehakase/0.5.6",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_3) AppleWebKit/535.20 (KHTML, like Gecko) Chrome/19.0.1036.7 Safari/535.20",
		"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; fr) Presto/2.9.168 Version/11.52"}
	length := len(UserAgents)
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(length)
	return UserAgents[index]
}

func reqScan(url string) []result {
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("User-Agent", getRandUa())
	resp, err := http.DefaultClient.Do(request)
	var results []result
	if err != nil {
		r := result{
			State:      Reject,
			StatusCode: -1,
			Url:        url,
			Title:      "",
			Length:     0,
		}
		results = append(results, r)
		return results
	}
	defer resp.Body.Close()

	var contentLength int64
	//由于 io.TeeReader() 函数将数据同时写入缓冲区和输出流中，一旦我们使用 goquery.NewDocumentFromReader(teeReader) 读取缓冲区中的数据后，teeReader 指向的是文件的末尾，所以再次读取其内容时，无法读取到任何内容。
	//如果需要在已读取数据后仍然能够访问缓冲区内容，可以将 buf 的内容存储到另一个变量中
	buf := new(bytes.Buffer)
	teeReader := io.TeeReader(resp.Body, buf)
	doc, err := goquery.NewDocumentFromReader(teeReader)
	if err != nil {
		fmt.Println("解析HTML出错：", err)
		return []result{}
	}
	statusCode := resp.StatusCode
	title := doc.Find("title").Text()
	//将 buf 的内容存储到 content 变量中，这样就可以在读取 doc 的内容后仍然访问之前读取的缓冲区中的数据了
	content := buf.Bytes()
	contentLength = int64(len(content))

	r := result{
		State:      Survival,
		StatusCode: int64(statusCode),
		Url:        url,
		Title:      title,
		Length:     contentLength,
	}
	results = append(results, r)
	return results
}

func infoPrint(results []result) {
	for _, res := range results {
		if res.State == 1 {
			if res.StatusCode == 200 {
				fmt.Printf("\u001B[32m[+]StatusCode:%d Url:%s Title:%s Length:%d\u001B[0m\n", res.StatusCode, res.Url, res.Title, res.Length)
			} else if res.StatusCode == 404 {
				fmt.Printf("\u001B[33m[+]StatusCode:%d Url:%s Title:%s Length:%d\u001B[0m\n", res.StatusCode, res.Url, res.Title, res.Length)

			} else if res.StatusCode == 403 {
				fmt.Printf("\u001B[34m[+]StatusCode:%d Url:%s Title:%s Length:%d\u001B[0m\n", res.StatusCode, res.Url, res.Title, res.Length)

			} else if res.StatusCode == 500 {
				fmt.Printf("\u001B[35m[+]StatusCode:%d Url:%s Title:%s Length:%d\u001B[0m\n", res.StatusCode, res.Url, res.Title, res.Length)
			} else {
				fmt.Printf("\u001B[37m[+]StatusCode:%d Url:%s Title:%s Length:%d\u001B[0m\n", res.StatusCode, res.Url, res.Title, res.Length)
			}
			///fmt.Printf("[+] StatusCode:%d Url:%s Title:%s Length:%d\n", res.StatusCode, res.Url, res.Title, res.Length)
		} else {
			fmt.Printf("\u001b[31m[-] Url:%s Reject\u001b[0m\n", res.Url)
			//fmt.Printf("[-] Url:%s Reject\n", res.Url)
		}
	}
}

func urlCheck(inputFile, outputFile string) error {
	in, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			fmt.Fprintln(out, line)
		} else {
			fmt.Fprintln(out, "http://"+line)
			fmt.Fprintln(out, "https://"+line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func startScan(file string, cmd *flagInfo) {
	readFile, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer readFile.Close()

	resultCh := make(chan []result)
	wg := sync.WaitGroup{}
	scanner := bufio.NewScanner(readFile)
	for scanner.Scan() {
		url := scanner.Text()
		wg.Add(1)
		go func(u string) {
			resultCh <- reqScan(u)
			wg.Done()
		}(url)
	}
	go func() {
		wg.Wait()
		close(resultCh)
	}()
	filename := "out.csv"
	for req := range resultCh {
		if cmd.File != "" {
			infoPrint(req)
			resultMap := cateByStatusCode(req)
			err := write2Csv(filename, resultMap)
			if err != nil {
				return
			}
		} else if cmd.Silence != "" {
			resultMap := cateByStatusCode(req)
			err := write2Csv(filename, resultMap)
			if err != nil {
				return
			}
		}
	}
}

func Flag(cmd *flagInfo) {
	flag.StringVar(&cmd.File, "f", "", "Batch URL Detection, for example: urls.txt")
	flag.StringVar(&cmd.Silence, "m", "", "Batch URL Detection(silence ouput), for example: urls.txt")
	flag.Parse()
}

func cateByStatusCode(results []result) map[int64][]result {
	resultMap := make(map[int64][]result)
	for _, r := range results {
		resultMap[r.StatusCode] = append(resultMap[r.StatusCode], r)
	}
	return resultMap
}

func write2Csv(filename string, resultMap map[int64][]result) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err != nil {
		return err
	}
	for _, results := range resultMap {
		for _, r := range results {
			err := writer.Write([]string{strconv.FormatInt(r.StatusCode, 10), r.Url, r.Title, strconv.FormatInt(r.Length, 10)})
			if err != nil {
				return err
			}
		}
	}
	writer.Flush()
	return nil
}

func writeCsvHeader(filename string) {
	csvFile, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer csvFile.Close()
	writer := csv.NewWriter(csvFile)
	err = writer.Write([]string{"StatusCode", "URL", "Title", "Length"})
	writer.Flush()
}

func dealCsv(filename string, output string) {
	csvFile, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	// 确定要排序的列的索引
	columnIndex := -1
	for i, colName := range records[0] {
		if colName == "StatusCode" {
			columnIndex = i
			break
		}
	}
	if columnIndex == -1 {
		panic("column StatusCode not found")
	}

	sortFunc := func(i, j int) bool {
		return records[i+1][columnIndex] < records[j+1][columnIndex]
	}

	sort.SliceStable(records[1:], sortFunc)

	outfile, err := os.Create(output)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	writer := csv.NewWriter(outfile)
	defer writer.Flush()

	for _, row := range records {
		err := writer.Write(row)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("结果保存在文件:%s\n", output)
}

func main() {
	var cmd flagInfo
	Flag(&cmd)
	output := "output.txt"
	filename := "out.csv"
	writeCsvHeader(filename)
	var outCsv string
	if cmd.File != "" {
		err := urlCheck(cmd.File, output)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("请输入导出结果的文件名(xxx.csv):")
		fmt.Scanln(&outCsv)
		startScan(output, &cmd)
		dealCsv(filename, outCsv)
	} else if cmd.Silence != "" {
		err := urlCheck(cmd.Silence, output)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("请输入导出结果的文件名(xxx.csv):")
		fmt.Scanln(&outCsv)
		fmt.Println("[+]Detecting...")
		startScan(output, &cmd)
		dealCsv(filename, outCsv)
	} else {
		fmt.Println("Please provide a argument (-h/-f/-m)")
	}
}
