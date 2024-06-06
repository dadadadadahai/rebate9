package main

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 获取返利流水
func GetRebateChips(dclient *mongo.Database) ([]flowing_final, error) {
	findOpt := &options.FindOptions{}
	findOpt.SetLimit(20)
	cursor, err := dclient.Collection("flowing_final").Find(context.TODO(), bson.D{}, findOpt)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())
	datas := make([]flowing_final, 0)
	err = cursor.All(context.TODO(), &datas)
	return datas, err
}

// 返利流水进行处理
func rebateChipTableHandle(dclient *mongo.Database, rebates []flowing_final) {
	for _, val := range rebates {
		rebateChipHandle(dclient, val)
	}
}

// 执行具体返利处理
func rebateChipHandle(dclient *mongo.Database, valdata flowing_final) {
	//删除该条返利记录
	delRebateChipfinal(dclient, valdata.Id)
	//获取关系表
	relationTable := getRelationTable(dclient, valdata.Uid)
	if relationTable == nil {
		return
	}
	ParentIds := relationTable.Parents
	//只往上返利一级
	if len(ParentIds) > 0 {
		parentId := ParentIds[len(ParentIds)-1]
		parentRelationInfo := getRelationTable(dclient, parentId)
		if parentRelationInfo != nil {
			//按比例返值
			oneUnderNum := parentRelationInfo.OneUnderNum
			//取时间戳判断是不是同一天
			rebateval := calcRebateVal(oneUnderNum, valdata.Tchip)
			fmt.Printf("AddFlowingTimes %+v\n  tiem%v   rebateval", parentRelationInfo.AddFlowingTimes, time.Now().Format(time.DateOnly), rebateval)
			if parentRelationInfo.AddFlowingTimes == time.Now().Format(time.DateOnly) {
				//是同一天 累加
				dclient.Collection("extension_relation").UpdateOne(context.TODO(), bson.M{"_id": parentId}, bson.M{"$inc": bson.M{"todayBetFall": rebateval}})
			} else {
				dclient.Collection("extension_relation").UpdateOne(context.TODO(), bson.M{"_id": parentId}, bson.M{"$inc": bson.M{"tomorrowFlowingChips": parentRelationInfo.TodayBetFall}})
				dclient.Collection("extension_relation").UpdateOne(context.TODO(), bson.M{"_id": parentId}, bson.M{"$set": bson.M{"todayBetFall": rebateval}})
				dclient.Collection("extension_relation").UpdateOne(context.TODO(), bson.M{"_id": parentId}, bson.M{"$set": bson.M{"addFlowingTimes": time.Now().Format(time.DateOnly)}})
			}
		}
	}
}

// 计算返利值
func calcRebateVal(underNum, tChip int) float64 {
	var rebateVal float64 = 0
	if underNum >= 1 && underNum <= 4 {
		rebateVal = float64(tChip) * 30 / 10000
	} else if underNum >= 5 && underNum <= 9 {
		rebateVal = float64(tChip) * 45 / 10000
	} else if underNum >= 10 && underNum <= 49 {
		rebateVal = float64(tChip) * 50 / 10000
	} else if underNum >= 50 && underNum <= 99 {
		rebateVal = float64(tChip) * 60 / 10000
	} else if underNum >= 100 && underNum <= 299 {
		rebateVal = float64(tChip) * 65 / 10000
	} else if underNum >= 300 && underNum <= 699 {
		rebateVal = float64(tChip) * 70 / 10000
	} else if underNum >= 700 && underNum <= 999 {
		rebateVal = float64(tChip) * 85 / 10000
	} else if underNum >= 1000 && underNum <= 1999 {
		rebateVal = float64(tChip) * 100 / 10000
	} else if underNum >= 2000 {
		rebateVal = float64(tChip) * 200 / 10000
	}
	return rebateVal
}

// 增加返利流水金额
func addUserInforebateflowing(dclient *mongo.Database, uid int, rebateflowingchip float64, parentInfo *relation) {
	coll := dclient.Collection("extension_relation")
	filter := bson.D{{"_id", uid}}
	update := bson.D{{"$inc", bson.M{"rebatechip": rebateflowingchip}}, {"$set", bson.M{"addFlowingTimes": time.Now().Unix()}}}
	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		//log.Println("addUserInforebatechip", err.Error())
		selflog.Error("addUserInforebatechip=" + err.Error())
	}
}

