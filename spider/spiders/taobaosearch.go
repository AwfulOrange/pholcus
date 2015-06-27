package spiders

// 基础包
import (
	// "github.com/PuerkitoBio/goquery" //DOM解析
	"github.com/henrylee2cn/pholcus/crawl/downloader/context" //必需
	"github.com/henrylee2cn/pholcus/reporter"                 //信息输出
	. "github.com/henrylee2cn/pholcus/spider"                 //必需
	// . "github.com/henrylee2cn/pholcus/spider/common"          //选用
)

// 设置header包
import (
// "net/http" //http.Header
)

// 编码包
import (
	// "encoding/xml"
	"encoding/json"
)

// 字符串处理包
import (
	"regexp"
	"strconv"
	"strings"
)

// 其他包
import (
// "fmt"
// "math"
)

func init() {
	TaobaoSearch.AddMenu()
}

var TaobaoSearch = &Spider{
	Name:        "淘宝搜索",
	Description: "淘宝天猫搜索结果 [s.taobao.com]",
	Keyword:     CAN_ADD,
	// Pausetime: [2]uint{uint(3000), uint(1000)},
	// Optional: &Optional{},
	RuleTree: &RuleTree{
		// Spread: []string{},
		Root: func(self *Spider) {
			self.AidRule("生成请求", map[string]interface{}{"loop": [2]int{0, 1}, "rule": "生成请求"})
		},

		Nodes: map[string]*Rule{

			"生成请求": &Rule{
				AidFunc: func(self *Spider, aid map[string]interface{}) interface{} {
					self.LoopAddQueue(
						aid["loop"].([2]int),
						func(i int) []string {
							return []string{"http://s.taobao.com/search?q=" + self.GetKeyword() + "&ie=utf8&cps=yes&app=vproduct&cd=false&v=auction&tab=all&vlist=1&bcoffset=1&s=" + strconv.Itoa(i*44)}
						},
						map[string]interface{}{
							"rule": aid["rule"].(string),
						},
					)
					return nil
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()
					src := query.Find("script").Text()
					if strings.Contains(src, "抱歉！没有找到与") {
						reporter.Log.Println("搜索结果为 0 ！")
						return
					}

					re, _ := regexp.Compile(`(?U)"totalCount":[\d]+}`)
					total := re.FindString(src)
					re, _ = regexp.Compile(`[\d]+`)
					total = re.FindString(total)
					totalCount, _ := strconv.Atoi(total)

					maxPage := (totalCount - 4) / 44
					if (totalCount-4)%44 > 0 {
						maxPage++
					}

					if self.GetMaxPage() > maxPage || self.GetMaxPage() == 0 {
						self.SetMaxPage(maxPage)
					} else if self.GetMaxPage() == 0 {
						reporter.Log.Printf("[消息提示：| 任务：%v | 关键词：%v | 规则：%v] 没有抓取到任何数据！!!\n", self.GetName(), self.GetKeyword(), resp.GetRuleName())
						return
					}

					reporter.Log.Printf("********************** 淘宝关键词 [%v] 的搜索结果共有 %v 页，计划抓取 %v 页 **********************", self.GetKeyword(), maxPage, self.GetMaxPage())
					// 调用指定规则下辅助函数
					self.AidRule("生成请求", map[string]interface{}{"loop": [2]int{1, self.GetMaxPage()}, "rule": "搜索结果"})
					// 用指定规则解析响应流
					self.CallRule("搜索结果", resp)
				},
			},

			"搜索结果": &Rule{
				//注意：有无字段语义和是否输出数据必须保持一致
				OutFeild: []string{
					"标题",
					"价格",
					"销量",
					"店铺",
					"发货地",
					"链接",
				},
				ParseFunc: func(self *Spider, resp *context.Response) {
					query := resp.GetDom()
					re, _ := regexp.Compile(`"auctions".*,"recommendAuctions"`)
					src := query.Find("script").Text()

					src = re.FindString(src)

					re, _ = regexp.Compile(`"auctions":`)
					src = re.ReplaceAllString(src, "")

					re, _ = regexp.Compile(`,"recommendAuctions"`)
					src = re.ReplaceAllString(src, "")

					re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
					// src = re.ReplaceAllStringFunc(src, strings.ToLower)
					src = re.ReplaceAllString(src, " ")

					src = strings.Trim(src, " \t\n")

					infos := []map[string]interface{}{}

					err := json.Unmarshal([]byte(src), &infos)

					if err != nil {
						reporter.Log.Printf("error is %v\n", err)
						return
					} else {
						for _, info := range infos {

							// 结果存入Response中转
							resp.AddItem(map[string]interface{}{
								self.GetOutFeild(resp, 0): info["raw_title"],
								self.GetOutFeild(resp, 1): info["view_price"],
								self.GetOutFeild(resp, 2): info["view_sales"],
								self.GetOutFeild(resp, 3): info["nick"],
								self.GetOutFeild(resp, 4): info["item_loc"],
								self.GetOutFeild(resp, 5): info["detail_url"],
							})
						}
					}
				},
			},
		},
	},
}
