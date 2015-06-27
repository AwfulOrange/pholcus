package spiders

// 基础包
import (
	"github.com/PuerkitoBio/goquery"                          //DOM解析
	"github.com/henrylee2cn/pholcus/crawl/downloader/context" //必需
	//"github.com/henrylee2cn/pholcus/reporter"           //信息输出
	. "github.com/henrylee2cn/pholcus/spider" //必需
	// . "github.com/henrylee2cn/pholcus/spider/common" //选用
	//"log"
)

// 设置header包
import (
// "net/http" //http.Header
)

// 编码包
import (
// "encoding/xml"
//"encoding/json"
)

// 字符串处理包
import (
	//"regexp"
	"strconv"
	//	"strings"
)

// 其他包
import (
// "fmt"
// "math"
)

func init() {
	Zolphone.AddMenu()
}

var Zolphone = &Spider{
	Name:        "中关村手机",
	Description: "中关村苹果手机数据 [Auto Page] [bbs.zol.com.cn/sjbbs/d544_p]",
	// Pausetime: [2]uint{uint(3000), uint(1000)},
	// Optional: &Optional{},
	RuleTree: &RuleTree{
		// Spread: []string{},
		Root: func(self *Spider) {
			self.AidRule("生成请求", map[string]interface{}{"loop": [2]int{1, 950}, "rule": "生成请求"})
		},

		Nodes: map[string]*Rule{

			"生成请求": &Rule{
				AidFunc: func(self *Spider, aid map[string]interface{}) interface{} {
					self.LoopAddQueue(
						aid["loop"].([2]int),
						func(i int) []string {
							return []string{"http://bbs.zol.com.cn/sjbbs/d544_p" + strconv.Itoa(i) + ".html#c"}
						},
						map[string]interface{}{
							"rule": aid["rule"].(string),
						},
					)
					return nil
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()
					//					log.Println(query.Find(".edition-topic-inner").Next().Html())
					ss := query.Find("tbody").Find("tr[id]")
					ss.Each(func(i int, goq *goquery.Selection) {
						//						log.Println(goq.Html())
						//						log.Println(i)
						//						log.Println(goq.Attr("id"))

						resp.SetTemp("html", goq)
						self.CallRule("获取结果", resp)

					})

					//					// 调用指定规则下辅助函数
					//					self.AidRule("生成请求", map[string]interface{}{"loop": [2]int{1, totalPage}, "rule": "搜索结果"})
					// 用指定规则解析响应流

				},
			},

			"获取结果": &Rule{
				//注意：有无字段语义和是否输出数据必须保持一致
				OutFeild: []string{
					"机型",
					"链接",
					"主题",
					"发表者",
					"发表时间",
					"总回复",
					"总查看",
					"最后回复者",
					"最后回复时间",
				},
				ParseFunc: func(self *Spider, resp *context.Response) {

					selectObj := resp.GetTemp("html").(*goquery.Selection)
					//					outHtml,_ := selectObj.Find("td").Eq(1).Attr("data-url")
					//url
					outUrls := selectObj.Find("td").Eq(1)
					outUrl, _ := outUrls.Attr("data-url")
					outUrl = "http://bbs.zol.com.cn/" + outUrl
					//					outUrls := selectObj.Find("td").Eq(1)
					//					outUrl := outUrls.Find("div a").Attr("href")

					//title type
					outTitles := selectObj.Find("td").Eq(1)
					outType := outTitles.Find(".iclass a").Text()
					outTitle := outTitles.Find("div a").Text()

					//author stime
					authors := selectObj.Find("td").Eq(2)
					author := authors.Find("a").Text()
					stime := authors.Find("span").Text()

					//reply read
					replys := selectObj.Find("td").Eq(3)
					reply := replys.Find("span").Text()
					read := replys.Find("i").Text()

					//ereply etime
					etimes := selectObj.Find("td").Eq(4)
					ereply := etimes.Find("a").Eq(0).Text()
					etime := etimes.Find("a").Eq(1).Text()

					// 结果存入Response中转
					resp.AddItem(map[string]interface{}{
						self.GetOutFeild(resp, 0): outType,
						self.GetOutFeild(resp, 1): outUrl,
						self.GetOutFeild(resp, 2): outTitle,
						self.GetOutFeild(resp, 3): author,
						self.GetOutFeild(resp, 4): stime,
						self.GetOutFeild(resp, 5): reply,
						self.GetOutFeild(resp, 6): read,
						self.GetOutFeild(resp, 7): ereply,
						self.GetOutFeild(resp, 8): etime,
					})

				},
			},
		},
	},
}
