# Influxdb Gateway

### 功能描述

* 采用udp协议接受数据
* 对数据进行格式检查
* 对数据进行gzip,解决跨机房流量问题

### 使用方式

```
cd $GOPATH/src/github.com/influxdata/influxdb
git checkout v1.2.0

go get github.com/uber-go/zap
mkdir $GOPATH/src/go.uber.org/
cd $GOPATH/src/go.uber.org/
mv $GOPATH/src/github.com/uber-go/zap .
cd zap
git checkout fbae0281ffd546fa6d1959fec6075ac5da7fb577

cd $GOPATH/src/github.com/pingliu/influxdb-gateway
make all
./bin/influxdb-gateway -h  #查看使用方式
```

### influxdb 高可用
[InfluxDB Proxy](https://github.com/shell909090/influx-proxy)
