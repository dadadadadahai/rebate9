package main

import (
	"encoding/json"
	"fmt"
)

type teamRebatePropor struct {
	ID         int
	min        int
	max        int
	proportion int
}
type tableTeamRebatePropor struct {
	teamRebatePropors []teamRebatePropor
}

func returnTable() tableTeamRebatePropor {

	var tableJson = "[{\"ID\":1,\"max\":1000,\"min\":1,\"proportion\":10\n},{\"ID\":2,\"max\":200000,\"min\":1001,\"proportion\":20\n},{\"ID\":3,\"max\":400000,\"min\":200001,\"proportion\":30\n},{\"ID\":4,\"max\":800000,\"min\":400001,\"proportion\":40\n},{\"ID\":5,\"max\":1600000,\"min\":800001,\"proportion\":50\n},{\"ID\":6,\"max\":3200000,\"min\":1600001,\"proportion\":60\n},{\"ID\":7,\"max\":6400000,\"min\":3200001,\"proportion\":70\n},{\"ID\":8,\"max\":12000000,\"min\":6400001,\"proportion\":80\n},{\"ID\":9,\"max\":24000000,\"min\":12000001,\"proportion\":90\n},{\"ID\":10,\"max\":50000000,\"min\":24000001,\"proportion\":100\n},{\"ID\":11,\"max\":100000000,\"min\":50000001,\"proportion\":110\n},{\"ID\":12,\"max\":150000000,\"min\":100000001,\"proportion\":120\n},{\"ID\":13,\"max\":200000000,\"min\":150000001,\"proportion\":130\n},{\"ID\":14,\"max\":300000000,\"min\":200000001,\"proportion\":140\n},{\"ID\":15,\"max\":400000000,\"min\":300000001,\"proportion\":150\n},{\"ID\":16,\"max\":500000000,\"min\":400000001,\"proportion\":160\n},{\"ID\":17,\"max\":600000000,\"min\":500000001,\"proportion\":170\n},{\"ID\":18,\"max\":700000000,\"min\":600000001,\"proportion\":180\n},{\"ID\":19,\"max\":800000000,\"min\":700000001,\"proportion\":190\n},{\"ID\":20,\"max\":1000000000,\"min\":800000001,\"proportion\":200\n}]"

	//定义一个Monster 实例
	var table tableTeamRebatePropor
	err := json.Unmarshal([]byte(tableJson), &table)
	if err != nil {
		fmt.Printf("unmarshal err=%v\n", err)
	}
	return table
}
