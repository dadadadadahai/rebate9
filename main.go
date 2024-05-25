package main

import (
	"context"
	"time"

	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongodbUrl = ""
var selflog *LogStruct

// 返利计算程序
func main() {
	selflog = &LogStruct{}
	selflog.Init()
	selflog.SetLevel(logrus.DebugLevel)
	cfgs, err := ini.Load("config.ini")
	if err != nil {
		//log.Fatalln(err.Error())
		selflog.Error(err.Error())
		return
	}
	mongodbUrl = cfgs.Section("DBInfo").Key("url").String()
	database := cfgs.Section("DBInfo").Key("database").String()
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongodbUrl))
	if err != nil {
		//log.Fatalln(err.Error())
		selflog.Error(err.Error())
		return
	}
	dclient := client.Database(database)
	//创建表索引
	CreateIndex(dclient)
	selflog.Info("主线程启动")
	for {
		timeStr := time.Now().Format("2006-01-02")
		//fmt.Println("timeStr:", timeStr)
		t, _ := time.Parse("2006-01-02", timeStr)
		timeNumber := t.Unix()
		//fmt.Println("timeNumber:", timeNumber)
		//零点结算
		if time.Now().Unix() == timeNumber {
			//结算流程
			rebateItemDatas, errno := GetRebateItem(dclient)
			if errno != nil {
				//log.Println("rebate_final", err.Error())
				selflog.Error(errno.Error())
			} else {
				if len(rebateItemDatas) > 0 {
					settlementRebate(dclient, rebateItemDatas)
				}
			}
		}

		rebateDatas, err := GetRebates(dclient)
		if err != nil {
			//log.Println("rebate_final", err.Error())
			selflog.Error(err.Error())
		} else {
			if len(rebateDatas) > 0 {
				rebateTableHandle(dclient, rebateDatas)
			}
		}
		rebateChipDatas, errno := GetRebateChips(dclient)
		if errno != nil {
			//log.Println("rebate_final", err.Error())
			selflog.Error(errno.Error())
		} else {
			if len(rebateChipDatas) > 0 {
				rebateChipTableHandle(dclient, rebateChipDatas)
			}
		}
		time.Sleep(2 * time.Second)
	}
}

// 创建表索引
func CreateIndex(dclient *mongo.Database) {
	dclient.Collection("rebateItem").Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.D{{Key: "uid", Value: 1},
			{Key: "chip", Value: 1},
			{Key: "lastupdatetime", Value: 1},
		},
	})
	dclient.Collection("rebatelog").Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.D{{Key: "parentid", Value: 1},
			{Key: "addTime", Value: 1}},
	})
}

// 获取需要返利,
func GetRebates(dclient *mongo.Database) ([]rebate_final, error) {
	findOpt := &options.FindOptions{}
	findOpt.SetLimit(20)
	cursor, err := dclient.Collection("rebate_final").Find(context.TODO(), bson.D{}, findOpt)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())
	datas := make([]rebate_final, 0)
	err = cursor.All(context.TODO(), &datas)
	return datas, err
}

// 返利进行处理
func rebateTableHandle(dclient *mongo.Database, rebates []rebate_final) {
	for _, val := range rebates {
		rebateHandle(dclient, val)
	}
}

// 执行具体返利处理
func rebateHandle(dclient *mongo.Database, valdata rebate_final) {
	//删除该条返利记录
	delRebatefinal(dclient, valdata.Id)
	//获取关系表
	relationTable := getRelationTable(dclient, valdata.Uid)
	if relationTable == nil {
		return
	}
	ParentIds := relationTable.Parents
	lev := 0
	//千分比
	rates := []int{50, 25, 15, 10}
	for index := len(ParentIds); index > 0; index-- {
		parentId := ParentIds[index-1]
		parentInfo := getRelationTable(dclient, parentId)
		rebateChip := int(float64(valdata.Price) * (float64(rates[lev]) / 1000))
		//执行返利
		addUserInforebatechip(dclient, parentId, rebateChip)
		//添加父类
		addRelations(dclient, parentInfo, valdata.Uid, valdata.Price, rebateChip, lev+1)
		//记录日志
		addRebateLog(dclient, valdata.Uid, parentId, lev+1, valdata.Price, rebateChip, valdata.OrderNo)
		lev++
		if lev >= 4 {
			break
		}
	}
}
func getRelationTable(dclient *mongo.Database, uid int) *relation {
	relationTable := &relation{}
	filter := bson.D{{"_id", uid}}
	err := dclient.Collection("extension_relation").FindOne(context.TODO(), filter).Decode(relationTable)
	if err != nil {
		//log.Println("getRelationTable=", err.Error())
		selflog.Error("getRelationTable=" + err.Error())
		return nil
	}
	return relationTable
}

