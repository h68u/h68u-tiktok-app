# h68u-tiktok-app

第三届字节跳动青训营 抖音APP项目 h68u组

## 项目简介

该项目是字节跳动青训营，抖声APP的后端实现，由 h68u 组 8 名成员共同完成

项目使用 MySQL 数据库、Redis 缓存和 OSS，配置了 CI/CD，并提供了详实的文档

项目文档不出意外均位于 doc/ 下

## 项目结构
项目目录大致分为
*注：目录中的文件夹或文件有部分添加到 `.gitignore` 中，所以并未显示*

- cmd/ 存放项目启动相关
  - bin/ 本地调试时存放项目生成的二进制可执行文件
  - runtime/ 存放项目运行时产生的文件
    - logs/ 存放项目运行时日志
- common/ 
  - config/ 项目的配置读取
  - cron/ 定时任务
  - db/ 数据库支持
  - log/ 日志配置
  - model/ 数据表
  - oss/ OSS，配置了 minio 和 aliyun 两种
  - result/ 规范了全局的错误码和返回值
- controller/ 控制层实现
- doc/ 存放项目文档
  - imgs/ 存放文档所需的图片
- middlewire/ 中间件
- service/ 服务层实现
- test/ 项目单元测试
- util/ 存放封装了项目通用逻辑的小工具
- app.example.yaml 项目配置文件示例

## 项目汇报文档(ppt)地址
[飞书文档](https://s97bh2semh.feishu.cn/file/boxcnHOxH4scTc8A2ODugoiW6ib)

## 项目核心点介绍
[青训营笔记](https://juejin.cn/post/7108371999004557319)
