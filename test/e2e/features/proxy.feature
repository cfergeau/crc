@proxy @linux
Feature: Behind proxy test

    User starts CRC behind a proxy. They expect a successful start 
    and to be able to deploy an app and check its accessibility.

    Scenario: Setup the proxy container using podman
        Given executing "podman run --name squid -d -p 3128:3128 quay.io/crcont/squid" succeeds

    Scenario: Start CRC
        Given execute crc setup command succeeds
        And  execute crc config set http-proxy http://192.168.130.1:3128 command succeeds
        Then execute crc config set https-proxy http://192.168.130.1:3128 command succeeds
        When starting CRC with default bundle succeeds
        Then stdout should contain "Started the OpenShift cluster"
        And executing "eval $(crc oc-env)" succeeds
        When checking that CRC is running
        Then login to the oc cluster succeeds

    Scenario: Remove the proxy container and host proxy env (which set because of oc-env)
        Given executing "podman stop squid" succeeds
        And executing "podman rm squid" succeeds
        And executing "unset HTTP_PROXY HTTPS_PROXY NO_PROXY" succeeds

    Scenario: CRC delete and remove proxy settings from config
        When execute crc delete -f" succeeds
        Then stdout should contain "Deleted the OpenShift cluster"
        And  execute crc config unset http-proxy command succeeds
        And execute crc config unset https-proxy command succeeds
        And execute crc cleanup command succeeds
        

