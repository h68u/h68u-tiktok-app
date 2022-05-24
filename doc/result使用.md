#  result使用

![image.png](https://p9-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/480437e5be51436c8f4ce2c6ec270856~tplv-k3u1fbpfcp-watermark.image?)

包名：res

# 返回错误json示例

```
res.Error(c, res.Status{
			StatusCode: res.LoginErrorStatus.StatusCode,
			StatusMsg:  res.LoginErrorStatus.StatusMsg,
		})
		return
```

# 返回成功json示例

```
res.Success(c, res.R{
		"userid": data.UserId,
		"token":  data.Token,
	})
```

# 添加错误码示例

在result包下

![image.png](https://p9-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/b7bb52cab7a44892abd83f647063bedd~tplv-k3u1fbpfcp-watermark.image?)