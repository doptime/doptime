---
slug: db-setup
title: 数据库搭建
type:  docs
sidebar_position: 2
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';


:::info doptime 使用 REDIS 数据库，该如何搭建？
## 通过 Docker搭建 REDIS 数据库
:::
<Tabs>
<TabItem value="Dragonfly" label="Dragonfly" default>

```yaml title="docker-compose.yml"
version: '3.8'
services:
  dragonfly:
    image: 'docker.dragonflydb.io/dragonflydb/dragonfly'
    ulimits:
      memlock: -1
    memswap_limit: -1
    ports:
      - "6379:6379"
    environment:
      - "DFLY_requirepass=XWkHqup04Gcm7Plieh5x"
    # For better performance, consider `host` mode instead `port` to avoid docker NAT. `host` mode is NOT currently supported in Swarm Mode.
    # network_mode: "host"
    command: [
      "--snapshot_cron", "5 * * * *", 
      "--keys_output_limit", "65536", 
      "--dbfilename", "dragonflydump", 
      "--tiered_prefix", "some/path/basename",
      "--maxmemory","512m",
    ]
    volumes:
      - /home/user/dragonflydb:/data
``` 
 </TabItem>


</Tabs>

## 直接安装请参见官网文档:
- [安装DragonflyDB](https://www.dragonflydb.io/docs/getting-started)  