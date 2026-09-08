
1. Edit Redis config:

```bash
sudo nvim /etc/redis/redis.conf
```
put
```nvim
bind 0.0.0.0
protected-mode no
```

i use valkey , its a fork

```nvim
/etc/valkey/valkey.conf
```
```bash
sudo systemctl restart valkey
```