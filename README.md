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
## Working Flow Diagram 

<img width="2599" alt="shapes at 25-01-30 12 23 41" src="https://github.com/user-attachments/assets/688612bc-4af9-4c58-b802-99dfd13fa0c8" />


## Discussion 

The goal of domain names is to provide a mechanism for naming resources
in such a way that the names are usable in different hosts, networks,
protocol families, internets, and administrative organizations.

From the user's point of view, domain names are useful as arguments to a
local agent, called a resolver, which retrieves information associated
with the domain name.

Resolvers are responsible for dealing with the distribution of
the domain space and dealing with the effects of name server failure by
consulting redundant databases in other servers.

<p align="center">
<img width="931" alt="Screenshot 2025-01-30 at 11 54 53 AM" src="https://github.com/user-attachments/assets/7f4db5c1-e0e9-4659-8a58-967045e65728" />
</p>

```go

                 Local Host                        |  Foreign
                                                   |
    +---------+               +----------+         |  +--------+
    |         | user queries  |          |queries  |  |        |
    |  User   |-------------->|          |---------|->|Foreign |
    | Program |               | Resolver |         |  |  Name  |
    |         |<--------------|          |<--------|--| Server |
    |         | user responses|          |responses|  |        |
    +---------+               +----------+         |  +--------+
                                |     A            |
                cache additions |     | references |
                                V     |            |
                              +----------+         |
                              |  cache   |         |
                              +----------+         |
```

## Resource Records (RRs)

```py

                                    1  1  1  1  1  1
      0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                                               |
    /                                               /
    /                      NAME                     /
    |                                               |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                      TYPE                     |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                     CLASS                     |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                      TTL                      |
    |                                               |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                   RDLENGTH                    |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--|
    /                     RDATA                     /
    /                                               /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+

where:

NAME            an owner name, i.e., the name of the node to which this
                resource record pertains.

TYPE            two octets containing one of the RR TYPE codes.

CLASS           two octets containing one of the RR CLASS codes.

TTL             a 32 bit signed integer that specifies the time interval
                that the resource record may be cached before the source
                of the information should again be consulted.  Zero
                values are interpreted to mean that the RR can only be
                used for the transaction in progress, and should not be
                cached.  For example, SOA records are always distributed
                with a zero TTL to prohibit caching.  Zero values can
                also be used for extremely volatile data.

RDLENGTH        an unsigned 16 bit integer that specifies the length in
                octets of the RDATA field.

RDATA           a variable length string of octets that describes the
                resource.  The format of this information varies
                according to the TYPE and CLASS of the resource record.

```

## Format of Message 

```rs
All communications inside of the domain protocol are carried in a single
format called a message.  The top level format of message is divided
into 5 sections (some of which are empty in certain cases) shown below:

    +---------------------+
    |        Header       |
    +---------------------+
    |       Question      | the question for the name server
    +---------------------+
    |        Answer       | RRs answering the question
    +---------------------+
    |      Authority      | RRs pointing toward an authority
    +---------------------+
    |      Additional     | RRs holding additional information
    +---------------------+

The header section is always present.  The header includes fields that
specify which of the remaining sections are present, and also specify
whether the message is a query or a response, a standard query or some
other opcode, etc.

Header section format

The header contains the following fields:

                                    1  1  1  1  1  1
      0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                      ID                       |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |QR|   Opcode  |AA|TC|RD|RA|   Z    |   RCODE   |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                    QDCOUNT                    |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                    ANCOUNT                    |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                    NSCOUNT                    |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                    ARCOUNT                    |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+

```

## Question section format

```ts
                                 1  1  1  1  1  1
      0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                                               |
    /                     QNAME                     /
    /                                               /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                     QTYPE                     |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                     QCLASS                    |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+


/*
Example:
question := dnsmessage.Question{
Name:  dnsmessage.MustNewName("www.google.com."),
Type:  dnsmessage.TypeNS,
Class: dnsmessage.ClassINET,
}
*/


where:

QNAME           a domain name represented as a sequence of labels, where
                each label consists of a length octet followed by that
                number of octets.  The domain name terminates with the
                zero length octet for the null label of the root.  Note
                that this field may be an odd number of octets; no
                padding is used.

QTYPE           a two octet code which specifies the type of the query.
                The values for this field include all codes valid for a
                TYPE field, together with some more general codes which
                can match more than one type of RR.


QCLASS          a two octet code that specifies the class of the query.
                For example, the QCLASS field is IN for the Internet.
```

## Overview 
<p align="center">
<img src="https://i0.wp.com/robertleggett.blog/wp-content/uploads/2019/10/domain-name.png?resize=648%2C382&ssl=1" alt="dns-zone" />
</p>

<p align="center">
<img src="https://i0.wp.com/robertleggett.blog/wp-content/uploads/2019/10/domain-hierarchy-.png?resize=648%2C459&ssl=1" alt="dns-zone" />
</p>

## What are the different Domain Name Hierarchy levels?
### What is a Root Domain?
All domains begin with a root domain, referring to the domain name hierarchy diagram above it is represented as a full stop.

The root domain is handled by the Root Nameserver.

### What is a Top Level Domain (TLD)?
The top level domain is what comes after the root domain, referring to the domain name hierarchy diagram above this would be gov, blog, com, org, … The top level domain is handled by the TLD Nameserver.

### What is a Lower Level Domain?
The lower level domain is any other level after the top level domain, referring to the domain name hierarchy diagram above the second level domain would be robertleggett, google, amazon, … and the third level domain also knows as a subdomain would be the www.

The lower level domain, second to the nth, is handled by the TLD Nameserver, until the final lower level domain which is handled by the Authoritative Nameserver.



## DNS Zone 
<p align="center">
<img src="https://i0.wp.com/robertleggett.blog/wp-content/uploads/2019/11/dns-zones-5.png?resize=648%2C476&ssl=1" alt="dns-zone" />
</p>


## Reference
https://robertleggett.blog/2019/11/25/deep-dive-dns/ \
https://datatracker.ietf.org/doc/html/rfc1035#autoid-2 \
https://codingchallenges.fyi/challenges/challenge-dns-resolver/ \
https://jvns.ca/blog/2022/02/01/a-dns-resolver-in-80-lines-of-go/ \
https://ops.tips/blog/raw-dns-resolver-in-go/ \
https://brunoscheufler.com/blog/2024-05-12-building-a-dns-message-parser \
https://jameshfisher.com/2017/08/03/golang-dns-lookup/ \
https://github.com/CodingChallengesFYI/SharedSolutions/blob/main/Solutions/challenge-dns-resolver.md
