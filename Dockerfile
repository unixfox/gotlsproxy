FROM linuxkit/ca-certificates:c1c73ef590dffb6a0138cf758fe4a4305c9864f4

COPY gotlsproxy gotlsproxy

ENTRYPOINT [ "/gotlsproxy" ]
