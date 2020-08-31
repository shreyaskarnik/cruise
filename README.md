我是光年实验室高级招聘经理。
我在github上访问了你的开源项目，你的代码超赞。你最近有没有在看工作机会，我们在招软件开发工程师，拉钩和BOSS等招聘网站也发布了相关岗位，有公司和职位的详细信息。
我们公司在杭州，业务主要做流量增长，是很多大型互联网公司的流量顾问。公司弹性工作制，福利齐全，发展潜力大，良好的办公环境和学习氛围。
公司官网是http://www.gnlab.com,公司地址是杭州市西湖区古墩路紫金广场B座，若你感兴趣，欢迎与我联系，
电话是0571-88839161，手机号：18668131388，微信号：echo 'bGhsaGxoMTEyNAo='|base64 -D ,静待佳音。如有打扰，还请见谅，祝生活愉快工作顺利。

# Heptio Cruise 🚢

Stress free HTTP monitoring for your Kubernetes hosted Web Applications.

**Maintainers:** [Heptio][0]

[![Build Status][1]][2]

## Overview
Cruise automatically configures HTTP monitoring of Kubernetes Ingress resources.

Read the [annoucement here][3].

## Installation

1. Deploy Cruise to your cluster
    ```
    % kubectl apply -f https://github.com/heptiolabs/cruise/blob/master/deployment/cruise.yaml
    ```
2. Configure your Pingdom API credentials
    ```
    % kubectl -n heptio-cruise create secret generic cruise \
            --from-literal=PINGDOM_USERNAME=you@yourdomain \
            --from-literal=PINGDOM_PASSWORD=yourpassword \
            --from-literal=PINGDOM_APIKEY=yourapikey
    ```

You're all set!
Pingdom will let you know if any of your web applications have run aground.

[0]: https://github.com/heptio
[1]: https://travis-ci.org/heptiolabs/cruise.svg?branch=master
[2]: https://travis-ci.org/heptiolabs/cruise
[3]: https://blog.heptio.com/hello-cruise-491852b98a89
