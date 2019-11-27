package proc

import (
	"log"
	"opensource.io/go_spider/core/common/page"
	"opensource.io/go_spider/core/pipeline"
	"opensource.io/go_spider/core/spider"
	"reflect"
	"strconv"
	"strings"
	"tesou.io/platform/foot-parent/foot-core/common/base/service/mysql"
	entity2 "tesou.io/platform/foot-parent/foot-core/module/elem/entity"
	"tesou.io/platform/foot-parent/foot-core/module/match/entity"
	"tesou.io/platform/foot-parent/foot-spider/module/win007"
	"time"
)

type MatchPageProcesser struct {
	//抓取的url
	MatchLast_url string
}

func GetMatchPageProcesser() *MatchPageProcesser {
	return &MatchPageProcesser{}
}

var (
	//联赛数据
	league_list           = make([]*entity2.League, 0)
	win007Id_leagueId_map = make(map[string]string)
	//比赛数据
	matchLast_list = make([]*entity.MatchLast, 0)
)

func init() {

}

func (this *MatchPageProcesser) Startup() {
	if this.MatchLast_url == "" {
		this.MatchLast_url = "http://m.win007.com/phone/Schedule_0_0.txt"
	}

	newSpider := spider.NewSpider(GetMatchPageProcesser(), "MatchPageProcesser")
	newSpider = newSpider.AddUrl(this.MatchLast_url, "text")
	newSpider = newSpider.AddPipeline(pipeline.NewPipelineConsole())
	newSpider.SetThreadnum(1).Run()
}

func (this *MatchPageProcesser) Process(p *page.Page) {
	request := p.GetRequest()
	if !p.IsSucc() {
		log.Println("URL:,", request.Url, p.Errormsg())
		return
	}

	rawText := p.GetBodyStr()
	if rawText == "" {
		log.Println("URL:,内容为空", request.Url)
		return
	}

	rawText_arr := strings.Split(rawText, "$$")
	if len(rawText_arr) < 2 {
		log.Println("URL:,解析失败,rawTextArr长度为:,小于所必需要的长度3", request.Url, len(rawText_arr))
		return
	}

	flag := this.findParamVal(request.Url)
	var league_str string
	var match_str string
	if flag == "0" {
		league_str = rawText_arr[0]
		match_str = rawText_arr[1]
	} else {
		league_str = rawText_arr[1]
		match_str = rawText_arr[2]
	}

	log.Println("联赛信息:", league_str)
	this.league_process(league_str)
	log.Println("比赛信息:", match_str)
	this.match_process(match_str)
}

func (this *MatchPageProcesser) findParamVal(url string) string {
	paramUrl := strings.Split(url, "_")[2]
	paramArr := strings.Split(paramUrl, ".")
	return paramArr[0]
}

func (this *MatchPageProcesser) league_process(rawText string) {
	league_arr := strings.Split(rawText, "!")

	league_list = make([]*entity2.League, len(league_arr))
	var index int
	for _, v := range league_arr {
		league_info_arr := strings.Split(v, "^")
		if len(league_info_arr) < 3 {
			continue
		}
		name := league_info_arr[0]
		win007Id := league_info_arr[1]

		league := new(entity2.League)
		league.Id = win007Id
		league.Name = name
		//league.Ext = make(map[string]interface{})
		//league.Ext["win007Id"] = win007Id
		win007Id_leagueId_map[win007Id] = league.Id

		level_str := league_info_arr[2]
		level, _ := strconv.Atoi(level_str)
		league.Level = level

		league_list[index] = league
		index++
	}
}

