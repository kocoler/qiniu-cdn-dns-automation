### qiniu-cdn-dns-automation

Simple dns auto update script for qiniu-cdn's dns, using as **renew-hook** script.


*Suggest use wildcard domain.*


#### WORK FLOW:

before:

    0. have installed openssl, acme.sh.
    
    1. acme.sh renew ssl record info.
    
    2. generate cer file.
    
    3. arouse this script
    
now:

    0. prepare env variables(QINIU_AK, QINIU_SK, QINIU_DOMAIN, ACME_KEY_PATH).
    
    1. generate pem file from cer to pem file. => move to openssl
    
    2. upload pem file to qiniu-cdn.
    
    3. update qiniu-cdn ssl file.



#### example

```bash
    acme.sh --issue --dns dns_ali -d *.muxixyz.com \
    --pre-hook "echo issuing ssl ... && echo !" \
    --post-hook "openssl rsa -in '/root/.acme.sh/*.muxixyz.com/*.muxixyz.com.key' -out '/root/.acme.sh/*.muxixyz.com/*.muxixyz.com.key.pem' -outform PEM && /opt/auto-cdn-dns/qiniu-cdn-dns-automation_linux_amd64" \
    --renew-hook "openssl rsa -in '/root/.acme.sh/*.muxixyz.com/*.muxixyz.com.key' -out '/root/.acme.sh/*.muxixyz.com/*.muxixyz.com.key.pem' -outform PEM && /opt/auto-cdn-dns/qiniu-cdn-dns-automation_linux_amd64" \
    --dnssleep
```
