# ckafka

## Config(example for Confluent Kafka)
test.cfg:
```
bootstrap.servers={HOST}:9092
security.protocol=SASL_SSL
sasl.mechanisms=PLAIN
sasl.username={USERNAME}
sasl.password={PASSWORD}
```

## Producing a message
```
ckafka -config /path/to/test.cfg -topic test -key 123 -headers key1=value1,key2=value2 -message {"key":"test key"} 
```

## Install
### Brew
```
brew tap incu6us/homebrew-tap
brew install ckafka
```

### Snap
```
snap install ckafka
```
