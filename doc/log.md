# log使用提醒

直接使用 log 包中的 Logger 即可

1. log基本使用
   ```
   // 在你需要的地方使用就行：如调试（Debug）,错误信息（Error）,提醒（Info）,警告（Warn）
   // 举个例子

   if err != nil {
       log.Logger.Error("验证出错了")
   }
   ```
   同时项目文件下会创建`cmd/runtime/log/error.log`
