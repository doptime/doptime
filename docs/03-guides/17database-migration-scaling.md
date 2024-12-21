---
slug: database-migration-scaling
title: 数据库的迁移扩容
type:  docs
sidebar_position: 6
---

:::info
## Redis数据的迁移扩容
:::
典型的项目从一个DragonflyDB开始。每当数据量逼近内存容量时，则转移若干数据到磁盘存储
- 这是数据库迁移的示例代码。把大量数据从内存存储迁移到磁盘存储。
- 源数据库中, 目标Key是hash类型。
- 目标数据中, 目标Key是string类型。
- 其它类型的数据迁移也是类似的。按需修改你的代码。
:::tip 
```go
package main

import (
	"github.com/samber/lo"
	"github.com/doptime/doptime/data"
)

type TreeData struct {
	Tree         *SkillTree
	SessionName  string
	Mindmap      string
	DemoComments []string
	TTSInfos     []string
	UpdatedTime  int64
}

func ExportTreeDataToRds(trees []*SkillTree) {
	var keyAllTreeDataOld = data.New[string, *TreeData]("AllTreeData")
	datas, _ := keyAllTreeDataOld.HGetAll()
	var keyAllTreeData = data.New[string, *TreeData]("AllTreeData").WithRedis("mmuEvo")
	keyAllTreeData.SetAll(datas)
}
```
:::
