# Voiper

TODO

### Asterisk test setup


Install asterisk on your favourite distribution or ubuntu:
```bash
sudo apt install asterisk asterisk-core-sounds-en -y
```


`/etc/asterisk/pjsip.conf`:

```
[global]
type=global
default_realm=ec2-18-198-48-194.eu-central-1.compute.amazonaws.com

[transport-udp]
type=transport
protocol=udp
bind=0.0.0.0

[baresip]
type=endpoint
context=baresip-context
disallow=all
allow=ulaw
auth=baresip-auth
aors=baresip

[baresip-auth]
type=auth
auth_type=userpass
password=password
username=baresip

[baresip]
type=aor
max_contacts=1
remove_existing=yes
```


`/etc/asterisk/rtp.conf`:

```
[general]
rtpstart=10000
rtpend=20000
```


`/etc/asterisk/extensions.conf`:

```
[baresip-context]
exten => baresip,1,Answer()
 same => n,Playback(hello-world)
 same => n,Hangup()
```


`~/.baresip/accounts`:

```
<sip:baresip@ec2-18-198-48-194.eu-central-1.compute.amazonaws.com>;auth_pass=password;outbound=sip:ec2-18-198-48-194.eu-central-1.compute.amazonaws.com
```


```bash
baresip -e "/dial baresip"
```