// 增加返利金额
func addUserInforebatechip(dclient *mongo.Database, uid, rebatechip int) {
	coll := dclient.Collection("extension_relation")
	filter := bson.D{{"_id", uid}}
	update := bson.D{{"$inc", bson.M{"rebatechip": rebatechip}}}
	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		//log.Println("addUserInforebatechip", err.Error())
		selflog.Error("addUserInforebatechip=" + err.Error())
	}
}

// 记录用户返利记录
// belowUid,price,rebatechip 下线id 下线总充值,下线返利
func addRelations(dclient *mongo.Database, parentInfo *relation, belowUid, price, rebatechip, lev int) {
	var rItem rebateItem
	filter := bson.D{{"uid", parentInfo.Id}, {"childId", belowUid}}
	err := dclient.Collection("rebateItem").FindOne(context.TODO(), filter).Decode(&rItem)
	if err != nil {
		//不存在记录
		rItemPtr := &rebateItem{
			Uid:              parentInfo.Id,
			ChildId:          belowUid,
			Chip:             rebatechip,
			Tchip:            price,
			Lev:              lev,
			TodayTchip:       price,
			TodayRebateTchip: rebatechip,
			TodayBetchip:     0,
			TodayTbetchip:    0,
			LastUpdateTime:   time.Now().Format("20060102"),
		}
		dclient.Collection("rebateItem").InsertOne(context.TODO(), rItemPtr)
	} else {
		rItem.Chip = rItem.Chip + rebatechip
		rItem.Tchip = rItem.Tchip + price
		// 兼容老数据第一次生成时间
		if rItem.LastUpdateTime == "" {
			rItem.LastUpdateTime = time.Now().Format("20060102")
		}
		// 如果是同一天
		if rItem.LastUpdateTime == time.Now().Format("20060102") {
			rItem.TodayTchip = rItem.TodayTchip + price
			rItem.TodayRebateTchip = rItem.TodayRebateTchip + rebatechip
		} else {
			rItem.TodayTchip = price
			rItem.TodayRebateTchip = rebatechip
			rItem.TodayBetchip = 0
			rItem.TodayTbetchip = 0
			rItem.LastUpdateTime = time.Now().Format("20060102")
		}
		dclient.Collection("rebateItem").UpdateOne(context.TODO(), filter, bson.D{{"$set", rItem}})
	}
	parentInfo.Tolrebate = parentInfo.Tolrebate + rebatechip
	parentInfo.TolBelowCharge = parentInfo.TolBelowCharge + price
	dclient.Collection("extension_relation").UpdateOne(context.TODO(), bson.D{{"_id", parentInfo.Id}}, bson.D{{"$set", bson.D{{"tolrebate", parentInfo.Tolrebate}, {"tolBelowCharge", parentInfo.TolBelowCharge}}}})
}
func addRebateLog(dclient *mongo.Database, uid, parentid, lev, price, rebatechip int, orderno string) {
	relog := &rebatelog{
		Uid:        uid,               //本人id
		Parentid:   parentid,          //上级id
		Lev:        lev,               //位于上级第几级
		Price:      price,             //充值了多少金币
		Rebatechip: rebatechip,        //直接返利金币
		Orderno:    orderno,           //订单号
		AddTime:    time.Now().Unix(), //时间戳
	}
	dclient.Collection("rebatelog").InsertOne(context.TODO(), relog)
}

// 删除已处理的返利记录
func delRebatefinal(dclient *mongo.Database, Id primitive.ObjectID) {
	dclient.Collection("rebate_final").DeleteOne(context.TODO(), bson.D{{"_id", Id}})
}
