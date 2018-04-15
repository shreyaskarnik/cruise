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