// 记录用户chips记录
func addChipRelations(dclient *mongo.Database, parentInfo *relation, belowUid int, ftbetchip float64, betchip, lev int) {
	var rItem rebateItem
	filter := bson.D{{"uid", parentInfo.Id}, {"childId", belowUid}}
	err := dclient.Collection("rebateItem").FindOne(context.TODO(), filter).Decode(&rItem)
	if err != nil {
		//不存在记录
		rItemPtr := &rebateItem{
			Uid:              parentInfo.Id,
			ChildId:          belowUid,
			Betchip:          int64(betchip),
			Tbetchip:         ftbetchip,
			Lev:              lev,
			TodayTchip:       0,
			TodayRebateTchip: 0,
			TodayBetchip:     int64(betchip),
			TodayTbetchip:    ftbetchip,
			LastUpdateTime:   time.Now().Format("20060102"),
		}
		dclient.Collection("rebateItem").InsertOne(context.TODO(), rItemPtr)
	} else {
		rItem.Betchip = rItem.Betchip + int64(betchip)
		rItem.Tbetchip = rItem.Tbetchip + ftbetchip
		// 兼容老数据第一次生成时间
		if rItem.LastUpdateTime == "" {
			rItem.LastUpdateTime = time.Now().Format("20060102")
		}
		// 如果是同一天
		if rItem.LastUpdateTime == time.Now().Format("20060102") {
			rItem.TodayBetchip = rItem.TodayBetchip + int64(betchip)
			rItem.TodayTbetchip = rItem.TodayTbetchip + ftbetchip
		} else {
			rItem.TodayTchip = 0
			rItem.TodayRebateTchip = 0
			rItem.TodayBetchip = int64(betchip)
			rItem.TodayTbetchip = ftbetchip
			rItem.LastUpdateTime = time.Now().Format("20060102")
		}
		dclient.Collection("rebateItem").UpdateOne(context.TODO(), filter, bson.D{{"$set", rItem}})
	}
	// 更新可领取金额
	parentInfo.AmountAvailableChip += rItem.TodayTbetchip
	parentInfo.TolBetAll = parentInfo.TolBetAll + int64(betchip)
	parentInfo.TolBetFall = parentInfo.TolBetFall + ftbetchip
	parentInfo.TodayBetAll = parentInfo.TolBetAll + int64(betchip)
	parentInfo.TodayBetFall = parentInfo.TolBetFall + ftbetchip
	dclient.Collection("extension_relation").UpdateOne(context.TODO(), bson.D{{"_id", parentInfo.Id}}, bson.D{{"$set", bson.D{{"tolBetAll", parentInfo.TolBetAll}, {"tolBetFall", parentInfo.TolBetFall}}}})

}

// 结算流程
func settlementRebate(dclient *mongo.Database, rebates []rebateItem) {
	returnTable := returnTable()
	for _, valdata := range rebates {
		if valdata.TodayBetchip > 0 {
			//获取所有下限
			filter := bson.D{{"TodayBetchip", bson.M{"$gt": 0}}, {"Lev", 1}}
			cursor, err := dclient.Collection("rebateItem").Find(context.TODO(), filter)
			if err != nil {
				return
			}
			defer cursor.Close(context.TODO())
			childdatas := make([]rebateItem, 0)
			err = cursor.All(context.TODO(), &childdatas)
			if err != nil {
				return
			}
			// 获取计算今日总返利
			for _, tableInfo := range returnTable.teamRebatePropors {
				if valdata.TodayBetchip >= int64(tableInfo.min) && valdata.TodayBetchip <= int64(tableInfo.max) {
					valdata.TodayTbetchip = valdata.TodayTbetchip + float64(valdata.TodayBetchip*(int64(tableInfo.proportion)/10000))
				}
			}
			for _, childdata := range childdatas {
				for _, tableInfo := range returnTable.teamRebatePropors {
					if childdata.TodayBetchip >= int64(tableInfo.min) && childdata.TodayBetchip <= int64(tableInfo.max) {
						valdata.TodayTbetchip = valdata.TodayTbetchip - float64(childdata.TodayBetchip*(int64(tableInfo.proportion)/10000))
					}
				}
			}
			dclient.Collection("rebateItem").UpdateOne(context.TODO(), bson.D{{"uid", valdata.Uid}}, bson.D{{"$set", bson.D{{"todaytbetchip", valdata.TodayTbetchip}}}})
			// 删除ex表的parent信心0
			dclient.Collection("extension_relation").UpdateOne(context.TODO(), bson.D{{"uid", valdata.Uid}}, bson.D{{"$set", bson.D{{"todayBetAll", 0}}}})
			dclient.Collection("extension_relation").UpdateOne(context.TODO(), bson.D{{"uid", valdata.Uid}}, bson.D{{"$set", bson.D{{"todayBetFall", 0}}}})
		}
	}
}
func GetRebateItem(dclient *mongo.Database) ([]rebateItem, error) {
	filter := bson.D{{"TodayBetchip", bson.M{"$gt": 0}}}
	cursor, err := dclient.Collection("rebateItem").Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())
	datas := make([]rebateItem, 0)
	err = cursor.All(context.TODO(), &datas)
	return datas, err
}

// 删除已处理的返利流水记录
func delRebateChipfinal(dclient *mongo.Database, Id string) {
	dclient.Collection("flowing_final").DeleteOne(context.TODO(), bson.D{{"_id", Id}})
}
