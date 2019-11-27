package launch

import (
	"tesou.io/platform/foot-parent/foot-core/module/core/service"
	"tesou.io/platform/foot-parent/foot-core/module/match/entity"
	"tesou.io/platform/foot-parent/foot-spider/module/win007"
	"tesou.io/platform/foot-parent/foot-spider/module/win007/proc"
)

/*func main() {
	//执行抓取亚赔数据
	Before_spider_asiaLast()
	Spider_asiaLast()
}*/

func Before_spider_asiaLast(){
	//抓取前清空当前比较表
	opsService := new(service.DBOpsService)
	//指定需要清空的数据表
	opsService.TruncateTable([]string{"t_asia_last"})
}

//查询标识为win007,且欧赔未抓取的配置数据,指定菠菜公司
func Spider_asiaLast() {
	matchLastConfig := new(entity.MatchLastConfig)
	matchLastConfig.S = win007.MODULE_FLAG
	matchLastConfig.AsiaSpided = false
	matchLastConfigs := matchLastConfig.Query()

	processer := proc.GetAsiaLastProcesser()
	processer.MatchLastConfig_list = matchLastConfigs
	processer.Startup()
}
