---
slug: db-selection
title: 数据库选型 - Redis
type:  docs
sidebar_position: 1
---
## 为什么doptime 选择Redis做数据库

### 为什么redis 是好选择
 - Redis 能最有效满足理想后端框架的需求。它使用极简，功能丰富，性能强大. 
 - 存在dragonfly 这样的redis兼容数据库，使用redis接口，使用内存缓存，同时可以使用硬盘的大容量


 ### 为什么redis 未来是更的好选择
 - Nvme 2.0 支持key-value存储。而且价格也不贵。这对redis的持久化有重大助益。  
 - 3D DRAM后续会流行并开始往单颗粒TB演化。内存存储具有长期技术潜力。


## Redis选型的最佳实践
推荐使用 [Dragonfly](https://github.com/dragonflydb/dragonfly) 

- [Dragonfly](https://github.com/dragonflydb/dragonfly) 支撑内存缓存策略。这意味着你可以使用远大于内存的数据量。
- 可以支撑极高并发和极高性能
- doptime采用Dragonfly作为默认最佳选择
- 建议每个数据文件大小应小于8GB




