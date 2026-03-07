# Requirements Document

## Project Description (Input)
次のような構成を取るコンテナ環境がある。

```
docker network create --internal=true prv_net
docker network create pub_net
docker run -d --name=lg --hostname=localgate.test -p 9000:9000 --network=prv_net localgate
docker network connect pub_net lg
docker run -d --name=foobar --network=prv_net foobarservice
```

foobarコンテナ内からサービスを登録するため`http://localgate.test:9000/services`に対してリクエストを送信すると、`localgate`サービスへのプロキシだと判断されて"service not found"となる。

この課題を解決してほしい。

## Requirements
<!-- Will be generated in /kiro:spec-requirements phase -->