/**
处理比赛信息
*/
func (this *MatchPageProcesser) match_process(rawText string) {
	match_arr := strings.Split(rawText, "!")
	match_len := len(match_arr)
	for i := 0; i < match_len; i++ {
		matchLast := new(entity.MatchLast)
		//matchLast.Ext = make(map[string]interface{})
		//matchLast.Id = bson.NewObjectId().Hex()

		//match_arr[0] is
		//1503881^284^0^20180909170000^^町田泽维亚^水户蜀葵^0^0^0^0^0^0^0^0^0.5^181^^0^2^12^^^0^0^0^0
		match_info_arr := strings.Split(match_arr[i], "^")
		index := 0
		win007Id := match_info_arr[index]
		//matchLast.Ext["win007Id"] = win007Id
		matchLast.Id = win007Id
		index++
		matchLast.LeagueId = win007Id_leagueId_map[match_info_arr[index]]
		index++
		index++
		match_date_str := match_info_arr[index]
		match_date_stamp, _ := time.Parse("20060102150405", match_date_str)
		matchLast.MatchDate = match_date_stamp.Format("2006-01-02 15:04:05")
		index++
		index++
		matchLast.MainTeamId = match_info_arr[index]
		/*		if regexp.MustCompile("^\\d*$").MatchString(dataDate_or_mainTeamName) {
					data_date_timestamp, _ := strconv.ParseInt(dataDate_or_mainTeamName, 10, 64)
					matchLast.DataDate = time.Unix(data_date_timestamp, 0).Format("2006-01-02 15:04:05")
				} else {
					matchLast.MainTeamId = dataDate_or_mainTeamName
				}*/
		index++
		matchLast.GuestTeamId = match_info_arr[index]
		index++
		mainTeamGoals_str := match_info_arr[index]
		mainTeamGoals, _ := strconv.Atoi(mainTeamGoals_str)
		matchLast.MainTeamGoals = mainTeamGoals
		index++
		guestTeamGoals_str := match_info_arr[index]
		guestTeamGoals, _ := strconv.Atoi(guestTeamGoals_str)
		matchLast.GuestTeamGoals = guestTeamGoals

		//最后加入数据中
		matchLast_list = append(matchLast_list, matchLast)
	}

}

func (this *MatchPageProcesser) Finish() {
	log.Println("比赛抓取解析完成,执行入库 \r\n")

	league_list_slice := make([]interface{}, 0)
	leagueConfig_list_slice := make([]interface{}, 0)
	for _, v := range league_list {
		if nil == v {
			continue
		}
	/*	bytes, _ := json.Marshal(v)
		log.Println(string(bytes))*/
		exists := v.FindExistsById()
		if exists {
			continue
		}
		league_list_slice = append(league_list_slice, v)

		leagueConfig := new(entity2.LeagueConfig)
		leagueConfig.S = win007.MODULE_FLAG
		leagueConfig.Sid = v.Id
		leagueConfig.LeagueId = v.Id
		leagueConfig.Id = v.Id
		leagueConfig_list_slice = append(leagueConfig_list_slice, leagueConfig)
	}
	mysql.SaveList(league_list_slice)
	mysql.SaveList(leagueConfig_list_slice)

	matchLast_list_slice := make([]interface{}, 0)
	matchLastConfig_list_slice := make([]interface{}, 0)
	for _, v := range matchLast_list {
		if nil == v {
			continue
		}

		exists := v.FindExists()
		if exists {
			continue
		}
		//v.Id = v.Ext["win007Id"].(string);
		matchLast_list_slice = append(matchLast_list_slice, v)

		//处理比赛配置信息
		matchLastConfig := new(entity.MatchLastConfig)
		matchLast_elem := reflect.ValueOf(v).Elem()
		matchLastConfig.MatchId = matchLast_elem.FieldByName("Id").String()
		matchLastConfig.AsiaSpided = false
		matchLastConfig.EuroSpided = false
		matchLastConfig.S = win007.MODULE_FLAG
		//ext := matchLast_elem.FieldByName("Ext").Interface().(map[string]interface{})
		//matchLastConfig.Sid = ext["win007Id"].(string)
		matchLastConfig.Sid = v.Id
		matchLastConfig.Id = v.Id
		matchLastConfig_list_slice = append(matchLastConfig_list_slice, matchLastConfig)
	}
	mysql.SaveList(matchLast_list_slice)
	mysql.SaveList(matchLastConfig_list_slice)

}
