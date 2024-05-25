package main

import "go.mongodb.org/mongo-driver/bson/primitive"

// 返利表
type rebate_final struct {
	Id      primitive.ObjectID `bson:"_id"`
	Uid     int                `bson:"uid"`
	Price   int                `bson:"price"`
	AddTime int64              `bson:"addTime"`
	OrderNo string             `bson:"orderNo"`
}
type flowing_final struct { // 下注需要添加
	Id    string `bson:"_id"`
	Uid   int    `bson:"uid"`
	Tchip int    `bson:"tchip"` // 当前下注值
}
type rebateItem struct {
	Uid              int     `bson:"uid"`     //本人id
	ChildId          int     `bson:"childId"` //下级id
	Chip             int     `bson:"chip"`    //总贡献
	Tchip            int     `bson:"tchip"`   //总充值
	Lev              int     `bson:"lev"`
	Betchip          int64   `bson:"betchip"`          //总下注
	Tbetchip         float64 `bson:"tbetchip"`         //总下注返利
	TodayTchip       int     `bson:"todaytchip"`       //当日总充值
	TodayRebateTchip int     `bson:"todayrebatetchip"` //当日总充值返利
	TodayBetchip     int64   `bson:"todaybetchip"`     //当日总下注
	TodayTbetchip    float64 `bson:"todaytbetchip"`    //当日 总下注返利
	LastUpdateTime   string  `bson:"lastupdatetime"`   //数据最后更新时间
}

// 关系表
type relation struct {
	Id      int   `bson:"_id"`
	Parents []int `bson:"parents"`
	//Rebates        []rebateItem `bson:"rebates"`
	BelowNum             int     `bson:"belowNum"`
	Tolrebate            int     `bson:"tolrebate"`            //总返利值
	TolBelowCharge       int     `bson:"tolBelowCharge"`       //下线总充值
	Rebatechip           float64 `bson:"rebatechip"`           //可领取返利
	FreeValidinViteChips float64 `bson:"freeValidinViteChips"` //免费有效玩家可领取金额
	TodayFlowingChips    float64 `bson:"todayFlowingChips"`    //今日可领取金额
	TomorrowFlowingChips float64 `bson:"tomorrowFlowingChips"` //明日可领取金额 #
	AddFlowingTimes      string  `bson:"addFlowingTimes"`      //上次添加流水时间
	TolBetAll            int64   `bson:"tolBetAll"`            //累计下线金币下注
	TolBetFall           float64 `bson:"tolBetFall"`           //累计下线金币总返利
	TodayBetAll          int64   `bson:"todayBetAll"`          //今日团队下线金币下注
	TodayBetFall         float64 `bson:"todayBetFall"`         //今日团队下线金币总返利 #
	AmountAvailableChip  float64 `bson:"amountavailablechip"`  //可领取金额
	OneUnderNum          int     `bson:"oneUnderNum"`          //一级下线数量
}

// 记录返利日志表
type rebatelog struct {
	//Id         primitive.ObjectID `bson:"_id"`
	Uid        int    `bson:"uid"`      //本人id
	Parentid   int    `bson:"parentid"` //上级id
	Lev        int    `bson:"lev"`
	Price      int    `bson:"price"`      //实际充值金额
	Rebatechip int    `bson:"rebatechip"` //返利金额
	Orderno    string `bson:"orderno"`    //订单号
	AddTime    int64  `bson:"addTime"`    //添加时间
}
