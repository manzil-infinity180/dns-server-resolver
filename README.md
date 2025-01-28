## Demo 


https://github.com/user-attachments/assets/6de030a8-80cb-428b-a94c-0f6b424253e6

## Steps 

```console
// run this in one terminal

$ go run main.go
```

```console
/* run this command in some other termianl */

$ dig @127.0.0.1 google.com
$ dig @127.0.0.1 amazon.com
$ dig @127.0.0.1 <website_url>

// here i am using `dig` you can also use some other tools like `host`
```

## Output

```console
rahulxf@dns-server-resolver:~ $ dig @127.0.0.1 google.com                      

; <<>> DiG 9.10.6 <<>> @127.0.0.1 google.com
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 6194
;; flags: qr; QUERY: 0, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0

;; ANSWER SECTION:
google.com.             300     IN      A       142.250.183.174

;; Query time: 437 msec
;; SERVER: 127.0.0.1#53(127.0.0.1)
;; WHEN: Tue Jan 28 14:41:06 IST 2025
;; MSG SIZE  rcvd: 38
```
