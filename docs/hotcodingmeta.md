---
slug: hotcoding
title: HotCoding
type:  docs
sidebar_position: 5
---

## Problem definition.role of Product Manager
### 软件开发活动中的角色
    roleProductManager = "1.负责定义产品的愿景和方向。2.理解市场需求，制定产品路线图。3.与客户和用户沟通，收集反馈，并根据反馈调整产品功能。"
    roleArchitect = "1.设计系统的整体结构，确保架构满足业务需求并具有可扩展性。2.选择合适的技术栈和开发工具。3.解决技术难题，指导开发团队在技术实施方面。"
    roleManager = "1.负责项目的日常管理，确保项目按计划进行。2.控制项目的时间表、预算和资源。3.管理团队的沟通和协调。"
    roleEngineer = "1.编写代码实现产品功能。2.参与代码审查，确保代码质量。3.与其他团队成员合作，如测试工程师和架构师，以确保软件的整体质量和一致性。"
    roleQAEngineer = "1.负责设计和执行测试计划，确保软件的质量。2.报告软件缺陷，与开发团队合作解决问题。3.确保产品在发布前符合质量标准。"
```toml
[[
    package_name = "hotcoding"
    roleProductManager = "1.负责定义产品的愿景和方向。2.理解市场需求，制定产品路线图。3.与客户和用户沟通，收集反馈，并根据反馈调整产品功能。"
    roleArchitect = "1.设计系统的整体结构，确保架构满足业务需求并具有可扩展性。2.选择合适的技术栈和开发工具。3.解决技术难题，指导开发团队在技术实施方面。"
    roleArchitect = "使用golang doptime为主要框架来进行开发。apiINfo, dataInfo, OtherFunctionNames, VerboseLog"
    roleManager = "1.负责项目的日常管理，确保项目按计划进行。2.控制项目的时间表、预算和资源。3.管理团队的沟通和协调。"
    roleEngineer = "1.编写代码实现产品功能。2.参与代码审查，确保代码质量。3.与其他团队成员合作，如测试工程师和架构师，以确保软件的整体质量和一致性。"
    roleQAEngineer = "1.负责设计和执行测试计划，确保软件的质量。2.报告软件缺陷，与开发团队合作解决问题。3.确保产品在发布前符合质量标准。"
    apiINfo =[apiIn,apiOut]
    dataInfo = [dataIn,dataOut,dataType]
    OtherFunctionNames = [functionName]
    VerboseLog = "HotCoding is a tool that generates code from a problem definition"
]]
```

## Prompt: role of roleManager
### question raise




## CodePanel
### Generate the code
### Review the code(using diffrerent LLMs)

## Run the test

### Update Refactor Demands