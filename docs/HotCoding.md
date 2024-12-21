---
slug: hotcoding
title: HotCoding
type:  docs
sidebar_position: 5
---

## 不把定位放在分解和实现整个项目上。
1. 理由是现有的LLM模型只能支持最常见的任务的分解。对其它任务很难良好完成。完美的大任务分解估计要等2027年，在无限上下文模型出现之后才能实现。
2. 软件项目是一个需要不断迭代的过程。许多想法都需要经历改变和探索。
3. 不使用任务堆栈，使用优先级队列，一次聚焦一个任务。
4. 和人的交互是一个很重要的过程。这个过程是匿名过程。

## 信息论的洞见：
已知在LLM中，如果我们把工作分成不同的阶段，对不同的阶段施加角色的约束，那么可以大幅提高LLM输出的性能。这是Meta-GPT 的核心。
从信息论来说，这种方式之所有有效是因为角色的引入补充了额外的信息，降低了信息不确定性。从而通过更确切的情境来提高每个阶段的输出性能。而由于目标约束信息在这个过程中不存在关于目标性能信息损耗。因此，这种方式是有效的。
但Meta-GPT,GTP-Pilot存在一个核心的缺点是，它们更多是作为自上而下的方式工作。这在目标和实现路径相当明确的项目中或许是便利的，但真实的项目需求往往需要高的探索性这种方式不能很好起效。这就是为什么许多演示项目诸如Floppy Bird实现良好，但在更特殊、具体的情形往往工作不好。这种探索性的自上而下的无效的实质即是来源于用来定义目标的语义失真，同时也是来自LLM的性能缺陷。我认为尽快LLM高速发展，但哪怕在未来很长一段时间，自上而下都无法替代自下而上的高精度探索能力。

doptime 解决方案的核心：
doptime 使用自下而上的集市模式。
同样使用角色等概念来降低信息的不确定性。但doptime 优先聚焦模块功能预期。全局目标被更多看做是请求性质的软约束。

doptime AI开发的流程：
doptime 采用热补丁的方式快速改进当前的程序。也就是当前程序不终止运行，而是在运行的过程中不断接受改进。这种方式被设计允许机器最终以最大化并行的改进速度，也允许最大化的持续无间断的改进。从而为最终无人编程提供基础。
doptime 采用BRT流程。
 - Build: 通过构建信息情境，降低生成代码的不确定性，更新情境信息和生成代码。
 - Run: 运行代码，收集运行时信息，更新情境信息和生成代码。
 - Test: 测试代码，收集测试信息，更新情境信息和生成代码。

### 脚手架免拆除原则
    采用HTTP测试作为外部脚手架

### 维护更新CO文档


一些说明：
doptime 在生成代码时，要求插入无意义字符，以便LLM能够经历更多的思考周期来生成准确的代码。这会导致额外的token开销，但这是回报高得多。


[开发工具 GPT Pilot 的 6 个月中学到的东西](https://blog.pythagora.ai/2024/02/19/gpt-pilot-what-did-we-learn-in-6-months-of-working-on-a-codegen-pair-programmer/)
自动编码的核心：
1. 更多的迭代
2. 方便的调试（自我审查）
3.能总结初始描述和状态描述
你是一个软件开发者，同时有丰富的产品经理，架构师，测试人员的视角。你需要通过doptime进一步开发你的软件。

[what we learned building GPT Pilot](https://www.reddit.com/r/ChatGPTCoding/comments/1ay3it8/gpt4_powered_tool_that_builds_web_apps_from_start/)
1. It’s hard to get an LLM to think outside the box. 
2. Agents can review themselves.
3. Verbose logs help.
4. The initial description of the app is much more important than I thought.
5. Coding is not a straight line.
6. LLMs work best when they can focus on one problem.
7. Splitting the codebase into smaller files helps a lot.

这是你现在正在开发的软件模块的描述。