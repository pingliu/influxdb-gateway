# Influxdb Gateway

### 功能描述

* 采用udp协议接受数据
* 对数据进行格式检查
* 对数据进行gzip,解决跨机房流量问题

### 使用方式

```
make b
./bin/influxdb-gateway -h  #查看使用方式
```

### influxdb 接入点高可用方式
[InfluxDB Proxy](https://github.com/shell909090/influx-proxy)
