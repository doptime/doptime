---
slug: data-backup
title: 数据库的备份
type:  docs
sidebar_position: 5
---


:::info
### 数据库的备份
:::
1. #### 基于redis数据库间的同步备份

2. #### 文件备份：使用crontab备份
- 每3小时一次备份
- 每天一次备份
- vi  /etc/crontab
:::tip append this to /etc/crontab
17 */3 * * * root /usr/bin/rsync -av --delete /src/* /des/$(/bin/date +\%H)  
17 5 * * * root /usr/bin/rsync -av --delete /src/* /des/$(/bin/date +\%A)
:::
