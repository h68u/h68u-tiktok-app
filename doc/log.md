# log使用提醒

1. log基本使用
   ```
   // 在你需要的地方使用就行：如调试（Debug）,错误信息（Error）,提醒（Info）,警告（Warn）
   // 举个例子
   if err != nil {
       log.Logger.Error(
           "验证出错了",
           zap.Error(err),
           // 或者通用的 zap.Any(key string, value interface{})
       )
   }
   ```
   同时项目文件下会创建`runtime/log/error.log`
2. 请不要使用log.Namespace(现在很可能删除了)
3. 如果你需要添加`namespace`可以这样使用
   ```
    log.Logger.Error(
        "msg",
        zap.Namespace("namespace"),
        zap.Error(err), // or use zap.Any("msg", Any)
   )
   ```