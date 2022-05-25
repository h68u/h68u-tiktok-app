# log使用提醒

1. 请不要使用log.Namespace(现在很可能删除了)如果你需要添加`namespace`可以这样使用
   ```
    log.Logger.Error(
        "msg",
        zap.Namespace("namespace"),
        zap.Error(err) // or use zap.Any("msg", Any)
   )
   ```
2.