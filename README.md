# 即时通信系统

## 一、语言版本

golang 1.13





## 二、文件说明

* myserver是服务端
  * main.go 主程序
  * server.go 服务端主要实现
  * user.go 用户类

* myclient是客户端
  * client.go 用户端实现



## 三、start

```shell
// 打开终端1号
cd myserver
./server

// 打开终端2号
cd ../myclient
./client

// 可以打开多个终端，重复上述第二步（开多个客户端）
```



## 四、实现功能

1. 用户上线及广播
2. 用户消息广播
3. 在线用户查询
4. 修改用户名
5. 超时强踢
6. 私聊功能



## 五、系统框图

![系统框图](.\img\system.png)
